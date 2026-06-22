package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConig struct {
	Ctx      context.Context
	ConnPath string
}

func MustPostgresDBOpen(pc *PostgresConig) *pgxpool.Pool {
	db, err := pgxpool.New(pc.Ctx, pc.ConnPath)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to db ❌:%s", err))
	}

	return db
}
