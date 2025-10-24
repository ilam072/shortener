package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/shortener/internal/link/repo"
	"github.com/ilam072/shortener/internal/link/types/domain"
	"github.com/ilam072/shortener/internal/link/types/dto"
	"github.com/ilam072/shortener/pkg/errutils"
)

type LinkRepo interface {
	CreateLink(ctx context.Context, link domain.Link) (string, error)
}

type Link struct {
	repo LinkRepo
}

func New(repo LinkRepo) *Link {
	return &Link{repo: repo}
}

var ErrAliasAlreadyExists = errors.New("alias already exists")

func (l *Link) CreateLink(ctx context.Context, link dto.Link) (string, error) {
	const op = "service.link.Create"

	domainLink := domain.Link{
		ID:    uuid.New(),
		URL:   link.URL,
		Alias: link.Alias,
	}

	alias, err := l.repo.CreateLink(ctx, domainLink)
	if err != nil {
		if errors.Is(err, repo.ErrAliasAlreadyExists) {
			return "", ErrAliasAlreadyExists
		}
		return "", errutils.Wrap(op, err)
	}

	return alias, nil
}
