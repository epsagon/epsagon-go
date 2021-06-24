package main

import (
	"fmt"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagonfiber "github.com/epsagon/epsagon-go/wrappers/fiber"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := epsagon.NewTracerConfig(
		"fiber-example", "",
	)
	config.MetadataOnly = false
	app := fiber.New()
	// Match all routes
	epsagonMiddleware := &epsagonfiber.FiberEpsagonMiddleware{
		Config: config,
	}
	app.Use(epsagonMiddleware.HandlerFunc())
	app.Post("/", func(c *fiber.Ctx) error {
		// any call wrapped by Epsagon should receive the user context from the fiber context
		client := http.Client{Transport: epsagonhttp.NewTracingTransport(c.UserContext())}
		client.Get(fmt.Sprintf("https://epsagon.com/"))
		return c.SendString("Hello, World 👋!")
	})

	app.Listen("0.0.0.0:3000")
}
