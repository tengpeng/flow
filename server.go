package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func server() *gin.Engine {
	r := gin.Default()

	//TODO:
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	r.GET("/ping", ping)
	r.POST("/targets", newTarget)
	r.POST("/targets/:name", newDeployment)
	r.POST("/flows", newFlow)
	r.GET("/flows", getFlows) //TODO:
	r.GET("/runs", getRuns)

	return r
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func newTarget(c *gin.Context) {
	//write to db
	var t target
	err := c.BindJSON(&t)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = db.Create(&t).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//new target
	// r := newRemote(t)

	//TODO: ping remote
}

func newDeployment(c *gin.Context) {
	name := c.Param("name")

	var t target
	err := db.Find(&t, "name = ?", name).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r := newRemote(t)
	r.deployBinary()
}

//TODO: or update
func newFlow(c *gin.Context) {
	var f Flow
	err := c.BindJSON(&f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = db.Create(&f).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func getFlows(c *gin.Context) {
	var f Flow
	db.First(&f, 1)
	c.JSON(http.StatusOK, f)
}

func getRuns(c *gin.Context) {
	var runs []FlowRun
	db.Find(&runs)
	c.JSON(http.StatusOK, runs)
}
