NAME=hooknode

all: ${NAME}

install-deps:
	test -x "$(shell which go)"
	test -x "$(shell which dep)" || go get -u github.com/golang/dep/cmd/dep
	test -x "$(shell which gin)" || go get github.com/codegangsta/gin
	dep ensure -update

${NAME}:
	test -x "$(shell which dep)" || ${MAKE} install-deps
	dep ensure && \
	CGO_LDFLAGS_ALLOW='.*' go build -o ./bin/${NAME}
	@echo "Build Success"
	@printf 'sha256: '
	@sha256sum ${NAME}

start: ${NAME}
	./bin/${NAME} 

start-dev:
	gin -i run main.go

.PHONY: ${NAME}

