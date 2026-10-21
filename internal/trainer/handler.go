package trainer

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"

	"github.com/wfrs-dev/ldvdt/internal/domain"
	"github.com/wfrs-dev/ldvdt/internal/web/ui"
)

// SubmitRequest is the JSON payload sent by the client typing engine.
type SubmitRequest struct {
	Level       int     `json:"level"`
	Lang        string  `json:"lang"`
	WordCount   int     `json:"wordCount"`
	WPM         float64 `json:"wpm"`
	Accuracy    float64 `json:"accuracy"`
	Consistency float64 `json:"consistency"`
	DurationMS  int64   `json:"durationMS"`
}

// Handler is the HTTP handler for the trainer routes.
type Handler struct {
	svc *Service
	kb  domain.Keyboard
}

// NewHandler builds the handler.
func NewHandler(svc *Service, kb domain.Keyboard) *Handler {
	return &Handler{svc: svc, kb: kb}
}

// Index renders the full SPA page.
func (h *Handler) Index(c *echo.Context) error {
	stats, err := h.svc.Progress(c.Request().Context(), "")
	if err != nil {
		stats = domain.Stats{}
	}
	return renderComponent(c, ui.Page(h.kb, h.svc.Levels(), stats))
}

// Practice renders the words fragment for a practice session.
func (h *Handler) Practice(c *echo.Context) error {
	level, _ := strconv.Atoi(c.QueryParam("level"))
	lang := domain.Language(c.QueryParam("lang"))
	count, _ := strconv.Atoi(c.QueryParam("count"))

	words, lvl, err := h.svc.Practice(level, lang, count)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return renderComponent(c, ui.Practice(words, lvl.Number, string(lang), len(words), string(lvl.NewChars)))
}

// Submit persists a completed practice and returns the updated stats as JSON.
func (h *Handler) Submit(c *echo.Context) error {
	var req SubmitRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cuerpo inválido"})
	}

	stats, err := h.svc.SaveResult(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"saved": true, "stats": stats})
}

// Progress renders the progress panel fragment.
func (h *Handler) Progress(c *echo.Context) error {
	lang := domain.Language(c.QueryParam("lang"))
	stats, err := h.svc.Progress(c.Request().Context(), lang)
	if err != nil {
		stats = domain.Stats{}
	}
	return renderComponent(c, ui.Progress(stats))
}

func renderComponent(c *echo.Context, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(c.Request().Context(), &buf); err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, buf.Bytes())
}
