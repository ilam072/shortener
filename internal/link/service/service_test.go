package service_test

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ilam072/shortener/internal/link/mocks"
	linkrepo "github.com/ilam072/shortener/internal/link/repo"
	"github.com/ilam072/shortener/internal/link/service"
	"github.com/ilam072/shortener/internal/link/types/domain"
	"github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

func TestLink_SaveLink(t *testing.T) {
	type fields struct {
		setup func(repo *mocks.MockLinkRepo)
	}
	type args struct {
		link dto.Link
	}
	type want struct {
		alias string
		err   error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "custom alias success",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo) {
					repo.EXPECT().
						CreateLink(gomock.Any(), gomock.AssignableToTypeOf(domain.Link{})).
						Return("custom", nil)
				},
			},
			args: args{
				link: dto.Link{
					URL:   "https://example.com",
					Alias: "custom",
				},
			},
			want: want{
				alias: "custom",
				err:   nil,
			},
		},
		{
			name: "custom alias already exists",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo) {
					repo.EXPECT().
						CreateLink(gomock.Any(), gomock.Any()).
						Return("", linkrepo.ErrAliasAlreadyExists)
				},
			},
			args: args{
				link: dto.Link{
					URL:   "https://example.com",
					Alias: "custom",
				},
			},
			want: want{
				alias: "",
				err:   service.ErrAliasAlreadyExists,
			},
		},
		{
			name: "auto alias success",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo) {
					repo.EXPECT().
						CreateLink(gomock.Any(), gomock.Any()).
						Return("abc123", nil)
				},
			},
			args: args{
				link: dto.Link{
					URL: "https://example.com",
				},
			},
			want: want{
				alias: "abc123",
				err:   nil,
			},
		},
		{
			name: "auto alias retry conflict then success",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo) {
					gomock.InOrder(
						repo.EXPECT().
							CreateLink(gomock.Any(), gomock.Any()).
							Return("", linkrepo.ErrAliasAlreadyExists),
						repo.EXPECT().
							CreateLink(gomock.Any(), gomock.Any()).
							Return("final123", nil),
					)
				},
			},
			args: args{
				link: dto.Link{
					URL: "https://example.com",
				},
			},
			want: want{
				alias: "final123",
				err:   nil,
			},
		},
		{
			name: "auto alias exhausted retry",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo) {
					repo.EXPECT().
						CreateLink(gomock.Any(), gomock.Any()).
						AnyTimes().
						Return("", linkrepo.ErrAliasAlreadyExists)
				},
			},
			args: args{
				link: dto.Link{
					URL: "https://example.com",
				},
			},
			want: want{
				alias: "",
				err:   service.ErrAliasAlreadyExists,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLinkRepo(ctrl)
			mockCache := mocks.NewMockLinkCache(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockRepo)
			}

			svc := service.New(mockRepo, mockCache)

			strategy := retry.Strategy{
				Attempts: 5,
				Delay:    40 * time.Millisecond,
				Backoff:  1.5,
			}

			gotAlias, err := svc.SaveLink(context.Background(), tt.args.link, strategy)

			if tt.want.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.want.err))
				require.Empty(t, gotAlias)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want.alias, gotAlias)
		})
	}
}

func TestLink_GetURLByAlias(t *testing.T) {
	type fields struct {
		setup func(repo *mocks.MockLinkRepo, cache *mocks.MockLinkCache)
	}
	type want struct {
		url string
		err error
	}

	tests := []struct {
		name   string
		alias  string
		fields fields
		want   want
	}{
		{
			name:  "cache hit",
			alias: "alias",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo, cache *mocks.MockLinkCache) {
					cache.EXPECT().
						GetURL(gomock.Any(), "alias").
						Return("https://example.com", nil)
				},
			},
			want: want{
				url: "https://example.com",
				err: nil,
			},
		},
		{
			name:  "cache miss repo success",
			alias: "alias",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo, cache *mocks.MockLinkCache) {
					gomock.InOrder(
						cache.EXPECT().
							GetURL(gomock.Any(), "alias").
							Return("", redis.NoMatches),
						repo.EXPECT().
							GetURLByAlias(gomock.Any(), "alias").
							Return("https://example.com", nil),
						cache.EXPECT().
							SetURL(gomock.Any(), "alias", "https://example.com").
							Return(nil),
					)
				},
			},
			want: want{
				url: "https://example.com",
				err: nil,
			},
		},
		{
			name:  "alias not found",
			alias: "alias",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo, cache *mocks.MockLinkCache) {
					cache.EXPECT().
						GetURL(gomock.Any(), "alias").
						Return("", redis.NoMatches)
					repo.EXPECT().
						GetURLByAlias(gomock.Any(), "alias").
						Return("", linkrepo.ErrAliasNotFound)
				},
			},
			want: want{
				url: "",
				err: service.ErrAliasNotFound,
			},
		},
		{
			name:  "repo unexpected error",
			alias: "alias",
			fields: fields{
				setup: func(repo *mocks.MockLinkRepo, cache *mocks.MockLinkCache) {
					cache.EXPECT().
						GetURL(gomock.Any(), "alias").
						Return("", redis.NoMatches)
					repo.EXPECT().
						GetURLByAlias(gomock.Any(), "alias").
						Return("", errors.New("db error"))
				},
			},
			want: want{
				url: "",
				err: errors.New("db error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLinkRepo(ctrl)
			mockCache := mocks.NewMockLinkCache(ctrl)

			if tt.fields.setup != nil {
				tt.fields.setup(mockRepo, mockCache)
			}

			svc := service.New(mockRepo, mockCache)

			gotURL, err := svc.GetURLByAlias(context.Background(), tt.alias)

			if tt.want.err != nil {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tt.want.err.Error()))
				require.Empty(t, gotURL)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want.url, gotURL)
		})
	}
}
