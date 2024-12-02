build:
	docker-compose build
run:
	docker-compose up -d

test:
	go test -v ./...

cover:
	go test -v -cover ./...