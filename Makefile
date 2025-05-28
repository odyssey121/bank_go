migrateup:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose down

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