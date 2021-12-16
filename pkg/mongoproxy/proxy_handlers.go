package mongoproxy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoerror"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins/mongo"
	"github.com/wish/mongoproxy/pkg/mongowire"
)

var (
	clientCommandCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_client_command_total",
		Help: "The total number of commands from clients",
	}, []string{"command"})
)

// HandleMongo needs to actually disbatch the command. This includes loading the command into a struct, processing the pipeline, and then returning
func (p *Proxy) HandleMongo(ctx context.Context, req *plugins.Request, d bson.D) (bson.D, error) {
	if len(d) == 0 {
		return nil, errors.New("invalid bson Doc")
	}

	cmd, ok := command.GetCommand(d[0].Key)
	if !ok {
		return mongoerror.CommandNotFound.ErrMessage("no such command: '" + d[0].Key + "'"), nil
	}
	clientCommandCounter.WithLabelValues(d[0].Key).Inc()

	if err := cmd.FromBSOND(d); err != nil {
		d, err := mongo.ErrorToDoc(err)
		if err != nil {
			return mongoerror.FailedToParse.ErrMessage(err.Error()), nil
		}
		return append(bson.D{{"ok", 0}}, d...), nil
	}

	req.CommandName = d[0].Key
	req.Command = cmd

	// handle error -- check if its a type we can convert; if so convert (so we don't close the connection)
	resp, err := p.pipe(ctx, req)
	if err != nil {
		// TODO: move this logic down; here we only want to check against some BSONError interface type; so other plugins can implement their own errors that become the same on the wire
		d, err := mongo.ErrorToDoc(err)
		if err != nil {
			return nil, err
		}
		return append(bson.D{{"ok", 0}}, d...), nil
	}

	return resp, nil
}

