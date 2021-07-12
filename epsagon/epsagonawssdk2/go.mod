module github.com/epsagon/epsagon-go/epsagon/epsagonawssdk2

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v0.23.0 // indirect
	github.com/epsagon/epsagon-go/epsagon latest
	github.com/epsagon/epsagon-go/protocol latest
	github.com/epsagon/epsagon-go/tracer latest
)

replace (
	github.com/epsagon/epsagon-go/epsagon latest => ../.
	github.com/epsagon/epsagon-go/protocol latest => ../../protocol
	github.com/epsagon/epsagon-go/tracer latest => ../../tracer

)