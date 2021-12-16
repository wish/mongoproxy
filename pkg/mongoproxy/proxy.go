package mongoproxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/models"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
	"github.com/wish/mongoproxy/pkg/mongowire"

	// Load all plugins
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/all"
)

var (
	clientConnectionCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "mongoproxy_client_accept_total",
		Help: "The total number of accepted client connections",
	})
	clientConnectionGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mongoproxy_client_connections_open",
		Help: "The current number of open client client connections",
	})
	clientMessageCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_client_message_total",
		Help: "The total number of messages from clients",
	}, []string{"opcode"})

	ErrServerClosed = errors.New("server closed")
	SKIP_RECOVER    = false
)

func init() {
	v, ok := os.LookupEnv("SKIP_RECOVER")
	if ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			fmt.Println("err", err)
		} else {
			SKIP_RECOVER = b
		}
	}
}

func NewProxy(l net.Listener, cfg *config.Config) (*Proxy, error) {
	// Create plugin chain
	ps, err := cfg.GetPlugins()
	if err != nil {
		return nil, err
	}

	p := &Proxy{
		l:           l,
		cfg:         cfg,
		doneChan:    make(chan struct{}),
		cursorCache: ttlcache.NewCache(),
	}

	// Create internal ClientConnection for "admin" tasks
	p.internalCC = plugins.NewClientConnection()
	// TODO: config
	p.internalCC.Addr, _ = net.ResolveIPAddr("ip", "127.0.0.1")

	if cfg.InternalIdentity != nil {
		p.internalCC.Identities = []plugins.ClientIdentity{cfg.InternalIdentity}
	}

	p.pipe = plugins.BuildPipeline(ps, p.baseRequestHandler)

	// Set up cursorCache
	p.cursorCache.SetTTL(p.cfg.IdleCursorTimeout) // default TTL -- config
	p.cursorCache.SetLoaderFunction(func(key string) (interface{}, time.Duration, error) {
		cursorID, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return nil, time.Duration(0), err
		}

		return plugins.NewCursorCacheEntry(cursorID), time.Duration(0), nil
	})
	// expiration handler to send killCursor commands
	p.cursorCache.SetExpirationReasonCallback(func(key string, reason ttlcache.EvictionReason, value interface{}) {
		logrus.Tracef("expire cursor %s", key)
		i, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return
		}

		// If the cursor expired (we timed out waiting) we want to kill the downstream cursor as we remove it from the cache
		if reason == ttlcache.Expired {
			p.HandleMongo(context.TODO(), &plugins.Request{CursorCache: p, CC: p.internalCC}, bson.D{
				{"killCursors", "admin"},
				{"cursors", primitive.A{i}},
			})
		}
	})

	return p, nil
}

type Proxy struct {
	l   net.Listener // Listener for incoming client connections
	cfg *config.Config

	pipe plugins.PipelineFunc

	doneChan chan struct{}

	activeConn     map[*conn]struct{}
	activeConnLock sync.Mutex

	// Cursor cache for plugins.CursorCacheEntry this is stored at the proxy
	// level because both the core proxy needs it (to handle OP_QUERY and OP_GETMORE)
	// as well as the plugins (e.g. mongo)
	cursorCache *ttlcache.Cache

	internalCC *plugins.ClientConnection
}

func (p *Proxy) GetCursor(cursorID int64) *plugins.CursorCacheEntry {
	v, err := p.cursorCache.Get(strconv.FormatInt(cursorID, 10))
	if err == ttlcache.ErrNotFound {
		panic("what")
	}

	return v.(*plugins.CursorCacheEntry)
}

func (p *Proxy) CloseCursor(cursorID int64) {
	p.cursorCache.Remove(strconv.FormatInt(cursorID, 10))
}

func (p *Proxy) Addr() string {
	return p.l.Addr().String()
}

