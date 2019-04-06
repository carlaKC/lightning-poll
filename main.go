package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"lightning-poll/db"
)

var router *gin.Engine

func main() {
	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("/Users/carla/personal/src/lightning-poll/templates/*")

	dbc, err := db.Connect()
	if err != nil {
		log.Fatalf("could not connect to DB: %v", err)
	}

	// Initialize the routes
	initializeRoutes(Env{db: dbc})

	// Start serving the application
	router.Run()

}
