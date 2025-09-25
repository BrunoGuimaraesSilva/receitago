# Variables
COMPOSE = docker-compose
DB_DSN = postgres://receitago:receitago@localhost:5432/receitago?sslmode=disable

# Run app + DBs
up:
	$(COMPOSE) up --build

# Stop containers
down:
	$(COMPOSE) down

# Run migrations
migrate-up:
	$(COMPOSE) run --rm migrate

migrate-down:
	$(COMPOSE) run --rm migrate down 1

# Create a new migration
migration:
	@read -p "Enter migration name: " name; \
	docker run --rm -v $$PWD/db/migrations:/migrations migrate/migrate \
	create -ext sql -dir /migrations $$name
