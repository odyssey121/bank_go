package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var sqlcQueries *Queries

const (
	dbDriver         = "postgres"
	dbConnerctionStr = "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable"
)

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbConnerctionStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	sqlcQueries = New(conn)

	os.Exit(m.Run())

}
