package main

import (
	"flag"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var limitPerDay int
var host string
var port string
var allowedIPs []string

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.POST("/rewrite", action(chatgptRewrite))
	router.POST("/continue", action(chatgptContinue))
	router.POST("/shorten", action(chatgptShorten))

	return router
}

func action(actionFunc func(text string) (string, error)) func(c *gin.Context) {

	return func(c *gin.Context) {

		if allowedIPs != nil && !slices.Contains(allowedIPs, c.ClientIP()) {
			c.String(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		if !probeLimitPerDay() {
			c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
			return
		}

		var json struct {
			Text string `json:"text" binding:"required"`
		}

		if c.BindJSON(&json) == nil {
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
	flag.IntVar(&limitPerDay, "limit", 1000, "request limit per day")
	flag.StringVar(&host, "host", "localhost", "service host")
	flag.StringVar(&port, "port", "8080", "service port")
	allowedIPsFlag := flag.String("allowed-ips", "", "allowed IPs")
	flag.Parse()

	if *allowedIPsFlag != "" {
		allowedIPs = strings.Split(*allowedIPsFlag, ",")
	}

	for i := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
	}

	initDailyLimit()
	router := setupRouter()
	router.Run(host + ":" + port)
}
