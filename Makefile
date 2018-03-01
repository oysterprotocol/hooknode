export:
        export GOROOT=/usr/local/go \ 
	export GOPATH=/home/ubuntu/go \ 
	export PATH=$GOPATH/bin:$GOROOT/bin:$PATH/

install-deps:
	go get -u github.com/golang/dep/cmd/dep \
	&& go get github.com/codegangsta/gin \
	&& dep ensure

build:
	export GIT_COMMIT=$(git rev-list -1 HEAD) \
	&& dep ensure \
	&& go build -o ./bin/main .

start: build
	./bin/main

start-dev:
	gin -i run main.go
