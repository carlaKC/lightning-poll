package main

import (
	"context"
	"database/sql"
	"lightning-poll/polls"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lightning-poll/lnd"
)

type Env struct {
	db  *sql.DB
	lnd lnd.Client
}

func (e Env) GetDB() *sql.DB {
	return e.db
}

func (e Env) GetLND() lnd.Client {
	return e.lnd
}

func initializeRoutes(e Env) {
	router.GET("/", e.showHomePage)
	router.GET("/create", e.createPollPage)
	router.GET("/view/:id", e.viewPollPage)

	router.POST("/create", e.createPollPost)
}

func (e *Env) showHomePage(c *gin.Context) {
	polls, err := polls.ListActivePolls(context.Background(), e)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"home.html",
		gin.H{
			"title": "Lightning Poll - Home",
			"polls": polls,
		},
	)

}

func (e *Env) createPollPage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"create.html",
		gin.H{
			"title": "Lightning Poll - Create",
		},
	)
}

func (e *Env) viewPollPage(c *gin.Context) {
	idStr, ok := c.Params.Get("id")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	poll, err := polls.LookupPoll(context.Background(), e, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Lightning Poll -View Poll",
			"poll":  poll,
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
