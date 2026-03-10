run:
	go run ./cmd/server

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	migrate -path migrations \
	-database "mysql://root:root@tcp(localhost:3306)/task_service" up

migrate-down:
	migrate -path migrations \
	-database "mysql://root:root@tcp(localhost:3306)/task_service" down