

deploy: build
	sls deploy

build: main

main: main.go
	go build -o main main.go

.PHONY: all build main
