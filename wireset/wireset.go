package wireset

import (
	"github.com/google/wire"
	"pkg.aiocean.dev/serviceutil/handler"
	"pkg.aiocean.dev/serviceutil/healthserver"
	"pkg.aiocean.dev/serviceutil/interceptor"
	"pkg.aiocean.dev/serviceutil/logger"
	"pkg.aiocean.dev/serviceutil/tracer"
)

var Default = wire.NewSet(
	logger.NewLogger,
	tracer.DefaultTracerSet,
	healthserver.WireSet,
	interceptor.WireSet,
	interceptor.DefaultWireSet,
	handler.WireSet,
)
