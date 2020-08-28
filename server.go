package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func server() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	//TODO:
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	//all
	r.GET("/ping", ping)
	//TODO: upload/download files

	//local only
	r.POST("/hosts", newHost)
	r.POST("/hosts/:name", newDeployment)

	//core (= centralized db)
	r.POST("/flows", newFlow) //TODO: how to handle notebook path
	r.GET("/flows", getFlows) //TODO:
	r.POST("/tasks")          //TODO:
	r.GET("/runs", getRuns)

	return r
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func newHost(c *gin.Context) {
	var t Host
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

	c.JSON(200, gin.H{
		"message": "New target created",
	})
}

func newDeployment(c *gin.Context) {
	name := c.Param("name")
	var h Host
	db.First(&h, "name = ?", name)
	if h.IP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " not found"})
		return
	}

	h.connect()
	err := h.deployBinary()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&h)

	//Set core
	var hs []Host
	db.Find(&hs)

	if len(hs) == 1 {
		var t Tunnel
		t.HostID = h.ID
		t.ServerAddr = h.IP + ":32768"
		t.LocalAddr = "127.0.0.1:8000"  //turn this on when env=prod getFreePort()
		t.RemoteAddr = "127.0.0.1:9000" //Get freeport -> write to env var -> read remote env
		db.Save(&t)
	}

	//jupyter
	var t Tunnel
	t.HostID = h.ID
	t.ServerAddr = h.IP + ":32768"
	t.LocalAddr = "127.0.0.1:8001"  //turn this on when env=prod getFreePort()
	t.RemoteAddr = "127.0.0.1:8888" //Get freeport -> write to env var -> read remote env
	db.Save(&t)

	c.JSON(200, gin.H{
		"message": "New deployment OK",
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

//TODO:
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
