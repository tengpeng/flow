package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	r.POST("/cmd/:input", cmd)

	//local only
	r.POST("/targets", newTarget)
	r.POST("/targets/:name", newDeployment)

	//core (= centralized db)
	r.POST("/flows", newFlow) //TODO: how to handle notebook path
	r.GET("/flows", getFlows) //TODO:
	r.POST("/tasks")          //TODO:
	r.POST("/notebooks/:name", newNotebook)
	r.GET("/runs", getRuns)

	return r
}

func newNotebook(c *gin.Context) {
	name := c.Param("name")

	var t Target
	err := db.Find(&t, "name = ?", name).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if t.LocalAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//start jupyter server at 8888
	fmt.Println("xxx: ", t.LocalAddr)
	err = onNewServer(t.LocalAddr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//write db
	t.JupyterAddr = "127.0.0.1:" + getFreePort()

	err = db.Save(&t).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//openbrowser
	openBrowser(t.JupyterAddr)

	c.JSON(200, gin.H{
		"message": "New Notebook OK",
	})
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

	t.getRemoteHome()

	//set core db flag
	var ts []Target
	db.Find(&ts)
	if len(ts) == 0 {
		t.Database = true
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
	var t Target
	db.First(&t, "name = ?", name)
	if t.IP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " not found"})
		return
	}

	t.connect()
	err := t.deployBinary()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.ServerAddr = t.IP + ":32768"
	t.LocalAddr = "127.0.0.1:" + getFreePort() //This can be get from getFreeport
	t.RemoteAddr = "127.0.0.1:9000"            //Get freeport -> write to env var -> read remote env
	t.Deployed = true

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

func cmd(c *gin.Context) {
	input := c.Param("input")
	err := onCMD(input)
	if err != nil {
		log.Error(err)
	}
	c.JSON(200, gin.H{
		"message": "OK",
	})
}
