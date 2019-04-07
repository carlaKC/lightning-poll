package main

import (
	"context"
	"database/sql"
	"fmt"
	"lightning-poll/votes"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	router.GET("/results/:id", e.viewPollResults)
	router.GET("/vote/:id", e.viewVotePage)

	router.POST("/create", e.createPollPost)
	router.POST("/vote", e.createVotePost)
	router.POST("/close", e.forceClosePoll)
}

func (e *Env) showHomePage(c *gin.Context) {
	open, err := polls.ListActivePolls(c.Request.Context(), e)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	inactive, err := polls.ListInactivePolls(c.Request.Context(), e)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"home.html",
		gin.H{
			"title":  "Lightning Poll - Home",
			"open":   open,
			"closed": inactive,
		},
	)

}

func (e *Env) createPollPage(c *gin.Context) {

	c.HTML(
		http.StatusOK,
		"create.html",
		gin.H{
			"title":     "Lightning Poll - Create",
			"repayment": types.GetRepaySchemes(),
		},
	)
}

func (e *Env) viewPollPage(c *gin.Context) {
	id := getInt(c, "id")

	poll, err := polls.LookupPoll(context.Background(), e, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title":   "Lightning Poll - View Poll",
			"poll":    poll,
			"is_open": time.Now().Before(poll.ClosesAt),
			"unix":    int64(poll.ClosesAt.Unix()),
		},
	)
}

func (e *Env) viewVotePage(c *gin.Context) {
	id := getInt(c, "id")

	vote, err := votes.Lookup(c.Request.Context(), e, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	poll, err := polls.LookupPoll(c.Request.Context(), e, vote.PollID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"vote.html",
		gin.H{
			"title": "Lightning Poll - View Vote",
			"poll":  poll,
			"vote":  vote,
		},
	)
}

type result struct {
	Value string
	Count int64
}

func (e *Env) viewPollResults(c *gin.Context) {
	pollID := getInt(c, "id")

	poll, err := polls.LookupPoll(c.Request.Context(), e, pollID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	results, err := votes.GetResults(c.Request.Context(), e, pollID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	var xScale []string
	var yScale []int64
	for _, r := range poll.Options {
		voteCount, ok := results[r.ID]
		if !ok {
			continue
		}

		xScale = append(xScale, r.Value)
		yScale = append(yScale, voteCount)
	}

	c.HTML(
		http.StatusOK,
		"results.html",
		gin.H{
			"title":  "Lightning Poll - View Poll Results",
			"poll":   poll,
			"xScale": xScale,
			"yScale": yScale,
		},
	)
}

func getInt(c *gin.Context, field string) int64 {
	str, ok := c.Params.Get(field)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	return num
}

func getPostInt(c *gin.Context, field string) int64 {
	num, err := strconv.ParseInt(c.PostForm(field), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	return num
}

func (e *Env) createPollPost(c *gin.Context) {
	question := c.PostForm("question")
	payReq := c.PostForm("invoice")
	sats := getPostInt(c, "satoshis")

	expiry := getPostInt(c, "expiry")
	expirySeconds := expiry * 60 * 60 // hours to seconds

	options := c.PostForm("added[]")

	getPostInt(c, "payout")

	id, err := polls.CreatePoll(context.Background(), e, question, payReq,
		getPostInt(c, "payout"), strings.Split(options, ","), expirySeconds, sats, 0)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	// TODO(carla): figure out non hacky redirect
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%v", id)})
	e.viewPollPage(c)
}

func (e *Env) createVotePost(c *gin.Context) {
	pollID := getPostInt(c, "poll_id")
	optionID := getPostInt(c, "id")

	poll, err := polls.LookupPoll(c.Request.Context(), e, pollID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	expirySeconds := time.Now().Sub(poll.ClosesAt).Seconds()
	note := fmt.Sprintf("Vote: %v for poll: %v", c.PostForm("opt_str"), c.PostForm("poll_str"))
	id, err := votes.Create(c.Request.Context(), e, pollID, optionID, poll.Cost, int64(expirySeconds), note)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	//TODO(carla): stop being hacky
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%v", id)})
	e.viewVotePage(c)
}

func (e *Env) forceClosePoll(c *gin.Context) {
	pollID := getPostInt(c, "id")

	if err := polls.FoceClosePoll(c.Request.Context(), e, pollID); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	//TODO(carla): stop being hacky
	c.Params = append(c.Params, gin.Param{Key: "id", Value: fmt.Sprintf("%v", pollID)})
	e.viewPollResults(c)
}
