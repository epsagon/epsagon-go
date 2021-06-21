aws_profile := $(AWS_CREDS)
stage_env := $(STAGE_ENV)
aws_region := $(AWS_REGION)

.PHONY: help build test coverage clean dep deploy undeploy

help:
	grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

functions := $(shell find cmd -name \*main.go | awk -F'/' '{print $$2}')

build:
	go mod download github.com/epsagon/epsagon-go
	go mod download github.com/pkg/errors
	go mod download github.com/aws/aws-lambda-go
	go mod download github.com/aws/aws-sdk-go

	go mod download github.com/aws/aws-sdk-go-v2
	go mod download github.com/golang/protobuf
	go mod download github.com/google/uuid
	go mod download github.com/onsi/gomega
	go mod download google.golang.org/protobuf
	go mod download github.com/jmespath/go-jmespath


	for function in $(functions) ; do \
		env GOOS=linux GOARCH=amd64 go build -v -ldflags="-d -s -w" -a -tags netgo -installsuffix netgo -o bin/$$function cmd/$$function/*.go ; \
	done

show:
	ls
	ls bin

test:
	golint ./...
	go vet ./...
	go test ./... -coverprofile=cover.out

coverage:
	go tool cover -html=cover.out -o coverage.html

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --stage $(stage_env) --region $(aws_region) --verbose

undeploy:
	sls remove --stage $(stage_env) --region $(aws_region) --verbose
    
