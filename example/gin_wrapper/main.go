package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagongin "github.com/epsagon/epsagon-go/wrappers/gin"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// r := gin.Default()
	config := epsagon.NewTracerConfig(
		"erez-test-gin", "",
	)
	config.MetadataOnly = false
	r := epsagongin.GinRouterWrapper{
		IRouter:  gin.Default(),
		Hostname: "my_site",
		Config:   config,
	}

	r.GET("/ping", func(c *gin.Context) {
		time.Sleep(time.Second * 1)
		fmt.Println("hello world")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/ping", func(c *gin.Context) {
		time.Sleep(time.Second * 1)
		body, err := ioutil.ReadAll(c.Request.Body)
		if err == nil {
			fmt.Println("Recieved body: ", string(body))
		} else {
			fmt.Println("Error reading body: ", err)
		}

		client := http.Client{
			Transport: epsagonhttp.NewTracingTransport(epsagongin.EpsagonContext(c))}
		resp, err := client.Get("http://example.com")

		if err == nil {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				fmt.Println("First 1000 bytes recieved: ", string(respBody[:1000]))
			}
		}

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// r.Run()
	r.IRouter.(*gin.Engine).Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
