package authz

import (
	"context"
	"log"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/fsnotify.v1"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoerror"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins/authz/authzlib"
)

const UNAUTHENTICATED_ROLE = "UNAUTHENTICATED"

var (
	configUpdates = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_authz_updates_total",
		Help: "The total config updates completed",
	}, []string{"success"})
	authzDeny = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_authz_deny_total",
		Help: "The total deny returns of a command",
	}, []string{"db", "collection", "command"})

	OPEN_COMMAND = map[string]struct{}{
		"isMaster":         {},
		"ismaster":         {},
		"buildInfo":        {},
		"buildinfo":        {},
		"connectionStatus": {},
		"saslStart":        {},
		"getnonce":         {},
		"logout":           {},
		"ping":             {},
	}
)

type contextKey string

func (c contextKey) String() string {
	return "authz context key " + string(c)
}

var (
	contextKeyResources = contextKey("authz.resources")
)

const Name = "authz"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &AuthzPlugin{}
	})
}

type AuthzPluginConfig struct {
	// Paths is the path on disk to load authz policies/roles/config from
	Paths              []string `bson:"paths"`
	LogUnauthenticated bool     `bson:"logUnauthenticated"` // Log all unauthenticated requests

	// DenyByDefault controls whether the default policy is to deny (true) or not (false)
	DenyByDefault           bool            `bson:"denyByDefault"`
	DenyByDefaultNamespaces map[string]bool `bson:"denyByDefaultNamespaces"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type AuthzPlugin struct {
	conf AuthzPluginConfig
	a    authzlib.Authz
}

func (p *AuthzPlugin) Name() string { return Name }

func (p *AuthzPlugin) LoadConfig() (err error) {
	defer func() {
		if err != nil {
			configUpdates.WithLabelValues("false").Add(1)
		} else {
			configUpdates.WithLabelValues("true").Add(1)
		}
	}()

	return p.a.LoadConfig(context.TODO(), p.conf.Paths, nil)
}

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *AuthzPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	// LoadConfig
	if err := p.LoadConfig(); err != nil {
		return err
	}

	// start watch
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Debugf("Schema watcher event: %v", event)
				p.LoadConfig()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("Schema watcher: %v", err)
			}
		}
	}()

	for _, pth := range p.conf.Paths {
		if err := watcher.Add(path.Dir(pth)); err != nil {
			return err
		}
	}

	return nil
}

func (p *AuthzPlugin) resourcesForCommand(r *plugins.Request, c command.Command) map[authzlib.AuthorizationMethod][]authzlib.Resource {
	resourceMap := make(map[authzlib.AuthorizationMethod][]authzlib.Resource)

	// Pick which commands we allow without authentication at all
	switch cmd := c.(type) {
	case *command.Aggregate:
		if collection := cmd.GetCollection(); collection != "" {
			resourceMap[authzlib.Read] = []authzlib.Resource{
				{
					DB:         cmd.GetDatabase(),
					Collection: collection,
				},
			}
		} else { // Otherwise it is a DB wide action
			resourceMap[authzlib.Read] = []authzlib.Resource{
				{
					DB: cmd.GetDatabase(),
				},
			}
		}
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.CollStats:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.Count:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.Create:
		resourceMap[authzlib.Create] = []authzlib.Resource{
			{
				DB: cmd.GetDatabase(),
			},
		}

	case *command.CreateIndexes:
		resourceMap[authzlib.Create] = []authzlib.Resource{
			{
				DB: cmd.GetDatabase(),
			},
		}

	case *command.CurrentOp:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.Delete:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.DeleteIndexes:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.Distinct:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
				Field:      cmd.Key,
			},
		}

	case *command.DropDatabase:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.Drop:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				DB: cmd.GetDatabase(),
			},
		}

	case *command.DropIndexes:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.EndSessions:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.Explain:
		resourceMap = p.resourcesForCommand(r, cmd.Cmd)
		switch cmd.Verbosity {
		case "queryPlanner":
			resourceMap[authzlib.Read] = append(resourceMap[authzlib.Read], authzlib.Resource{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			})
		default:
			resourceMap[authzlib.Read] = append(resourceMap[authzlib.Read], authzlib.Resource{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			})
			resourceMap[authzlib.Read] = append(resourceMap[authzlib.Read], authzlib.Resource{
				Global: true,
			})
		}

	case *command.FindAndModify:
		// Read permissions
		switch len(cmd.Fields) {
		case 0:
			resourceMap[authzlib.Read] = []authzlib.Resource{
				{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      "*",
				},
			}
		default:
			resourceMap[authzlib.Read] = make([]authzlib.Resource, 0, len(cmd.Fields))
			for _, item := range cmd.Fields {
				if !bsonutil.BoolNumber(item.Value) {
					continue
				}
				resourceMap[authzlib.Read] = append(resourceMap[authzlib.Read], authzlib.Resource{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      item.Key,
				})
			}
		}
		// Update permissions
		f := bsonutil.ExpandUpdate(cmd.Update, cmd.Upsert)
		if len(f.Create) > 0 {
			resourceMap[authzlib.Create] = make([]authzlib.Resource, len(f.Create))
			for i, c := range f.Create {
				resourceMap[authzlib.Create][i] = authzlib.Resource{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      c,
				}
			}
		}
		if len(f.Update) > 0 {
			resourceMap[authzlib.Update] = make([]authzlib.Resource, len(f.Update))
			for i, c := range f.Update {
				resourceMap[authzlib.Update][i] = authzlib.Resource{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      c,
				}
			}
		}
		if len(f.Delete) > 0 {
			resourceMap[authzlib.Delete] = make([]authzlib.Resource, len(f.Delete))
			for i, c := range f.Delete {
				resourceMap[authzlib.Delete][i] = authzlib.Resource{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      c,
				}
			}
		}

	// TODO: should we require authz for filter fields?
	// TODO: should we require authz for sort fields?
	case *command.Find:
		switch len(cmd.Projection) {
		case 0:
			resourceMap[authzlib.Read] = []authzlib.Resource{
				{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      "*",
				},
			}
		default:
			resourceMap[authzlib.Read] = make([]authzlib.Resource, 0, len(cmd.Projection))
			for _, item := range cmd.Projection {
				if !bsonutil.BoolNumber(item.Value) {
					continue
				}
				resourceMap[authzlib.Read] = append(resourceMap[authzlib.Read], authzlib.Resource{
					DB:         cmd.GetDatabase(),
					Collection: cmd.GetCollection(),
					Field:      item.Key,
				})
			}
		}

	case *command.GetMore:
		cursorResources := r.CursorCache.GetCursor(cmd.CursorID).Map[contextKeyResources]
		if cr, ok := cursorResources.(map[authzlib.AuthorizationMethod][]authzlib.Resource); ok {
			return cr
		}
		return nil

	case *command.HostInfo:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.Insert:
		resourceMap[authzlib.Create] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.KillAllSessions:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.KillCursors:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.KillOp:
		resourceMap[authzlib.Delete] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.ListCollections:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB: cmd.GetDatabase(),
			},
		}

	case *command.ListDatabases:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.ListIndexes:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				DB:         cmd.GetDatabase(),
				Collection: cmd.GetCollection(),
			},
		}

	case *command.ServerStatus:
		resourceMap[authzlib.Read] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.ShardCollection:
		resourceMap[authzlib.Update] = []authzlib.Resource{
			{
				Global: true,
			},
		}

	case *command.Update:
		for _, update := range cmd.Updates {
			f := bsonutil.ExpandUpdate(update.U, update.Upsert)
			if len(f.Create) > 0 {
				resourceMap[authzlib.Create] = make([]authzlib.Resource, len(f.Create))
				for i, c := range f.Create {
					resourceMap[authzlib.Create][i] = authzlib.Resource{
						DB:         cmd.GetDatabase(),
						Collection: cmd.GetCollection(),
						Field:      c,
					}
				}
			}
			if len(f.Update) > 0 {
				resourceMap[authzlib.Update] = make([]authzlib.Resource, len(f.Update))
				for i, c := range f.Update {
					resourceMap[authzlib.Update][i] = authzlib.Resource{
						DB:         cmd.GetDatabase(),
						Collection: cmd.GetCollection(),
						Field:      c,
					}
				}
			}
			if len(f.Delete) > 0 {
				resourceMap[authzlib.Delete] = make([]authzlib.Resource, len(f.Delete))
				for i, c := range f.Delete {
					resourceMap[authzlib.Delete][i] = authzlib.Resource{
						DB:         cmd.GetDatabase(),
						Collection: cmd.GetCollection(),
						Field:      c,
					}
				}
			}
		}
	}

	return resourceMap
}

// Process is the function executed when a message is called in the pipeline.
func (p *AuthzPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	// If the command is in the list of unauthenticated commands; move on
	if _, ok := OPEN_COMMAND[r.CommandName]; ok {
		return next(ctx, r)
	}

	resourceMap := p.resourcesForCommand(r, r.Command)

	// If there is no resource; we don't allow the call through
	if len(resourceMap) == 0 {
		return mongoerror.Unauthorized.ErrMessage("unauthorized no resource for " + r.CommandName), nil
	}

	// Now we do the batch of authorization calls
	identities := r.CC.Identities
	roles := make([]string, 0, 100)
	rolesM := make(map[string]struct{})
	// If there is no identity on the connection then we set a "canned" UNAUTHENTICATED_ROLE
	// so that the policies can handle unauthenticated users directly
	if identities == nil {
		if p.conf.LogUnauthenticated {
			logrus.NewEntry(logrus.StandardLogger()).WithFields(logrus.Fields{
				"addr":        r.CC.GetAddr(),
				"commandName": r.CommandName,
				"database":    command.GetCommandDatabase(r.Command),
				"collection":  command.GetCommandCollection(r.Command),
			}).Warningf("Unauthenticated request")
		}

		identities = append(identities, plugins.NewStaticIdentity(Name, UNAUTHENTICATED_ROLE, UNAUTHENTICATED_ROLE))
	}

	// Expand the identities into roles
	for _, ident := range identities {
		for _, r := range ident.Roles() {
			if _, ok := rolesM[r]; ok {
				continue
			}
			roles = append(roles, r)
			rolesM[r] = struct{}{}
		}
	}

	q := p.a.Querier()
	authorizeResults := make([]authzlib.AuthorizeResult, 0, len(resourceMap))
	for method, resources := range resourceMap {
		for _, resource := range resources {
			authorizeResults = append(authorizeResults, q.Authorize(ctx, roles, method, resource))
		}
	}

	var identitiesStrings [][]string
	// Handle all log rules
	for _, result := range authorizeResults {
		for _, logRule := range result.LogOnlyRules {
			if identitiesStrings == nil {
				identitiesStrings = make([][]string, len(identities))
				for i, id := range identities {
					identitiesStrings[i] = []string{id.Type(), id.User()}
				}
			}
			logrus.NewEntry(logrus.StandardLogger()).WithFields(logrus.Fields{
				"identities": identitiesStrings,
				"policy":     logRule.PolicyName,
				"ruleNumber": logRule.RuleNumber,
				"effect":     logRule.Effect.String(),
				"message":    logRule.Message,
				"method":     result.AuthorizationMethod.String(),
				"resource":   result.Resource.String(),
			}).Infof("Authz LOGONLY")
		}
	}

	// Process authorization
	for _, result := range authorizeResults {
		// If no rule was found; then we need to handle the default case
		if result.Rule == nil {
			// TODO: move to schema annotations
			// Check the DB/Collection level defaults
			if p.conf.DenyByDefaultNamespaces != nil {
				deny, ok := p.conf.DenyByDefaultNamespaces[command.GetCommandDatabase(r.Command)+"."+command.GetCommandCollection(r.Command)]
				if ok {
					// If we  had a value; we honor that regardless
					if deny {
						authzDeny.WithLabelValues(command.GetCommandDatabase(r.Command), command.GetCommandCollection(r.Command), r.CommandName).Inc()
						return mongoerror.Unauthorized.ErrMessage("unauthorized"), nil
					} else {
						continue
					}
				}
			}
			// If nothing was found in the defaultNamespaces; continue with the global default
			if p.conf.DenyByDefault {
				authzDeny.WithLabelValues(command.GetCommandDatabase(r.Command), command.GetCommandCollection(r.Command), r.CommandName).Inc()
				return mongoerror.Unauthorized.ErrMessage("unauthorized"), nil
			}
			continue
		}

		// If a rule is found; enforce it
		if !result.Rule.Effect.IsAllow() {
			authzDeny.WithLabelValues(command.GetCommandDatabase(r.Command), command.GetCommandCollection(r.Command), r.CommandName).Inc()
			return mongoerror.Unauthorized.ErrMessage("unauthorized"), nil
		}
	}

	result, err := next(ctx, r)
	if cursorIDRaw, ok := bsonutil.Lookup(result, "cursor", "id"); ok {
		if cursorID, ok := cursorIDRaw.(int64); ok && cursorID > 0 {
			r.CursorCache.GetCursor(cursorID).Map[contextKeyResources] = resourceMap
		}
	}

	return result, err
}
