package wireset

import (
	"github.com/google/wire"
	"pkg.aiocean.dev/serviceutil/datadogtrace"
	"pkg.aiocean.dev/serviceutil/handler"
	"pkg.aiocean.dev/serviceutil/healthserver"
	"pkg.aiocean.dev/serviceutil/interceptor"
	"pkg.aiocean.dev/serviceutil/logger"
)

var Default = wire.NewSet(
	logger.NewLogger,
	datadogtrace.DefaultDataDogTraceWireSet,
	healthserver.WireSet,
	interceptor.DefaultWireSet,
	handler.WireSet,
)
