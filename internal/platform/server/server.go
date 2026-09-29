package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Run(router *gin.Engine, debug bool) error {
	if debug {
		log.Println("HTTP Server is running on :8080")
		if err := router.Run(":8080"); err != nil {
			return fmt.Errorf("http server failed to start: %w", err)
		}
		return nil
	}

	log.Println("HTTPS Server is running on :8443")
	gin.SetMode(gin.ReleaseMode)
	if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", router); err != nil {
		return fmt.Errorf("https server failed to start: %w", err)
	}
	return nil
}
