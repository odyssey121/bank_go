package main

import (
	"database/sql"
	"log"

	"github.com/bank_go/api"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/util"
	_ "github.com/lib/pq"
)

func main() {
	var err error

	cfg, err := util.LoadConfig("./.config")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(cfg.DbDriver, cfg.DbConnectionString)

	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(store, cfg)

	if err != nil {
		log.Fatal("cannot create to new server: ", err)
	}

	err = server.Start(cfg.WebServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

}