func (p *Proxy) baseRequestHandler(ctx context.Context, r *plugins.Request) (bson.D, error) {
	switch cmd := r.Command.(type) {
	case *command.ConnectionStatus:
		var (
			authenticatedUsers     = make([]models.AuthenticatedUser, 0)
			authenticatedUserRoles = make([]models.AuthenticatedUserRole, 0)
		)

		// maps to dedupe
		users := make(map[string]struct{})
		roles := make(map[string]struct{})

		for _, identity := range r.CC.Identities {
			userKey := identity.User() + "." + "admin"
			if _, ok := users[userKey]; !ok {
				authenticatedUsers = append(authenticatedUsers, models.AuthenticatedUser{User: identity.User(), DB: "admin"}) // TODO: different DBs?
				users[userKey] = struct{}{}
			}
			for _, role := range identity.Roles() {
				roleKey := role + "." + "admin"
				if _, ok := roles[roleKey]; !ok {
					authenticatedUserRoles = append(authenticatedUserRoles, models.AuthenticatedUserRole{
						Role: role,
						DB:   "admin", // TODO: different DBs?
					})
					roles[roleKey] = struct{}{}
				}
			}
		}

		return bson.D{
			primitive.E{"ok", 1},
			{"authenticatedUsers", authenticatedUsers},
			{"authenticatedUserRoles", authenticatedUserRoles},
		}, nil

	case *command.HostInfo:
		return bson.D{
			{"system", bson.D{
				//{"currentTime", }
				{"hostname", "EXAMPLEHOSTNAME"},
				{"cpuAddrSize", 64}, // TODO: discover
				//"memSizeMB" : <number>,
				//"memLimitMB" : <number>,  // Available starting in MongoDB 4.0.9 (and 3.6.13)
				//"numCores" : <number>,
				//"cpuArch" : "<identifier>",
				//"numaEnabled" : <boolean>
			}},
			{"os", bson.D{
				//"type" : "<string>",
				//"name" : "<string>",
				//"version" : "<string>"
			}},
			{"extra", bson.D{
				//"versionString" : "<string>",
				//"libcVersion" : "<string>",
				//"kernelVersion" : "<string>",
				//"cpuFrequencyMHz" : "<string>",
				//"cpuFeatures" : "<string>",
				//"pageSize" : <number>,
				//"numPages" : <number>,
				//"maxOpenFiles" : <number>
			}},
		}, nil

	case *command.Logout:
		r.CC.Identities = nil
		return bson.D{
			primitive.E{"ok", 1},
		}, nil

	case *command.BuildInfo:
		return bson.D{
			{"bits", 64}, //TODO dynamically pull
			{"debug", false},
			{"version", "4.3.1"}, // TODO pull from downstream? Or have set in config
			{"maxBsonObjectSize", bsonutil.MaxBsonObjectSize},
			{"ok", 1},
		}, nil

	case *command.Ping:
		return bson.D{
			{"ok", 1},
		}, nil

	// TODO: complete more options
	case *command.ServerStatus:
		return bson.D{
			//{"host" : <string>,
			//{"advisoryHostFQDNs" : <array>,
			{"version", "4.3.1"}, // TODO pull from downstream? Or have set in config
			//{"process" : <"mongod"|"mongos">,
			//{"pid" : <num>,
			//{"uptime" : <num>,
			//{"uptimeMillis" : <num>,
			//{"uptimeEstimate" : <num>,
			//{"localTime" : ISODate(""),

			{"ok", 1},
		}, nil

	// Pretend we are mongoS
	case *command.IsDBGrid:
		return bson.D{
			{"isdbgrid", 1},
			{"hostname", "EXAMPLEHOSTNAME"},
			{"ok", 1},
		}, nil

	case *command.IsMaster:
		ret := bson.D{
			{"ismaster", true},
			{"localTime", time.Now().Truncate(time.Millisecond)},
			{"logicalSessionTimeoutMinutes", 30},
			{"maxBsonObjectSize", bsonutil.MaxBsonObjectSize},
			{"maxMessageSizeBytes", 48000000},
			{"maxWireVersion", 8},
			{"maxWriteBatchSize", 100000},
			{"minWireVersion", 0},
			{"msg", "isdbgrid"},
			{"ok", 1},
		}

		// TODO: validate compressors
		if len(p.cfg.Compressors) > 0 && len(cmd.Compression) > 0 {
			var compressors primitive.A
			for _, clientC := range cmd.Compression {
				for _, serverC := range p.cfg.Compressors {
					if clientC == serverC {
						compressors = append(compressors, clientC)
						break
					}
				}
			}
			ret = append(ret, bson.E{"compression", compressors})
		}
		return ret, nil
	}
	return nil, fmt.Errorf("unhandled command %s: %v", r.CommandName, r.Command)
}

