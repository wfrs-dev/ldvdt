package trainer

import (
	"github.com/labstack/echo/v5"

	"github.com/wfrs-dev/ldvdt/internal/domain"
)

// Module registers the trainer routes on the Echo instance.
type Module struct {
	handler *Handler
}

// NewModule builds the trainer module.
func NewModule(svc *Service, kb domain.Keyboard) *Module {
	return &Module{handler: NewHandler(svc, kb)}
}

// Register wires the trainer HTTP routes.
func (m *Module) Register(e *echo.Echo) {
	e.GET("/", m.handler.Index)
	e.GET("/practice", m.handler.Practice)
	e.POST("/practice", m.handler.Submit)
	e.GET("/progress", m.handler.Progress)
}
