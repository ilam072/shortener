package db

import (
	"fmt"
	"github.com/ilam072/shortener/internal/config"
	"github.com/wb-go/wbf/dbpg"
)

func OpenDB(cfg config.DBConfig) (*dbpg.DB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.PgUser,
		cfg.PgPassword,
		cfg.PgHost,
		cfg.PgPort,
		cfg.PgDatabase,
	)
	db, err := dbpg.New(connString, nil, &dbpg.Options{
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
