DB_URL="postgresql://username:pass@localhost:5432/bank_go?sslmode=disable"


migratecreate:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database  -verbose up

migrateup1:
	migrate -path db/migration -database $(DB_URL) -verbose up 1

migratedown:
	migrate -path db/migration -database $(DB_URL) -verbose down

migratedown1:
	migrate -path db/migration -database $(DB_URL) -verbose down 1

clear:
	clear

test:
	go test -v -cover --short ./...

sqlc:
	sqlc generate

server:
	nodemon --exec go run main.go --ext go --signal SIGTERM

mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/bank_go/db/sqlc Store

proto:
	rm -fv pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    proto/*.proto

evans:
	evans -r repl --host localhost --port 9090

# not file in dir
.PHONY: clear migrateup migratedown test server sqlc mock migrateup1 proto evans