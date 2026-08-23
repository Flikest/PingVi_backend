package repository

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2"
)

type RepositorySSO struct {
	Log *slog.Logger
	DB  *pgxpool.Pool
	RDB *redis.Client
}

type RepositoryMessenger struct {
	Log     *slog.Logger
	Session *gocqlx.Session
}

type RepositoryFileStorage struct {
	Log         *slog.Logger
	MinIOClient *minio.Client
}

type RepositoryBots struct {
	Log *slog.Logger
	DB  *pgxpool.Pool
	RDB *redis.Client
}

func NewRepositorySSO(r *RepositorySSO) *RepositorySSO {
	return &RepositorySSO{
		Log: r.Log,
		DB:  r.DB,
	}
}

func NewRepositoryMessenger(r *RepositoryMessenger) *RepositoryMessenger {
	return &RepositoryMessenger{
		Log:     r.Log,
		Session: r.Session,
	}
}

func NewReposossoryFileStorage(r *RepositoryFileStorage) *RepositoryFileStorage {
	return &RepositoryFileStorage{
		Log:         r.Log,
		MinIOClient: r.MinIOClient,
	}
}

func NewRepositoryBots(r *RepositoryBots) *RepositoryBots {
	return &RepositoryBots{
		Log: r.Log,
		DB:  r.DB,
		RDB: r.RDB,
	}
}
