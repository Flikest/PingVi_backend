package skilla

import (
	"log"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

func MustSkyllaDBOpen() {
	cluster := gocql.NewCluster("127.0.0.1:9042")

	rawSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}

	session := gocqlx.WrapSession(rawSession)
	defer session.Close()

	log.Println("Успешно подключено к ScyllaDB через localhost!")

}
