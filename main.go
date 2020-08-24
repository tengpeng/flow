package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

func main() {
	log.Info("Bayesnote flow started")

	// os.Remove("flow.db")

	// initDB()
	// setUpTestDB()

	//testFlow()
	//cronTrigger()

	deploy()
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
