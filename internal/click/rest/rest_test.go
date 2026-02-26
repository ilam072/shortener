package rest_test

import (
	"encoding/json"
	"errors"
	"github.com/wb-go/wbf/ginext"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ilam072/shortener/internal/click/mocks"
	"github.com/ilam072/shortener/internal/click/rest"
	"github.com/ilam072/shortener/internal/click/types/dto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(method, path string) (*ginext.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, nil)
	c.Request = req

	return c, w
}

func TestClickHandler_GetAnalytics(t *testing.T) {
	type fields struct {
		setup func(click *mocks.MockClick)
	}
	type want struct {
		status int
		check  func(t *testing.T, body []byte)
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
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:  "service error",
			alias: "abc",
			fields: fields{
				setup: func(click *mocks.MockClick) {
					click.EXPECT().
						GetClicksSummary(gomock.Any(), "abc").
						Return(dto.GetClicks{}, errors.New("db error"))
				},
			},
			want: want{
				status: http.StatusInternalServerError,
			},
		},
		{
			name:  "success",
			alias: "abc",
			fields: fields{
				setup: func(click *mocks.MockClick) {
					click.EXPECT().
						GetClicksSummary(gomock.Any(), "abc").
						Return(dto.GetClicks{
							Alias: "abc",
							ByDay: []dto.ClicksByDay{
								{Date: "2025-01-01", Clicks: 10},
							},
						}, nil)
				},
			},
			want: want{
				status: http.StatusOK,
				check: func(t *testing.T, body []byte) {
					var res dto.GetClicks
					err := json.Unmarshal(body, &res)
					require.NoError(t, err)
					require.Equal(t, "abc", res.Alias)
					require.Len(t, res.ByDay, 1)
					require.Equal(t, 10, res.ByDay[0].Clicks)
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClick := mocks.NewMockClick(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockClick)
			}

			handler := rest.NewClickHandler(mockClick)

			c, w := newTestContext(http.MethodGet, "/analytics/"+tt.alias)
			c.Params = gin.Params{{Key: "alias", Value: tt.alias}}

			handler.GetAnalytics(c)

			require.Equal(t, tt.want.status, w.Code)

			if tt.want.check != nil {
				tt.want.check(t, w.Body.Bytes())
			}
		})
	}
}
