package main

import (
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func initDB() {
	openDB()
	initialMigration()
}

func openDB() {
	var err error
	db, err = gorm.Open("sqlite3", "flow.db")
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	//defer db.Close()

	db = db.Set("gorm:auto_preload", true)
}

func initialMigration() {
	db, err := gorm.Open("sqlite3", "flow.db")
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&Flow{})
	db.AutoMigrate(&target{})
	db.AutoMigrate(&remote{})
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&dep{})
	db.AutoMigrate(&FlowRun{})
	db.AutoMigrate(&TaskRun{})
}

func setUpTestDB() {
	t := &target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}

	t1 := Task{FlowName: "wf1", Name: "nb1", Path: "data/nb1.ipynb", Next: "nb3"}
	t2 := Task{FlowName: "wf1", Name: "nb2", Path: "data/nb2.ipynb", Next: "nb3"}
	t3 := Task{FlowName: "wf1", Name: "nb3", Path: "data/nb3.ipynb"}

	f := &Flow{FlowName: "wf1", Target: "test", Schedule: "* * * * *", Tasks: []Task{t1, t2, t3}}

	db.Create(f)
	db.Create(t)
}

func getData() {
	resp, err := http.Get("")
	if err != nil {
		log.Error(err)
	}

	if resp.StatusCode != 200 {
		log.Error("x 200")
	}

	resp, err = http.Post("", "", resp.Body)
	if err != nil {
		log.Error(err)
	}
}
