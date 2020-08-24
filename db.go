package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
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

	db.AutoMigrate(&flow{})
	db.AutoMigrate(&target{})
	db.AutoMigrate(&task{})
	db.AutoMigrate(&dep{})
	db.AutoMigrate(&flowRun{})
	db.AutoMigrate(&taskRun{})
	// db.AutoMigrate(&vertex{})
}

func setUpTestDB() {
	t := &target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
	t1 := &task{FlowName: "wf1", Name: "nb1", Path: "data/nb1.ipynb", Next: "nb3"}
	t2 := &task{FlowName: "wf1", Name: "nb2", Path: "data/nb2.ipynb", Next: "nb3"}
	t3 := &task{FlowName: "wf1", Name: "nb3", Path: "data/nb3.ipynb"}

	f := &flow{FlowName: "wf1", Target: "test", Schedule: "*/5 * * * *"}

	db.Create(f)
	db.Create(t)
	db.Create(t1)
	db.Create(t2)
	db.Create(t3)
}
