package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func server() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	//all
	r.GET("/ping", ping)

	//local only
	r.POST("/hosts", newHost)
	r.POST("/hosts/:name", newDeployment)
	r.POST("/notebooks/:name", newNotebook)

	//core (= centralized db)
	r.POST("/flows", newFlow)
	r.GET("/flows", getFlows)
	r.GET("/runs", getRuns)

	return r
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func newHost(c *gin.Context) {
	var h Host
	err := c.BindJSON(&h)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.connect()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = db.Create(&h).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "New host created",
	})
}

func newDeployment(c *gin.Context) {
	name := c.Param("name")

	var h Host
	if db.First(&h, "name = ?", name).RecordNotFound() {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " not found"})
		return
	}

	_, err := h.connect()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.deployBinary()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "New deployment OK",
	})
}

func newNotebook(c *gin.Context) {
	name := c.Param("name")

	var h Host
	if db.First(&h, "name = ?", name).RecordNotFound() {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " not found"})
		return
	}

	newTunnel(h, false)

	c.JSON(200, nil)
}

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
	var fs []Flow
	db.Find(&fs)
	c.JSON(http.StatusOK, fs)
}

func getRuns(c *gin.Context) {
	var runs []FlowRun
	db.Find(&runs)
	c.JSON(http.StatusOK, runs)
}
