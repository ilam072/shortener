package postgres

import (
	"context"
	"github.com/ilam072/shortener/internal/click/types/domain"
	"github.com/ilam072/shortener/pkg/errutils"
	"github.com/wb-go/wbf/dbpg"
)

type ClickRepo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *ClickRepo {
	return &ClickRepo{db: db}
}

func (r *ClickRepo) CreateClick(ctx context.Context, click domain.Click) error {
	const op = "repo.click.Create"

	query := `
		INSERT INTO clicks(id, alias, user_agent, client_name, device_type, ip)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		click.ID,
		click.Alias,
		click.UserAgent,
		click.Client,
		click.Device,
		click.IP,
	); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
