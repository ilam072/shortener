package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/shortener/internal/link/repo"
	"github.com/ilam072/shortener/internal/link/types/domain"
	"github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/ilam072/shortener/pkg/errutils"
	"github.com/ilam072/shortener/pkg/random"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

//go:generate mockgen -source=link.go -destination=../mocks/service_mocks.go -package=mocks
type LinkRepo interface {
	CreateLink(ctx context.Context, link domain.Link) (string, error)
	GetURLByAlias(ctx context.Context, alias string) (string, error)
}

type LinkCache interface {
	SetURL(ctx context.Context, alias string, url string) error
	GetURL(ctx context.Context, alias string) (string, error)
}

type Link struct {
	repo  LinkRepo
	cache LinkCache
}

func New(repo LinkRepo, cache LinkCache) *Link {
	return &Link{repo: repo, cache: cache}
}

var (
	ErrAliasNotFound      = errors.New("alias not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")
)

func (l *Link) SaveLink(ctx context.Context, link dto.Link, strategy retry.Strategy) (string, error) {
	const op = "service.link.Save"

	alias := link.Alias
	if alias != "" {
		domainLink := domain.Link{
			ID:    uuid.New(),
			URL:   link.URL,
			Alias: alias,
		}
		resAlias, err := l.repo.CreateLink(ctx, domainLink)
		if err != nil {
			if errors.Is(err, repo.ErrAliasAlreadyExists) {
				return "", ErrAliasAlreadyExists
			}
			return "", errutils.Wrap(op, err)
		}
		return resAlias, nil
	}

	var resAlias string
	err := retry.Do(func() error {
		tmpAlias := random.NewString(6)
		domainLink := domain.Link{
			ID:    uuid.New(),
			URL:   link.URL,
			Alias: tmpAlias,
		}

		var err error
		resAlias, err = l.repo.CreateLink(ctx, domainLink)
		if err != nil {
			if errors.Is(err, repo.ErrAliasAlreadyExists) {
				return err
			}
			return errutils.Wrap(op, err)
		}
		return nil
	}, strategy)

	if err != nil {
		if errors.Is(err, repo.ErrAliasAlreadyExists) {
			return "", ErrAliasAlreadyExists
		}
		return "", err
	}

	return resAlias, nil
}

func (l *Link) GetURLByAlias(ctx context.Context, alias string) (string, error) {
	const op = "service.link.GetURLByAlias"

	url, err := l.cache.GetURL(ctx, alias)
	if err == nil {
		return url, nil
	}
	if !errors.Is(err, redis.NoMatches) {
		zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to get url from cache")
	}

	url, err = l.repo.GetURLByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repo.ErrAliasNotFound) {
			return "", errutils.Wrap(op, ErrAliasNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	if err = l.cache.SetURL(ctx, alias, url); err != nil {
		zlog.Logger.Error().Err(err).Str("alias", alias).Str("url", url).Msg("failed to cache url")
	}

	return url, nil
}
