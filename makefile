.PHONY: build run
build:
	go build -o ./bin/backend .
run: build
	./bin/backend -config ./config.ini
