package scylla

import (
	"fmt"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

func MustScyllaDBOpen() *gocqlx.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Port = 9042

	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 5 * time.Second

	username := os.Getenv("SCYLLADB_USERNAME")
	password := os.Getenv("SCYLLADB_PASSWORD")

	if username == "" || password == "" {
		panic("SCYLLA_DB_USERNAME and SCYLLA_DB_PASSWORD must be set")
	}

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}

	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		panic(fmt.Errorf("error with wrap session: %w", err))
	}

	return &session
}
