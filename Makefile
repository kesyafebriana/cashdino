.PHONY: up down logs seed migrate-up migrate-down reset

up:
	docker-compose up -d --build

down:
	docker-compose down

logs:
	docker-compose logs -f

seed:
	docker-compose exec api /app/seed

migrate-up:
	docker-compose exec api migrate -path /app/migrations -database "$${DATABASE_URL}" up

migrate-down:
	docker-compose exec api migrate -path /app/migrations -database "$${DATABASE_URL}" down 1

reset: down up migrate-up seed