// handleOpQuery handles parsing out the OP_QUERY and converting it into commands to run through the plugin framework
// Unfortunately this method is a bit long because of the logic required to do the conversion; but it is all the conversion
// logic is consolidated here in an attempt to make this easier to understand.
func (p *Proxy) handleOpQuery(ctx context.Context, cc *plugins.ClientConnection, q *mongowire.OP_QUERY) (*mongowire.OP_REPLY, error) {
	request := &plugins.Request{
		CC:          cc,
		CursorCache: p,
	}
	defer request.Close()

	reply := &mongowire.OP_REPLY{
		Header: q.Header,
	}

	names := strings.Split(q.FullCollectionName, ".")

	// For weird names (e.g. old currentOp format) just return a CommandNot found for now; if we need to
	// add support in the future we can; this seems to just force the client to the newer format.
	if len(names) > 2 {
		reply.Documents = append(reply.Documents, mongoerror.CommandNotFound.ErrMessage("no such command"))
		return reply, nil
	}

	switch names[1] {
	// $cmd is all "command" methods
	case "$cmd":
		var downstreamQuery bson.D
		if q.Query[0].Key[0] != '$' {
			downstreamQuery = q.Query
		} else {
			downstreamQuery = append(q.Query[0].Value.(primitive.D), q.Query[1:]...)
		}

		downstreamQuery = append(downstreamQuery, primitive.E{Key: "$db", Value: names[0]})

		if q.NumberToSkip > 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "skip", Value: q.NumberToSkip})
		}
		// For a command we specifically do NOT use the NumberToReturn as this is a limit on the number of documents
		// we are allowed to return in OP_REPLY; and all commands in $cmd will do that by definition
		if len(q.ReturnFieldsSelector) > 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "projection", Value: q.ReturnFieldsSelector})
		}

		// Convert flags from request
		if q.Flags.SlaveOk() {
			switch downstreamQuery[0].Key {
			case "count", "distinct":
				downstreamQuery = append(downstreamQuery, primitive.E{Key: "readPreference", Value: bson.D{{"mode", "secondaryPreferred"}}})
			}
		}

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Query Converted: %v", mongowire.ToJson(downstreamQuery, p.cfg.RequestLengthLimit))
		}

		// run the converted query through the handlers
		result, err := p.HandleMongo(ctx, request, downstreamQuery)
		if err != nil {
			return nil, err
		}

		if !bsonutil.Ok(result) {
			reply.Documents = []bson.D{result}
			return reply, nil
		}

		if downstreamQuery[0].Key == "aggregate" {
			var ok bool
			_, ok = bsonutil.Lookup(downstreamQuery, "cursor")
			if !ok {
				_, ok = bsonutil.Lookup(downstreamQuery, "useCursor")
			}
			// For this older wire protocol a cursor wasn't required. So if a cursor
			// wasn't requested we need to unwrap the results.
			if !ok {
				v, ok := bsonutil.Lookup(result, "cursor", "firstBatch")
				if !ok {
					return nil, fmt.Errorf("missing firstBatch in cursor response")
				}

				if cursorID, ok := bsonutil.Lookup(result, "cursor", "id"); ok {
					p.GetCursor(cursorID.(int64)).CursorConsumed += len(v.(primitive.A))
				}

				reply.Documents = []bson.D{{{"ok", 1}, {"result", v}}}
				return reply, nil
			}
		}

		// If there is a cursor to return; we want to capture that cursor and set consumed etc.
		// We also need to make sure to *NOT* set a cursorID in the response header as that is
		// only allowed for find queries
		if cursorDataRaw, ok := bsonutil.Lookup(result, "cursor"); ok {
			cursorData, ok := cursorDataRaw.(bson.D)
			if !ok {
				return nil, fmt.Errorf("wrong type for cursor")
			}
			var cursorEntry *plugins.CursorCacheEntry
			if cursorID, ok := bsonutil.Lookup(cursorData, "id"); ok {
				cursorEntry = p.GetCursor(cursorID.(int64))
			}
			firstBatchRaw, ok := bsonutil.Lookup(cursorData, "firstBatch")
			if ok {
				cursorEntry.CursorConsumed += len(firstBatchRaw.(primitive.A))
			}
		}
		reply.Documents = []bson.D{result}
	// All other methods are "find" queries
	default:
		downstreamQuery := []primitive.E{
			{Key: "find", Value: names[1]},
			{Key: "$db", Value: names[0]},
		}

		// If there are multiple items in the query; the first item is "query" if we have other flags (e.g. orderby)
		// otherwise the entire Query is the filter
		if len(q.Query) > 0 && q.Query[0].Key == "$query" {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "filter", Value: q.Query[0].Value})
			for _, qV := range q.Query[1:] {
				switch qV.Key {
				case "orderby", "$orderby":
					downstreamQuery = append(downstreamQuery, primitive.E{Key: "sort", Value: qV.Value})
				// Cases we strip the first character
				case "$maxTimeMS", "$hint", "$snapshot", "$comment", "$collation":
					downstreamQuery = append(downstreamQuery, primitive.E{Key: qV.Key[1:], Value: qV.Value})
				// Cases we pass through unmapped
				case "hint", "snapshot", "$readPreference", "comment", "collation":
					downstreamQuery = append(downstreamQuery, qV)
				default:
					fmt.Println("TODO", qV)
					panic("what")
				}

			}
		} else {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "filter", Value: q.Query})
		}

		if q.NumberToSkip > 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "skip", Value: q.NumberToSkip})
		}
		if q.NumberToReturn > 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "batchSize", Value: q.NumberToReturn})
		} else if q.NumberToReturn < 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "batchSize", Value: -q.NumberToReturn})
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "singleBatch", Value: true})
		}
		if len(q.ReturnFieldsSelector) > 0 {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "projection", Value: q.ReturnFieldsSelector})
		}

		var returnDirect bool

		// if its an explain; convert! in this older wire protocol it *seems* that explains are only supported in filters
		if newDownstreamQuery, v, ok := bsonutil.Pop(downstreamQuery, "filter", "$explain"); ok {
			if v, ok := v.(bool); ok && v {
				returnDirect = true
				downstreamQuery = []primitive.E{
					{Key: "explain", Value: newDownstreamQuery},
					{Key: "$db", Value: names[0]},
				}
			}
		}

		// Convert flags from request
		if q.Flags.TailableCursor() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "tailable", Value: true})
		}
		if q.Flags.SlaveOk() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "readPreference", Value: bson.D{{"mode", "secondaryPreferred"}}})
		}
		if q.Flags.OplogReplay() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "oplogReplay", Value: true})
		}
		if q.Flags.NoCursorTimeout() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "noCursorTimeout", Value: true})
		}
		if q.Flags.AwaitData() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "awaitData", Value: true})
		}
		if q.Flags.Partial() {
			downstreamQuery = append(downstreamQuery, primitive.E{Key: "allowPartialResults", Value: true})
		}

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Query Converted: %v", mongowire.ToJson(downstreamQuery, p.cfg.RequestLengthLimit))
		}

		result, err := p.HandleMongo(ctx, request, downstreamQuery)
		if err != nil {
			return nil, err
		}

		if !bsonutil.Ok(result) {
			reply.Documents = []bson.D{result}
			return reply, nil
		}

		if returnDirect {
			reply.Documents = []bson.D{result}
			return reply, nil
		}

		// This is supposed to send the docs as if we did a bunch of GET_MORE over and over for the find. Since this isn't used today
		// we'll delay implementation
		if q.Flags.Exhaust() {
			return nil, fmt.Errorf("not implemented")
		}

		// Get documents
		documents := make([]bson.D, 0, 100) // TODO: sizing?
		if cursorDataRaw, ok := bsonutil.Lookup(result, "cursor"); ok {
			cursorData, ok := cursorDataRaw.(bson.D)
			if !ok {
				return nil, fmt.Errorf("wrong type for cursor")
			}
			var cursorEntry *plugins.CursorCacheEntry
			if cursorID, ok := bsonutil.Lookup(cursorData, "id"); ok {
				cursorEntry = p.GetCursor(cursorID.(int64))
				reply.CursorID = cursorID.(int64)
			}
			firstBatchRaw, ok := bsonutil.Lookup(cursorData, "firstBatch")
			if ok {
				for _, doc := range firstBatchRaw.(primitive.A) {
					documents = append(documents, doc.(bson.D))
				}
				cursorEntry.CursorConsumed += len(documents)
			}
		}

		reply.Documents = documents
	}

	return reply, nil
}

