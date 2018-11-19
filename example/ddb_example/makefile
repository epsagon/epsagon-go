deploy: build
	sls deploy

build: trigger/main write/main

local: trigger/local_main write/local_main

trigger/main: trigger/main.go
    GOOS=linux go build -o trigger/main trigger/main.go

write/main: write/main.go
    GOOS=linux go build -o write/main write/main.go

trigger/local_main: trigger/main.go
    GOOS=darwin go build -o trigger/local_main trigger/main.go

write/local_main: write/main.go
    GOOS=darwin go build -o write/local_main write/main.go

.PHONY: all build trigger write
