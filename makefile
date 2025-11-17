.PHONY: postgres run stop

POSTGRES_CONTAINER = my-postgres
POSTGRES_PASSWORD = password
POSTGRES_USER = lev-demchenko
POSTGRES_DB = postgres
POSTGRES_PORT = 5432

postgres:
	@docker run --rm -d \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		--name $(POSTGRES_CONTAINER) \
		postgres:14
	@echo "Waiting for Postgres to be ready..."
	@sleep 5 # Можно тут заменить на pg_isready с docker exec, для продвинутой проверки

run: postgres app
	@echo "Running Go app..."
	@go run ./cmd/main.go

app: 
	@go run ./cmd/main.go

stop:
	@docker stop $(POSTGRES_CONTAINER)

