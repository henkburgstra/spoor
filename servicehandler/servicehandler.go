package servicehandler

import (
	"github.com/kardianos/service"
	"github.com/henkburgstra/spoor"
)

var logger service.Logger

type ServiceHandler struct {
	spoor.LogHandler
	service service.Service
}

func NewServiceHandler(service service.Service) *ServiceHandler {
	serviceHandler := new(ServiceHandler)
	serviceHandler.service = service
	serviceHandler.LogHandler = *spoor.NewLogHandler()
	return serviceHandler
}

func (h *ServiceHandler) Handle(logRecord *spoor.LogRecord) {
	h.Emit(logRecord)
}

func (h *ServiceHandler) Emit(logRecord *spoor.LogRecord) {
	msg := h.Format(logRecord)
	switch logRecord.GetLevel() {
	case spoor.WARNING:
		logger.Warning(msg)
	case spoor.ERROR:
		logger.Error(msg)
	default:
		logger.Info(msg)
	}
}
