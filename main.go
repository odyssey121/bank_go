package main

import (
	"database/sql"
	"log"

	"github.com/bank_go/api"
	db "github.com/bank_go/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver         = "postgres"
	dbConnerctionStr = "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable"
)

func main() {
	var err error
	conn, err := sql.Open(dbDriver, dbConnerctionStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start("localhost:8080")
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
