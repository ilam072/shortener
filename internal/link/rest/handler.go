package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ilam072/shortener/internal/link/service"
	"github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/ilam072/shortener/internal/response"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

type Link interface {
	SaveLink(ctx context.Context, link dto.Link) (string, error)
}

type Validator interface {
	Validate(i interface{}) error
}

type LinkHandler struct {
	link      Link
	validator Validator
}

func NewLinkHandler(link Link, validator Validator) *LinkHandler {
	return &LinkHandler{link: link, validator: validator}
}

func (h *LinkHandler) CreateLink(c *ginext.Context) {
	var link dto.Link
	if err := json.NewDecoder(c.Request.Body).Decode(&link); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to decode request body")
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
			zlog.Logger.Error().Err(err).Msg("alias already exists")
			response.Error("url with such alias already exists").WriteJSON(c, http.StatusConflict)
			return
		}
		zlog.Logger.Error().Err(err).Str("alias", link.Alias).Msg("failed to save short link")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Success(alias).WriteJSON(c, http.StatusCreated)
}
