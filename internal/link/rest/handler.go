package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/ilam072/shortener/internal/click/types/dto"
	clickdto "github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/internal/link/service"
	_ "github.com/ilam072/shortener/internal/link/types/dto"
	linkdto "github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/ilam072/shortener/internal/response"
	"github.com/mssola/user_agent"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"net"
	"net/http"
	"strings"
)

//go:generate mockgen -source=handler.go -destination=../mocks/rest_mocks.go -package=mocks
type Link interface {
	SaveLink(ctx context.Context, link linkdto.Link, strategy retry.Strategy) (string, error)
	GetURLByAlias(ctx context.Context, alias string) (string, error)
}

type Click interface {
	SaveClick(ctx context.Context, click clickdto.Click) error
}

type Validator interface {
	Validate(i interface{}) error
}

type LinkHandler struct {
	link      Link
	click     Click
	validator Validator
	strategy  retry.Strategy
}

func NewLinkHandler(link Link, click Click, validator Validator, strategy retry.Strategy) *LinkHandler {
	return &LinkHandler{link: link, click: click, validator: validator, strategy: strategy}
}

// CreateLink godoc
// @Summary Создать короткую ссылку
// @Description Создаёт новую короткую ссылку. Alias можно передать вручную или он будет сгенерирован автоматически
// @Tags Links
// @Accept json
// @Produce json
// @Param input body dto.Link true "Данные для создания ссылки"
// @Success 201 {object} response.Response "alias созданной ссылки"
// @Failure 400 {object} response.Response "invalid request body или validation error"
// @Failure 409 {object} response.Response "alias already exists"
// @Failure 500 {object} response.Response "internal server error"
// @Router /shorten [post]
func (h *LinkHandler) CreateLink(c *ginext.Context) {
	var link linkdto.Link
	if err := json.NewDecoder(c.Request.Body).Decode(&link); err != nil {
		response.Error("invalid request body").WriteJSON(c, http.StatusBadRequest)
		return
	}
	if err := h.validator.Validate(link); err != nil {
		response.Error(fmt.Sprintf("validation error: %s", err.Error())).WriteJSON(c, http.StatusBadRequest)
		return
	}
	alias, err := h.link.SaveLink(c.Request.Context(), link, h.strategy)
	if err != nil {
		if errors.Is(err, service.ErrAliasAlreadyExists) {
			response.Error("url with such alias already exists").WriteJSON(c, http.StatusConflict)
			return
		}
		zlog.Logger.Error().Err(err).Str("alias", link.Alias).Msg("failed to save short link")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}
	response.Success(alias).WriteJSON(c, http.StatusCreated)
}

// Redirect godoc
// @Summary Редирект по короткой ссылке
// @Description Перенаправляет пользователя на оригинальный URL по alias и сохраняет информацию о клике
// @Tags Links
// @Param alias path string true "Alias ссылки"
// @Success 302 "Redirect to original URL"
// @Failure 400 {object} response.Response "alias must not be empty"
// @Failure 404 {object} response.Response "alias not found"
// @Failure 500 {object} response.Response "internal server error"
// @Router /s/{alias} [get]
func (h *LinkHandler) Redirect(c *ginext.Context) {
	alias := c.Param("alias")
	if alias == "" {
		response.Error("alias must not be empty.").WriteJSON(c, http.StatusBadRequest)
		return
	}

	url, err := h.link.GetURLByAlias(c.Request.Context(), alias)
	if err != nil {
		if errors.Is(err, service.ErrAliasNotFound) {
			response.Error("alias not found").WriteJSON(c, http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get url by alias")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	userAgent := c.GetHeader("User-Agent")
	client, device := parseClientInfo(userAgent)

	click := clickdto.Click{
		Alias:     alias,
		UserAgent: userAgent,
		Client:    client,
		Device:    device,
		IP:        getClientIP(c),
	}

	if err = h.click.SaveClick(c.Request.Context(), click); err != nil {
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to save click")
	}

	http.Redirect(c.Writer, c.Request, url, http.StatusFound)
}

func getClientIP(c *ginext.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

func parseClientInfo(uaString string) (string, string) {
	ua := user_agent.New(uaString)
	name, _ := ua.Browser()
	device := "desktop"
	if ua.Mobile() {
		device = "mobile"
	} else if ua.Bot() {
		device = "bot"
	}
	return name, device
}
