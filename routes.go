package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lightning-poll/lnd"
	"lightning-poll/polls"
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
	id, err := getInt(c, "id")
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

func getInt(c *gin.Context, field string) (int64, error) {
	str, ok := c.Params.Get(field)
	if !ok {
		return 0, errors.New("Field name not set")
	}

	return strconv.ParseInt(str, 10, 64)
}

func (e *Env) createPollPost(c *gin.Context) {
	question, ok := c.Params.Get("question")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	payReq, ok := c.Params.Get("invoice")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	sats, err := getInt(c, "satohis")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	expiry, err := getInt(c, "expiry")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}
	expirySeconds := expiry * 60 * 60 // hours to seconds

	id, err := polls.CreatePoll(context.Background(), e, question, payReq, polls.RepaySchemeMajority, expirySeconds, sats, 0)
	// TODO(carla): handle post
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Create Poll",
		},
	)
	c.Re
	e.viewPollPage(c)
}
