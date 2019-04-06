package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"lightning-poll/lnd"
)

type Env struct {
	db  *sql.DB
	lnd lnd.Client
}

func initializeRoutes(e Env) {
	router.GET("/", e.showHomePage)
	router.GET("/create", e.createPollPage)
	router.GET("/view", e.viewPollPage)

	router.POST("/create", e.createPollPost)
}

func (e *Env) showHomePage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"home.html",
		gin.H{
			"title": "Home Page",
		},
	)

}

func (e *Env) createPollPage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"create.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}

func (e *Env) viewPollPage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}

func (e *Env) createPollPost(c *gin.Context) {
	// TODO(carla): handle post
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}