func (p *Proxy) Serve() error {
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		c, err := p.l.Accept()
		if err != nil {
			select {
			case <-p.doneChan:
				return ErrServerClosed
			default:
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				logrus.Infof("Accept error: %v; retrying in %v", err, tempDelay)

				time.Sleep(tempDelay)
				continue

			}
			return err
		}
		clientConnectionCounter.Inc()
		clientConnectionGauge.Inc()

		go func(c net.Conn) {
			defer func() {
				clientConnectionGauge.Dec()
				if !SKIP_RECOVER {
					if err := recover(); err != nil {
						logrus.Errorf("Panic in connection: %v", err)
						sentry.CurrentHub().Recover(err)
						sentry.Flush(time.Second * 5)
					}
				}
			}()
			logrus.Debugf("Starting connection: %v", c)
			if err := p.clientServeLoop(c); err != nil && err != io.EOF {
				logrus.Errorf("Error serving client: %s %v -- %s", reflect.TypeOf(err), err, err.Error())
			}
		}(c)
	}
}

func (p *Proxy) trackConn(c *conn, add bool) {
	p.activeConnLock.Lock()
	defer p.activeConnLock.Unlock()

	if p.activeConn == nil {
		p.activeConn = make(map[*conn]struct{})
	}

	if add {
		p.activeConn[c] = struct{}{}
	} else {
		delete(p.activeConn, c)
	}
}

func (p *Proxy) closeIdleConns() bool {
	p.activeConnLock.Lock()
	defer p.activeConnLock.Unlock()

	quiescent := true

	for c := range p.activeConn {
		st, unixSec := c.getState()
		// TODO: timeout for this idle
		// If the connection isn't idle or has activity in the last 5s its not idle
		if st != StateIdle || unixSec >= time.Now().Unix()-5 {
			quiescent = false
			continue
		}

		c.c.Close()
		delete(p.activeConn, c)
	}

	return quiescent
}

func (p *Proxy) closeDoneChan() {
	select {
	case <-p.doneChan:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		close(p.doneChan)
	}
}

