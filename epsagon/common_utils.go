package epsagon

import (
	"context"
	"fmt"

	"github.com/epsagon/epsagon-go/internal"
	"github.com/epsagon/epsagon-go/tracer"
)

// Config is the configuration for Epsagon's tracer
type Config struct {
	tracer.Config
}

// GeneralEpsagonRecover recover function that will send exception to epsagon
// exceptionType, msg are strings that will be added to the exception
func GeneralEpsagonRecover(exceptionType, msg string, currentTracer tracer.Tracer) {
	if r := recover(); r != nil && currentTracer != nil {
		currentTracer.AddExceptionTypeAndMessage(exceptionType, fmt.Sprintf("%s:%+v", msg, r))
	}
}

// NewTracerConfig creates a new tracer Config
func NewTracerConfig(applicationName, token string) *Config {
	return &Config{
		Config: tracer.Config{
			ApplicationName: applicationName,
			Token:           token,
			MetadataOnly:    true,
			Debug:           false,
			SendTimeout:     "1s",
		},
	}
}

// Label adds a label to the sent trace
func Label(key string, value interface{}, args ...context.Context) {
	currentTracer := internal.ExtractTracer(args)
	if currentTracer != nil {
		currentTracer.AddLabel(key, value)
	}
}
