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
	config := epsagon.NewTracerConfig(
		"gin-wrapper-test", "",
	)
	config.MetadataOnly = false
	r := epsagongin.GinRouterWrapper{
		IRouter:  gin.Default(),
		Hostname: "my-site",
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
		resp, err := client.Get("http://www.example.com")

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

	/* listen and serve on 0.0.0.0:<PORT> */
	/* for windows - localhost:<PORT> */
	/* Port arg format: ":<PORT>" "*/
	err := r.Run(":3001")

	if err != nil {
		fmt.Println(err)
	}
}
