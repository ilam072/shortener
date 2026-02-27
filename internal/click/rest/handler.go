package rest

import (
	"context"
	"github.com/ilam072/shortener/internal/click/types/dto"
	_ "github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/internal/response"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

//go:generate mockgen -source=handler.go -destination=../mocks/rest_mocks.go -package=mocks
type Click interface {
	GetClicksSummary(ctx context.Context, alias string) (dto.GetClicks, error)
}

type ClickHandler struct {
	click Click
}

func NewClickHandler(click Click) *ClickHandler {
	return &ClickHandler{click: click}
}

// GetAnalytics godoc
// @Summary Получить аналитику по ссылке
// @Description Возвращает статистику кликов по alias: по дням, месяцам и user-agent
// @Tags Analytics
// @Produce json
// @Param alias path string true "Alias ссылки"
// @Success 200 {object} dto.GetClicks "Статистика кликов"
// @Failure 400 {object} response.Response "alias must not be empty"
// @Failure 500 {object} response.Response "internal server error"
// @Router /analytics/{alias} [get]
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
