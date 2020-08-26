package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
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
	err = os.Chmod("flow.db", 0644)
	if err != nil {
		log.Println(err)
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

	db.AutoMigrate(&Flow{}, &Target{}, &Remote{}, &Task{}, &dep{}, FlowRun{}, &TaskRun{})
}

// func setUpTestDB() {
// 	t := &Target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}

// 	t1 := Task{FlowName: "wf1", Name: "nb1", Path: "data/nb1.ipynb", Next: "nb3"}
// 	t2 := Task{FlowName: "wf1", Name: "nb2", Path: "data/nb2.ipynb", Next: "nb3"}
// 	t3 := Task{FlowName: "wf1", Name: "nb3", Path: "data/nb3.ipynb"}

// 	f := &Flow{FlowName: "wf1", Schedule: "* * * * *", Tasks: []Task{t1, t2, t3}}

// 	db.Create(f)
// 	db.Create(t)
// }

//different entry point
func pollData() {
	for {
		var rs []Remote
		err := db.Find(&rs).Error
		if err != nil {
			log.Error(err)
		}

		for _, t := range rs {
			time.Sleep(5 * time.Second)

			url := "http://" + "127.0.0.1:8000" + "/sync"
			resp, err := http.Get(url)
			if err != nil {
				log.Error(err)
				continue
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
				continue
			}

			//TODO: handle empty return
			var frs []FlowRun
			err = json.Unmarshal(b, &frs)
			if err != nil {
				log.Error(err)
				continue
			}

			if len(frs) > 0 {
				for _, v := range frs {
					log.Println(v)
					err := db.Create(&v).Error
					if err != nil {
						log.Error(err)
					}
				}

				log.WithFields(logrus.Fields{
					"remote": t.Name,
					"count":  len(frs),
				}).Info("Poll data OK")
			}
		}
	}
}
