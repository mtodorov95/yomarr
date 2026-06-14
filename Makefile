APP_NAME=yomarr
PORT=9191

build:
	docker build -t $(APP_NAME) .

run:
	docker run -d -p $(PORT):$(PORT) --name $(APP_NAME) \
		--user $(shell id -u):$(shell id -g) \
		-v $(PWD)/temp:/app/temp \
		-v /home/mario/Hestia/Manga:/Manga \
		-v /home/mario/Hestia/Downloads:/downloads \
		--env-file .env.prod \
		$(APP_NAME)

stop:
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

rebuild: stop build run

clean:
	docker system prune -f
