package wireset

import (
	"github.com/aiocean/serviceutil/handler"
	"github.com/aiocean/serviceutil/healthserver"
	"github.com/aiocean/serviceutil/interceptor"
	"github.com/aiocean/serviceutil/logger"
	"github.com/aiocean/serviceutil/tracer"
	"github.com/google/wire"
)

var Default = wire.NewSet(
	logger.NewLogger,
	tracer.DefaultTracerSet,
	healthserver.WireSet,
	interceptor.WireSet,
	interceptor.DefaultWireSet,
	handler.WireSet,
)
