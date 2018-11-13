

deploy: build
	sls deploy

build: hello

hello: main.go
	go build -o main main.go

.PHONY: all build hello
