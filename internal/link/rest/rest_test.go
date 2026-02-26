package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/wb-go/wbf/ginext"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/require"

	clickdto "github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/internal/link/mocks"
	"github.com/ilam072/shortener/internal/link/rest"
	"github.com/ilam072/shortener/internal/link/service"
	linkdto "github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/wb-go/wbf/retry"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(method, path string, body []byte) (*ginext.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	c.Request = req

	return c, w
}

func TestLinkHandler_CreateLink(t *testing.T) {
	type fields struct {
		setup func(link *mocks.MockLink, validator *mocks.MockValidator)
	}
	type want struct {
		status int
	}

	tests := []struct {
		name   string
		body   interface{}
		fields fields
		want   want
	}{
		{
			name: "invalid json",
			body: "invalid",
			want: want{status: http.StatusBadRequest},
		},
		{
			name: "validation error",
			body: linkdto.Link{URL: "bad"},
			fields: fields{
				setup: func(link *mocks.MockLink, validator *mocks.MockValidator) {
					validator.EXPECT().
						Validate(gomock.Any()).
						Return(errors.New("validation failed"))
				},
			},
			want: want{status: http.StatusBadRequest},
		},
		{
			name: "alias already exists",
			body: linkdto.Link{URL: "https://example.com", Alias: "abc"},
			fields: fields{
				setup: func(link *mocks.MockLink, validator *mocks.MockValidator) {
					validator.EXPECT().
						Validate(gomock.Any()).
						Return(nil)
					link.EXPECT().
						SaveLink(gomock.Any(), gomock.Any(), gomock.Any()).
						Return("", service.ErrAliasAlreadyExists)
				},
			},
			want: want{status: http.StatusConflict},
		},
		{
			name: "internal error",
			body: linkdto.Link{URL: "https://example.com"},
			fields: fields{
				setup: func(link *mocks.MockLink, validator *mocks.MockValidator) {
					validator.EXPECT().
						Validate(gomock.Any()).
						Return(nil)
					link.EXPECT().
						SaveLink(gomock.Any(), gomock.Any(), gomock.Any()).
						Return("", errors.New("db error"))
				},
			},
			want: want{status: http.StatusInternalServerError},
		},
		{
			name: "success",
			body: linkdto.Link{URL: "https://example.com"},
			fields: fields{
				setup: func(link *mocks.MockLink, validator *mocks.MockValidator) {
					validator.EXPECT().
						Validate(gomock.Any()).
						Return(nil)
					link.EXPECT().
						SaveLink(gomock.Any(), gomock.Any(), gomock.Any()).
						Return("abc123", nil)
				},
			},
			want: want{status: http.StatusCreated},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLink := mocks.NewMockLink(ctrl)
			mockClick := mocks.NewMockClick(ctrl)
			mockValidator := mocks.NewMockValidator(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockLink, mockValidator)
			}
			strategy := retry.Strategy{}
			handler := rest.NewLinkHandler(mockLink, mockClick, mockValidator, strategy)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			c, w := newTestContext(http.MethodPost, "/links", bodyBytes)

			handler.CreateLink(c)

			require.Equal(t, tt.want.status, w.Code)
		})
	}
}

func TestLinkHandler_Redirect(t *testing.T) {
	type fields struct {
		setup func(link *mocks.MockLink, click *mocks.MockClick)
	}
	type want struct {
		status int
	}

	tests := []struct {
		name   string
		alias  string
		fields fields
		want   want
	}{
		{
			name:  "empty alias",
			alias: "",
			want:  want{status: http.StatusBadRequest},
		},
		{
			name:  "alias not found",
			alias: "abc",
			fields: fields{
				setup: func(link *mocks.MockLink, click *mocks.MockClick) {
					link.EXPECT().
						GetURLByAlias(gomock.Any(), "abc").
						Return("", service.ErrAliasNotFound)
				},
			},
			want: want{status: http.StatusNotFound},
		},
		{
			name:  "internal error",
			alias: "abc",
			fields: fields{
				setup: func(link *mocks.MockLink, click *mocks.MockClick) {
					link.EXPECT().
						GetURLByAlias(gomock.Any(), "abc").
						Return("", errors.New("db error"))
				},
			},
			want: want{status: http.StatusInternalServerError},
		},
		{
			name:  "success",
			alias: "abc",
			fields: fields{
				setup: func(link *mocks.MockLink, click *mocks.MockClick) {
					link.EXPECT().
						GetURLByAlias(gomock.Any(), "abc").
						Return("https://example.com", nil)

					click.EXPECT().
						SaveClick(gomock.Any(), gomock.AssignableToTypeOf(clickdto.Click{})).
						Return(nil)
				},
			},
			want: want{status: http.StatusFound},
		},
		{
			name:  "click save error does not break redirect",
			alias: "abc",
			fields: fields{
				setup: func(link *mocks.MockLink, click *mocks.MockClick) {
					link.EXPECT().
						GetURLByAlias(gomock.Any(), "abc").
						Return("https://example.com", nil)

					click.EXPECT().
						SaveClick(gomock.Any(), gomock.Any()).
						Return(errors.New("click error"))
				},
			},
			want: want{status: http.StatusFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLink := mocks.NewMockLink(ctrl)
			mockClick := mocks.NewMockClick(ctrl)
			mockValidator := mocks.NewMockValidator(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockLink, mockClick)
			}

			strategy := retry.Strategy{}
			handler := rest.NewLinkHandler(mockLink, mockClick, mockValidator, strategy)

			c, w := newTestContext(http.MethodGet, "/"+tt.alias, nil)
			c.Params = gin.Params{{Key: "alias", Value: tt.alias}}

			handler.Redirect(c)

			require.Equal(t, tt.want.status, w.Code)
		})
	}
}
