package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

/*
How to send new flow? local database export to JSON -> Remote
How to pick up new new flow ? watch JSON file ->
*/

// var (
// 	username         = "root"
// 	password         = "z"
// 	serverAddrString = "0.0.0.0:32772"
// 	localAddrString  = "localhost:8000"
// 	remoteAddrString = "localhost:9000"
// )

func main() {
	log.Info("Bayesnote flow started")

	// os.Remove("flow.db")
	initDB()

	// setUpTestDB()

	//testFlow()
	//cronTrigger()

	//deploy()

	r := server()
	r.Run(":9000")
	//testForward()

}

func testFlow() {
	var f flow
	db.First(&f, 1)
	f.generateDep()

	f.run()
}

//user -> db ->
func cronTrigger() {
	c := cron.New()
	c.AddFunc("* * * * *", func() { fmt.Println("Every minute") })
	c.AddFunc("* * * * *", func() { startFlowRun() })
	c.Start()
	time.Sleep(5 * time.Minute)
}

func startFlowRun() {
	var f flow
	db.First(&f, 1)
	f.generateDep()

	f.run()
}

func deploy() {
	t := target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
	r := newRemote(t)
	r.deployBinary()
}

func exportFlow() {
	var f flow
	db.First(&f, 1)
	b, err := json.Marshal(f)
	if err != nil {
		log.Error(err)
	}
	ioutil.WriteFile("temp.json", b, 0666)
}

func testForward() {
	t := target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
	r := newRemote(t)
	r.localAddr = "localhost:8000"
	r.remoteAddr = "0.0.0.0:9000"
	r.serverAddr = "0.0.0.0:32772"
	r.forward()
}
