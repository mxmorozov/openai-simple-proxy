package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var limit int
var host string
var port string

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/rewrite", action(chatgptRewrite))
	router.POST("/continue", action(chatgptContinue))
	router.POST("/shorten", action(chatgptShorten))

	return router
}

func action(actionFunc func(text string) (string, error)) func(c *gin.Context) {

	return func(c *gin.Context) {
		var json struct {
			Text string `json:"text" binding:"required"`
		}

		if c.BindJSON(&json) == nil {
			if !probeLimitPerDay() {
				c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
				return
			}
			content, err := actionFunc(strings.TrimSpace(json.Text))
			if err != nil {
				log.Println(err)
				c.String(http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))
			} else {
				c.String(http.StatusOK, content)
			}
		}
	}
}

func main() {
	flag.IntVar(&limit, "limit", 1000, "request limit per day")
	flag.StringVar(&host, "host", "localhost", "service host")
	flag.StringVar(&port, "port", "8080", "service port")
	flag.Parse()

	initDailyLimit()
	router := setupRouter()
	router.Run(host + ":" + port)
}
