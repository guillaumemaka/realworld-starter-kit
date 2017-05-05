.PHONY: build run
build:
	go build -o ./bin/backend .
run: build
	./bin/backend -host "" -port 8080 -dburl "root:password@/conduit?parseTime=true"