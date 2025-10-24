package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ilam072/shortener/internal/link/repo"
	"github.com/ilam072/shortener/internal/link/types/domain"
	"github.com/ilam072/shortener/pkg/errutils"
	"github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
)

type LinkRepo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) CreateLink(ctx context.Context, link domain.Link) (string, error) {
	const op = "repo.link.Create"

	query := `
		INSERT INTO links(id, url, alias)
		VALUES ($1,$2,$3)
		RETURNING alias;
	`

	var alias string
	if err := r.db.QueryRowContext(ctx, query, link.ID, link.URL, link.Alias).Scan(&alias); err != nil {
		if isUniqueViolation(err) {
			return "", errutils.Wrap(op, repo.ErrAliasAlreadyExists)
		}
		return "", errutils.Wrap(op, err)
	}

	return alias, nil
}

func (r *LinkRepo) GetURLByAlias(ctx context.Context, alias string) (string, error) {
	const op = "repo.link.GetURLByAlias"

	query := `
		SELECT url
		FROM links
		WHERE alias = $1
		LIMIT 1;
	`

	var url string
	if err := r.db.QueryRowContext(ctx, query, alias).Scan(&url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errutils.Wrap(op, repo.ErrAliasNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	return url, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