// Responsible to kill the requested cursors
func (p *Proxy) handleOpKillCursors(ctx context.Context, cc *plugins.ClientConnection, q *mongowire.OP_KILL_CURSORS) error {
	request := &plugins.Request{
		CC:          cc,
		CursorCache: p,
	}
	defer request.Close()

	result, err := p.HandleMongo(ctx, request, []primitive.E{
		{Key: "killCursors", Value: "admin"}, // TODO: fix? For now we don't know the name (since this is the old wire protocol)
		{Key: "cursors", Value: q.CursorIDs},
	})

	// Barring major er
	if err == nil {
		if bsonutil.Ok(result) {
			for _, cursorID := range q.CursorIDs {
				p.CloseCursor(cursorID)
			}
		}
	}

	return err
}

func (p *Proxy) handleOpGetMore(ctx context.Context, cc *plugins.ClientConnection, q *mongowire.OP_GETMORE) (*mongowire.OP_REPLY, error) {
	request := &plugins.Request{
		CC:          cc,
		CursorCache: p,
	}
	defer request.Close()

	names := strings.Split(q.FullCollectionName, ".")

	reply := &mongowire.OP_REPLY{
		Header: q.Header,
	}

	cursorEntry := p.GetCursor(q.CursorID)

	result, err := p.HandleMongo(ctx, request, []primitive.E{
		{Key: "getMore", Value: q.CursorID},
		{Key: "batchSize", Value: q.NumberToReturn},
		{Key: "$db", Value: names[0]},
		{Key: "collection", Value: names[1]},
	})
	if err != nil {
		return nil, err
	}

	// Get documents
	if cursorDataRaw, ok := bsonutil.Lookup(result, "cursor"); ok {
		cursorData, ok := cursorDataRaw.(bson.D)
		if !ok {
			return nil, fmt.Errorf("wrong type for cursor")
		}
		if cursorID, ok := bsonutil.Lookup(cursorData, "id"); ok {
			reply.CursorID = cursorID.(int64)
			if cursorID.(int64) == 0 {
				p.CloseCursor(cursorEntry.ID)
			}
		}
		nextBatchRaw, ok := bsonutil.Lookup(cursorData, "nextBatch")
		if ok {
			for _, doc := range nextBatchRaw.(primitive.A) {
				reply.Documents = append(reply.Documents, doc.(bson.D))
			}
		}
	}

	reply.StartingFrom = int32(cursorEntry.CursorConsumed)
	cursorEntry.CursorConsumed += len(reply.Documents)
	reply.NumberReturned = int32(len(reply.Documents))
	reply.Header.OpCode = mongowire.OpReply
	reply.Header.ResponseTo = reply.Header.RequestID

	return reply, nil
}

func (p *Proxy) handleOpMsg(ctx context.Context, cc *plugins.ClientConnection, m *mongowire.OP_MSG) (*mongowire.OP_MSG, error) {
	request := &plugins.Request{
		CC:          cc,
		CursorCache: p,
	}
	defer request.Close()

	reply := &mongowire.OP_MSG{
		Header:   m.Header,
		Sections: []mongowire.MSGSection{},
	}
	reply.Header.ResponseTo = m.Header.RequestID

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("IN OP_MSG %d %s", len(m.Sections), mongowire.ToJson(m, p.cfg.RequestLengthLimit))
	}

	var d bson.D

	for _, sectionRaw := range m.Sections {
		switch sectionTyped := sectionRaw.(type) {
		case mongowire.MSGSection_Body:
			// If not started add to doc
			d = append(sectionTyped.Document, d...)

		case mongowire.MSGSection_DocumentSequence:
			// TODO: handle nested fields in SequenceIdentifier (if there is a "." in there we need to do something different)
			if strings.Contains(sectionTyped.SequenceIdentifier, ".") {
				return nil, fmt.Errorf("not implemented")
			}
			d = append(d, primitive.E{sectionTyped.SequenceIdentifier, sectionTyped.Documents})
		default:
			return nil, fmt.Errorf("not implemented")
		}
	}

	// run command
	result, err := p.HandleMongo(ctx, request, d)
	if err != nil {
		return nil, err
	}
	// TODO: something smarter about the size of that result; if too big we can do a document sequence
	reply.Sections = append(reply.Sections, mongowire.MSGSection_Body{result})

	return reply, nil
}
