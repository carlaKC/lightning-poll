package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func initializeRoutes() {
	router.GET("/", showHomePage)
	router.GET("/create", createPollPage)
	router.GET("/view", viewPollPage)

	router.POST("/create", createPollPost)
}

func showHomePage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"home.html",
		gin.H{
			"title": "Home Page",
		},
	)

}

func createPollPage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"create.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}

func viewPollPage(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}

func createPollPost(c *gin.Context) {
	// TODO(carla): handle post
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"title": "Create Poll",
		},
	)
}
