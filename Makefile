install-deps:
	go get -u github.com/golang/dep/cmd/dep \
	&& go get github.com/codegangsta/gin \
	&& dep ensure

start:
	go build -o ./bin/main . && ./bin/main

start-dev:
	gin -i run main.go
