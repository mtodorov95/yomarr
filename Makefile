APP_NAME=yomarr
PORT=8080

build:
	docker build -t $(APP_NAME) .

run:
	docker run -d -p $(PORT):$(PORT) --name $(APP_NAME) -v $(PWD)/temp:/data $(APP_NAME) 
stop:
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

rebuild: stop build run

clean:
	docker system prune -f
