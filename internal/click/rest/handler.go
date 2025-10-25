package rest

import (
	"context"
	"github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/internal/response"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

type Click interface {
	GetClicksSummary(ctx context.Context, alias string) (dto.GetClicks, error)
}

type ClickHandler struct {
	click Click
}

func NewClickHandler(click Click) *ClickHandler {
	return &ClickHandler{click: click}
}

func (h *ClickHandler) GetAnalytics(c *ginext.Context) {
	alias := c.Param("alias")
	if alias == "" {
		response.Error("alias must not be empty.").WriteJSON(c, http.StatusBadRequest)
		return
	}

	summary, err := h.click.GetClicksSummary(c.Request.Context(), alias)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get click summary")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Raw(c, http.StatusOK, summary)
}
