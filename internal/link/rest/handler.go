package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	clickdto "github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/internal/link/service"
	linkdto "github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/ilam072/shortener/internal/response"
	"github.com/mssola/user_agent"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net"
	"net/http"
	"strings"
)

type Link interface {
	SaveLink(ctx context.Context, link linkdto.Link) (string, error)
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
}

func NewLinkHandler(link Link, click Click, validator Validator) *LinkHandler {
	return &LinkHandler{link: link, click: click, validator: validator}
}

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
	alias, err := h.link.SaveLink(c.Request.Context(), link)
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

	if err := h.click.SaveClick(c.Request.Context(), click); err != nil {
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
