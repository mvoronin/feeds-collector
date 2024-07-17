IMAGE_NAME = feeds-collector-frontend-app-image
CONTAINER_NAME = feeds-collector-frontend-app-container

.PHONY: all build up down clean

all: build up

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down -v
	docker system prune -f
	docker rmi $(IMAGE_NAME)
