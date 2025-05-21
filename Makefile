migrateup:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://username:pass@localhost:5432/bank_go?sslmode=disable" -verbose down

clear:
	clear

# not file in dir
.PHONY: clear migrateup migratedown