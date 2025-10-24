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
)

type LinkRepo interface {
	CreateLink(ctx context.Context, link domain.Link) (string, error)
	GetURLByAlias(ctx context.Context, alias string) (string, error)
}

type Link struct {
	repo LinkRepo
}

func New(repo LinkRepo) *Link {
	return &Link{repo: repo}
}

var (
	ErrAliasNotFound      = errors.New("alias not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")
)

func (l *Link) SaveLink(ctx context.Context, link dto.Link) (string, error) {
	const op = "service.link.Save"

	// TODO: Обработать случай коллизий при генерации рандомного алиаса
	alias := link.Alias
	if alias == "" {
		alias = random.NewString(6)
	}

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

func (l *Link) GetURLByAlias(ctx context.Context, alias string) (string, error) {
	const op = "service.link.GetURLByAlias"

	url, err := l.repo.GetURLByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repo.ErrAliasNotFound) {
			return "", errutils.Wrap(op, ErrAliasNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	return url, nil
}
