.PHONY: run fmt lint build

run:
	cd backend && go run . -config ../config.yaml

build:
	docker build -t pastebooks:dev .

fmt:
	cd backend && go fmt ./...

lint:
	cd backend && go vet ./...