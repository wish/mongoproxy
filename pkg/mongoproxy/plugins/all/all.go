package all

import (
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/authz"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/dedupe"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/defaults"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/filtercommand"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/insort"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/limits"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/mongo"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/opentracing"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/schema"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/slowlog"
	_ "github.com/wish/mongoproxy/pkg/mongoproxy/plugins/writeconcernoverride"
)
