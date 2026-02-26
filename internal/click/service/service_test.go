package service_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilam072/shortener/internal/click/mocks"
	"github.com/ilam072/shortener/internal/click/service"
	"github.com/ilam072/shortener/internal/click/types/domain"
	"github.com/ilam072/shortener/internal/click/types/dto"
)

func TestClickService_SaveClick(t *testing.T) {
	type fields struct {
		setup func(repo *mocks.MockClickRepo)
	}
	type args struct {
		click dto.Click
	}
	type want struct {
		err bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "success",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						CreateClick(gomock.Any(), gomock.AssignableToTypeOf(domain.Click{})).
						Return(nil)
				},
			},
			args: args{
				click: dto.Click{
					Alias:     "abc",
					UserAgent: "ua",
					Client:    "chrome",
					Device:    "desktop",
					IP:        "127.0.0.1",
				},
			},
			want: want{err: false},
		},
		{
			name: "repo error",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						CreateClick(gomock.Any(), gomock.Any()).
						Return(errors.New("db error"))
				},
			},
			args: args{
				click: dto.Click{
					Alias: "abc",
				},
			},
			want: want{err: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockClickRepo(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockRepo)
			}

			svc := service.New(mockRepo)

			err := svc.SaveClick(context.Background(), tt.args.click)

			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClickService_GetClicksSummary(t *testing.T) {
	type fields struct {
		setup func(repo *mocks.MockClickRepo)
	}
	type want struct {
		err bool
	}

	tests := []struct {
		name   string
		alias  string
		fields fields
		want   want
	}{
		{
			name:  "success",
			alias: "abc",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						GetClicksByDay(gomock.Any(), "abc").
						Return([]domain.ClickRow{
							{Aggregation: "2025-01-01", Clicks: 10},
						}, nil)

					repo.EXPECT().
						GetClicksByMonth(gomock.Any(), "abc").
						Return([]domain.ClickRow{
							{Aggregation: "2025-01", Clicks: 100},
						}, nil)

					repo.EXPECT().
						GetClicksByUserAgent(gomock.Any(), "abc").
						Return([]domain.ClickRow{
							{Aggregation: "chrome", Clicks: 50},
						}, nil)
				},
			},
			want: want{err: false},
		},
		{
			name:  "error on get by day",
			alias: "abc",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						GetClicksByDay(gomock.Any(), "abc").
						Return(nil, errors.New("db error"))
				},
			},
			want: want{err: true},
		},
		{
			name:  "error on get by month",
			alias: "abc",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						GetClicksByDay(gomock.Any(), "abc").
						Return(nil, nil)

					repo.EXPECT().
						GetClicksByMonth(gomock.Any(), "abc").
						Return(nil, errors.New("db error"))
				},
			},
			want: want{err: true},
		},
		{
			name:  "error on get by user agent",
			alias: "abc",
			fields: fields{
				setup: func(repo *mocks.MockClickRepo) {
					repo.EXPECT().
						GetClicksByDay(gomock.Any(), "abc").
						Return(nil, nil)

					repo.EXPECT().
						GetClicksByMonth(gomock.Any(), "abc").
						Return(nil, nil)

					repo.EXPECT().
						GetClicksByUserAgent(gomock.Any(), "abc").
						Return(nil, errors.New("db error"))
				},
			},
			want: want{err: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockClickRepo(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockRepo)
			}

			svc := service.New(mockRepo)

			res, err := svc.GetClicksSummary(context.Background(), tt.alias)

			if tt.want.err {
				require.Error(t, err)
				require.Empty(t, res)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.alias, res.Alias)
		})
	}
}
