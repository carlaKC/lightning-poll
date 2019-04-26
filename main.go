package main

import (
	"flag"
	"lightning-poll/polls"
	"lightning-poll/votes"
	"log"

	"github.com/gin-gonic/gin"

	"lightning-poll/db"
	"lightning-poll/lnd"
)

var router *gin.Engine

func main() {
	flag.Parse()

	// Set the router as the default one provided by Gin
	router = gin.Default()

	router.LoadHTMLGlob("/Users/carla/personal/src/lightning-poll/templates/*")

	dbc, err := db.Connect()
	if err != nil {
		log.Fatalf("could not connect to DB: %v", err)
	}

	lndCl, err := lnd.New()
	if err != nil {
		log.Fatalf("could not connect to LND: %v", err)
	}
	env := &Env{db: dbc, lnd: lndCl}

	votes.StartLoops(env)
	polls.StartLoops(env)

	// Initialize the routes
	initializeRoutes(env)

	// Start serving the application
	router.Run()

}
