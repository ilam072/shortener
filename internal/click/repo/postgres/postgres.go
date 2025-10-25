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

func (r *ClickRepo) GetClicksByDay(ctx context.Context, alias string) ([]domain.ClickRow, error) {
	const op = "repo.click.GetByDay"

	query := `
		SELECT DATE(clicked_at)::text AS aggregation, COUNT(*) AS clicks
		FROM clicks
		WHERE alias = $1
		GROUP BY DATE(clicked_at)
		ORDER BY aggregation;
	`

	rows, err := r.db.QueryContext(ctx, query, alias)
	if err != nil {
		return nil, errutils.Wrap(op, err)
	}
	defer rows.Close()

	var clicks []domain.ClickRow
	for rows.Next() {
		var row domain.ClickRow
		if err := rows.Scan(&row.Aggregation, &row.Clicks); err != nil {
			return nil, errutils.Wrap(op, err)
		}
		clicks = append(clicks, row)
	}

	return clicks, nil
}

func (r *ClickRepo) GetClicksByMonth(ctx context.Context, alias string) ([]domain.ClickRow, error) {
	const op = "repo.click.GetByMonth"

	query := `
		SELECT TO_CHAR(DATE_TRUNC('month', clicked_at), 'YYYY-MM') AS aggregation, COUNT(*) AS clicks
		FROM clicks
		WHERE alias = $1
		GROUP BY TO_CHAR(DATE_TRUNC('month', clicked_at), 'YYYY-MM')
		ORDER BY aggregation;
	`

	rows, err := r.db.QueryContext(ctx, query, alias)
	if err != nil {
		return nil, errutils.Wrap(op, err)
	}
	defer rows.Close()

	var clicks []domain.ClickRow
	for rows.Next() {
		var row domain.ClickRow
		if err := rows.Scan(&row.Aggregation, &row.Clicks); err != nil {
			return nil, errutils.Wrap(op, err)
		}
		clicks = append(clicks, row)
	}

	return clicks, nil
}

func (r *ClickRepo) GetClicksByUserAgent(ctx context.Context, alias string) ([]domain.ClickRow, error) {
	const op = "repo.click.GetByUserAgent"

	query := `
		SELECT client_name AS aggregation, COUNT(*) AS clicks
		FROM clicks
		WHERE alias = $1
		GROUP BY client_name
		ORDER BY clicks DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, alias)
	if err != nil {
		return nil, errutils.Wrap(op, err)
	}
	defer rows.Close()

	var clicks []domain.ClickRow
	for rows.Next() {
		var row domain.ClickRow
		if err := rows.Scan(&row.Aggregation, &row.Clicks); err != nil {
			return nil, errutils.Wrap(op, err)
		}
		clicks = append(clicks, row)
	}

	return clicks, nil
}
