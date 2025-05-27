package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/bank_go/util"
	_ "github.com/lib/pq"
)

var sqlcQueries *Queries
var testDb *sql.DB

func TestMain(m *testing.M) {
	var err error

	cfg, err := util.LoadConfig("../../config")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	testDb, err = sql.Open(cfg.DbDriver, cfg.DbConnectionString)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	sqlcQueries = New(testDb)

	os.Exit(m.Run())

}
