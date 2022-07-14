package datadogtrace

import (
	"github.com/google/wire"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var DefaultDataDogTraceWireSet = wire.NewSet(
	NewDataDogTrace,
)

type DataDogTrace struct {
}

func NewDataDogTrace() (*DataDogTrace, func(), error) {
	ddTrace := &DataDogTrace{}

	return ddTrace, ddTrace.StopTrace, nil
}

// startTrace

func (dd *DataDogTrace) StartTrace() {
	tracer.Start()
}

func (dd *DataDogTrace) StopTrace() {
	tracer.Stop()
}
