package main

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/bank_go/db/sqlc"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("start working!")
	conn, err := sql.Open("postgres", "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable")
	if err != nil {
		fmt.Println("err => ", err)
	}
	fmt.Println("conn =", conn)
	sqlcQuery := db.New(conn)
	fmt.Println("d:", sqlcQuery)
	res, err := sqlcQuery.CreateAccount(context.Background(), db.CreateAccountParams{Owner: "test", Balance: 1, Currency: "RUB"})

	if err != nil {
		fmt.Println("err CreateAccount= ", err)
	}

	fmt.Println("res:", res)

}
