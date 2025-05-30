migratecreate:
	migrate create -dir db/migration -ext sql -seq 6 $(name)

migrateup:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose down 1

clear:
	clear

test:
	go test -v -cover ./...

sqlc:
	sqlc generate

server:
	nodemon --exec go run main.go --ext go --signal SIGTERM

mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/bank_go/db/sqlc Store

# not file in dir
.PHONY: clear migrateup migratedown test server sqlc mock