func (p *Proxy) Shutdown(ctx context.Context) error {
	p.closeDoneChan()

	// Close the listener
	lnerr := p.l.Close()

	ticker := time.NewTicker(time.Millisecond * 200) // TODO: config?
	defer ticker.Stop()
	for {
		if p.closeIdleConns() {
			return lnerr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (p *Proxy) handleOp(ctx context.Context, clientConn *plugins.ClientConnection, req *mongowire.Request) (mongowire.WireSerializer, error) {
	logrus.Debugf("header received: %v", req.GetHeader())

	clientMessageCounter.WithLabelValues(req.GetHeader().OpCode.String()).Inc()

	switch req.GetHeader().OpCode {
	case mongowire.OpQuery:
		q := req.GetOpQuery()
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("IN OP_QUERY %s", mongowire.ToJson(q, p.cfg.RequestLengthLimit))
		}

		reply, err := p.handleOpQuery(ctx, clientConn, q)
		if err != nil {
			return nil, err
		}

		reply.NumberReturned = int32(len(reply.Documents))
		reply.Header.OpCode = mongowire.OpReply
		reply.Header.ResponseTo = reply.Header.RequestID

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("OUT OP_QUERY %s", mongowire.ToJson(reply, p.cfg.RequestLengthLimit))
		}
		return reply, nil

	case mongowire.OpKillCursors:
		q := req.GetOpKillCursors()
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("IN OP_KILL_CURSORS %s", mongowire.ToJson(q, p.cfg.RequestLengthLimit))
		}
		p.handleOpKillCursors(ctx, clientConn, q)

	case mongowire.OpGetMore:
		q := req.GetOpMore()
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("IN OP_GETMORE %s", mongowire.ToJson(q, p.cfg.RequestLengthLimit))
		}

		reply, err := p.handleOpGetMore(ctx, clientConn, q)
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("OUT OP_GETMORE %s", mongowire.ToJson(reply, p.cfg.RequestLengthLimit))
		}
		return reply, err

	case mongowire.OpMsg:
		m := req.GetOpMsg()

		// If the OP_MSG has set moreToCome we aren't allowed to respond
		// https://docs.mongodb.com/manual/reference/mongodb-wire-protocol/#flag-bits
		if m.Flags.MoreToCome() {
			// TODO: do we want to background this? Or just run it without a return. As
			// it stands how this causes race conditions with writeConcern=0 writes and subsequent writes
			// from the same client on the same connection (e.g. test.test_common.TestCommon.test_mongo_client)
			go p.handleOpMsg(ctx, clientConn, m)
			return nil, nil
		}

		reply, err := p.handleOpMsg(ctx, clientConn, m)
		if err != nil {
			return nil, err
		}

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("OUT OP_MSG %s", mongowire.ToJson(reply, p.cfg.RequestLengthLimit))
		}
		return reply, nil

	case mongowire.OpCompressed:
		m := req.GetOpCompressed()
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("IN OP_COMPRESSED %s", mongowire.ToJson(req, p.cfg.RequestLengthLimit))
		}

		// Decompress
		b, err := driver.DecompressPayload(m.CompressedMessage, driver.CompressionOpts{
			Compressor:       m.CompressorID,
			UncompressedSize: m.UncompressedSize,
		})
		if err != nil {
			panic(err) // TODO
		}

		newReq := mongowire.NewRequestWithHeader(*req.GetHeader(), bytes.NewReader(b))
		newReq.GetHeader().OpCode = m.OriginalOpcode
		newReq.GetHeader().MessageLength = m.UncompressedSize + mongowire.HeaderLen
		reply, err := p.handleOp(ctx, clientConn, newReq)
		if err != nil {
			return nil, err
		}

		// Generate output of inner message
		buf := bytes.NewBuffer(nil)
		if err := reply.WriteTo(buf); err != nil {
			return nil, err
		}
		// Wrap
		compressedReply := &mongowire.OP_COMPRESSED{
			Header:           *req.GetHeader(),
			CompressorID:     m.CompressorID,
			OriginalOpcode:   reply.GetHeader().OpCode,
			UncompressedSize: int32(len(buf.Bytes()[mongowire.HeaderLen:])),
		}

		// Compress message with the same thing that it came in with
		compressedB, err := driver.CompressPayload(buf.Bytes()[mongowire.HeaderLen:], driver.CompressionOpts{
			Compressor: compressedReply.CompressorID,
			// TODO: options
			ZlibLevel: wiremessage.DefaultZlibLevel,
			ZstdLevel: wiremessage.DefaultZstdLevel,
		})
		if err != nil {
			panic(err) // TODO
		}
		compressedReply.CompressedMessage = compressedB

		// return
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("OUT OP_COMPRESSED %s", mongowire.ToJson(compressedReply, p.cfg.RequestLengthLimit))
		}
		return compressedReply, nil

	default:
		logrus.Debugf("Unhandled opcode: %v", req.GetHeader().OpCode)
		return nil, fmt.Errorf("unhandled opcode: %v", req.GetHeader().OpCode)
	}

	return nil, nil
}

func (p *Proxy) clientServeLoop(c net.Conn) error {
	conn := &conn{
		p: p,
		c: c,
	}
	conn.setState(StateNew)

	clientConn := plugins.NewClientConnection()
	clientConn.Addr = c.RemoteAddr()
	defer func() {
		c.Close()
		clientConn.Close()
		conn.setState(StateClosed)

		logrus.Debugf("Closing connection: %v", c)
	}()

	for {
		conn.setState(StateIdle)
		logrus.Debugf("waiting for request %v", c)
		req, err := mongowire.NewRequest(c)
		if err != nil {
			return err
		}
		conn.setState(StateActive)

		// TODO: context that will close when the client connection closes
		ctx := context.Background()

		// Unpack request

		// Handle Reply (write to wire)

		reply, err := p.handleOp(ctx, clientConn, req)
		if err != nil {
			return err
		}

		// If we have a reply, write it back out
		if reply != nil {
			if err := reply.WriteTo(c); err != nil {
				return err
			}
		}
	}
}
