package main

import (
	"flag"
	"log"

	"github.com/carlaKC/lightning-poll/db"
	"github.com/carlaKC/lightning-poll/lnd"
	"github.com/carlaKC/lightning-poll/polls"
	"github.com/carlaKC/lightning-poll/votes"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

var baseTemplates = flag.String("templates_base",
	"/Users/carla/personal/src/github.com/carlaKC", "location of templates")

func main() {
	flag.Parse()

	// Set the router as the default one provided by Gin
	router = gin.Default()

	router.LoadHTMLGlob(*baseTemplates + "/lightning-poll/templates/*")

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
