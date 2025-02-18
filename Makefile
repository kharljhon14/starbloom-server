migrateup:
	migrate -path migrations -database postgresql://root:postgres@localhost:5432/starbloom?sslmode=disable up

migratedown:
	migrate -path migrations -database postgresql://root:postgres@localhost:5432/starbloom?sslmode=disable down

server:
	go run main.go


.PHONY: migrateup migratedown server