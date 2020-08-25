package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func server() *gin.Engine {
	r := gin.Default()

	//TODO:
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	r.GET("/ping", ping)
	r.POST("/flows", newFlow)
	r.GET("/runs", getRuns)
	r.POST("/runs", newRuns)

	return r
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

//TODO: or update
func newFlow(c *gin.Context) {
	var f flow
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

func getRuns(c *gin.Context) {
	var runs []flowRun
	db.Find(&runs)
	c.JSON(http.StatusOK, runs)
}

//TODO:
type flowRuns struct {
	Runs []flowRun `json:"runs"`
}

func newRuns(c *gin.Context) {
	var r flowRuns
	err := c.BindJSON(&r)
	if err != nil {
		log.Error(err)
	}

	for _, v := range r.Runs {
		db.Create(&v)
	}
	// db.Find(&runs)
	// c.JSON(200, gin.H{
	// 	"runs": runs,
	// })
}
