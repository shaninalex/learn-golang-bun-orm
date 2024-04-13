db:
	docker compose up -d --build

db_down:
	docker compose down -v

start:
	go run main.go
