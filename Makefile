export APP_VERSION := $(shell git describe --tags --always || echo "v0.0.0-dev")

build:
	docker compose build

run:
	docker compose up -d

stop:
	docker compose down

rebuild: stop build run

clean:
	docker system prune -f
