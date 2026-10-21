DB_URL="postgres://sekisq3l:3110_step@localhost:5432/confetki?sslmode=disable"

run:
	CONN_STRING=$(DB_URL) go run cmd/server/main.go

test:
	go test ./...

build:
	go build -o app cmd/server/main.go