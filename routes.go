package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"lightning-poll/lnd"
	"lightning-poll/polls"
	"lightning-poll/types"
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
		c.AbortWithError(http.StatusInternalServerError, err)
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
	question := c.PostForm("question")
	payReq := c.PostForm("invoice")

	sats, err := strconv.ParseInt(c.PostForm("satoshis"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	expiry, err := strconv.ParseInt(c.PostForm("expiry"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	expirySeconds := expiry * 60 * 60 // hours to seconds

	options := c.PostForm("added[]")

	id, err := polls.CreatePoll(context.Background(), e, question, payReq,
		types.RepaySchemeMajority, strings.Split(options, ","), expirySeconds, sats, 0)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	// TODO(carla): figure out non hacky redirect
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%v", id)})
	e.viewPollPage(c)
}
