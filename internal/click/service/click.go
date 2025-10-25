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
	GetClicksByDay(ctx context.Context, alias string) ([]domain.ClickRow, error)
	GetClicksByMonth(ctx context.Context, alias string) ([]domain.ClickRow, error)
	GetClicksByUserAgent(ctx context.Context, alias string) ([]domain.ClickRow, error)
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

func (c *Click) GetClicksSummary(ctx context.Context, alias string) (dto.GetClicks, error) {
	const op = "service.click.GetClicksSummary"

	byDay, err := c.repo.GetClicksByDay(ctx, alias)
	if err != nil {
		return dto.GetClicks{}, errutils.Wrap(op, err)
	}

	byMonth, err := c.repo.GetClicksByMonth(ctx, alias)
	if err != nil {
		return dto.GetClicks{}, errutils.Wrap(op, err)
	}

	byUserAgent, err := c.repo.GetClicksByUserAgent(ctx, alias)
	if err != nil {
		return dto.GetClicks{}, errutils.Wrap(op, err)
	}

	return dto.GetClicks{
		Alias:       alias,
		ByDay:       mapToClicksByDay(byDay),
		ByMonth:     mapToClicksByMonth(byMonth),
		ByUserAgent: mapToClicksByUserAgent(byUserAgent),
	}, nil
}

func mapToClicksByDay(rows []domain.ClickRow) []dto.ClicksByDay {
	result := make([]dto.ClicksByDay, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ClicksByDay{
			Date:   row.Aggregation,
			Clicks: row.Clicks,
		})
	}
	return result
}

func mapToClicksByMonth(rows []domain.ClickRow) []dto.ClicksByMonth {
	result := make([]dto.ClicksByMonth, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ClicksByMonth{
			Month:  row.Aggregation,
			Clicks: row.Clicks,
		})
	}
	return result
}

func mapToClicksByUserAgent(rows []domain.ClickRow) []dto.ClicksByUserAgent {
	result := make([]dto.ClicksByUserAgent, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ClicksByUserAgent{
			UserAgent: row.Aggregation,
			Clicks:    row.Clicks,
		})
	}
	return result
}
