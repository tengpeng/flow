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

	r.POST("/flows", newFlow)
	r.GET("/flows", getFlows) //TODO:

	r.GET("/runs", getRuns)

	r.GET("/sync", sync)

	return r
}

func sync(c *gin.Context) {
	c.Header("Connection", "keep-alive")
	var frs []FlowRun
	db.Find(&frs, "polled = ?", false)
	db.Model(frs).Update("polled", true)
	if len(frs) > 0 {
		c.JSON(http.StatusOK, frs)
	} else {
		//TODO:
		return
	}

}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func newTarget(c *gin.Context) {
	var t Target
	err := c.BindJSON(&t)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = t.isSSHOK()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.ServerAddr = t.IP + ":22"
	t.LocalAddr = "0.0.0.0:8000"
	t.RemoteAddr = "0.0.0.0:9000"

	err = db.Create(&t).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// r := newRemote(t)
	// r.deployBinary()

	c.JSON(200, gin.H{
		"message": "New target created",
	})
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
