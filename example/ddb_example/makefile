deploy: build
	sls deploy

build: triggered write

triggered: triggered/main.go
	go build -o triggered/main triggered/main.go

write: write/main.go
	go build -o write/main write/main.go

.PHONY: all build triggered write
