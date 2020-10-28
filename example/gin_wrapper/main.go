package main

import (
	"fmt"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/gin"
	"github.com/gin-gonic/gin"
)

func main() {
	r := epsagongin.GinRouterWrapper{
		Engine:   gin.New(),
		Hostname: "my_site",
		Config: epsagon.NewTracerConfig(
			"erez-test-gin", "",
		),
	}

	r.GET("/ping", func(c *gin.Context) {
		time.Sleep(time.Second * 1)
		fmt.Println("hello world")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
