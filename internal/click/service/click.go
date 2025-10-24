package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilam072/shortener/internal/click/types/domain"
	"github.com/ilam072/shortener/internal/click/types/dto"
	"github.com/ilam072/shortener/pkg/errutils"
)

type ClickRepo interface {
	CreateClick(ctx context.Context, click domain.Click) error
}

type Click struct {
	repo ClickRepo
}

func New(repo ClickRepo) *Click {
	return &Click{repo: repo}
}

func (c *Click) SaveClick(ctx context.Context, click dto.Click) error {
	const op = "service.click.Save"

	domainClick := domain.Click{
		ID:        uuid.New(),
		Alias:     click.Alias,
		UserAgent: click.UserAgent,
		Client:    click.Client,
		Device:    click.Device,
		IP:        click.IP,
	}

	if err := c.repo.CreateClick(ctx, domainClick); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
