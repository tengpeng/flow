package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func server() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:5000"},
		// AllowMethods:     []string{"PUT", "PATCH"},
		// AllowHeaders:     []string{"Origin"},
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "https://github.com"
		// },
		// MaxAge: 12 * time.Hour,
	}))
	//all
	r.GET("/ping", ping)

	//local only
	r.POST("/hosts", newHost)
	r.GET("/hosts", allHosts)
	r.POST("/hosts/:ip", newDeployment)
	r.POST("/notebooks/:ip", newNotebook)

	//core (= centralized db)
	r.POST("/flows", newFlow)
	r.GET("/flows", getFlows)
	r.GET("/runs", getRuns)

	return r
}

func allHosts(c *gin.Context) {
	var hs []Host
	db.Find(&hs)
	c.JSON(200, hs)
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
	ip := c.Param("ip")

	var h Host
	if db.First(&h, "ip = ?", ip).RecordNotFound() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ip + " not found"})
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
		"message": "Installation OK",
	})
}

func newNotebook(c *gin.Context) {
	ip := c.Param("ip")

	var h Host
	if db.First(&h, "ip = ?", ip).RecordNotFound() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ip + " not found"})
		return
	}

	newTunnel(h, false)

	c.JSON(200, gin.H{
		"message": "Open new notebook OK",
	})
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
