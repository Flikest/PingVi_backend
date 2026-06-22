package repository

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/scylladb/gocqlx/v2"
)

type RepositorySSO struct {
	Log *slog.Logger
	DB  *pgxpool.Pool
}

type RepositoryMessenger struct {
	Log     *slog.Logger
	Session *gocqlx.Session
}

type RepossitoryS3 struct {
	Log         *slog.Logger
	MinIOClient *minio.Client
}

func NewRepositorySSO(s *RepositorySSO) *RepositorySSO {
	return &RepositorySSO{
		Log: s.Log,
		DB:  s.DB,
	}
}

func NewRepositoryMessenger(s *RepositoryMessenger) *RepositoryMessenger {
	return &RepositoryMessenger{
		Log:     s.Log,
		Session: s.Session,
	}
}

func NewReposossoryS3(s *RepossoryS3) *RepossoryS3 {
	return &RepossoryS3{
		Log: s.Log,
	}
}
