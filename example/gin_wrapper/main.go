package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagongin "github.com/epsagon/epsagon-go/wrappers/gin"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// r := gin.Default()
	r := epsagongin.GinRouterWrapper{
		IRouter:  gin.Default(),
		Hostname: "my_site",
		Config: epsagon.NewTracerConfig(
			"erez-test-gin", "",
		),
	}

	r.POST("/ping", func(c *gin.Context) {
		time.Sleep(time.Second * 1)
		body, err := ioutil.ReadAll(c.Request.Body)
		if err == nil {
			fmt.Println(body)
		} else {
			fmt.Println("Error reading body: ", err)
		}
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// r.Run()
	r.IRouter.(*gin.Engine).Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
