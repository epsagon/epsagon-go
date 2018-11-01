

deploy: build
	sls deploy

build: triggered hello

triggered: triggered/main.go
	go build -o triggered/main triggered/main.go

hello: hello/main.go
	go build -o hello/main hello/main.go

.PHONY: all build triggered hello
