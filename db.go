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
		log.Error(err)
	}
	err = os.Chmod("flow.db", 0644)
	if err != nil {
		log.Error(err)
	}
}

func initialMigration() {
	db, err := gorm.Open("sqlite3", "flow.db")
	if err != nil {
		fmt.Println(err.Error())
		panic("failed to connect database")
	}

	db.AutoMigrate(&Flow{}, &Target{}, &Task{}, &dep{}, FlowRun{}, &TaskRun{})
	db = db.Set("gorm:auto_preload", true)
}

func pollData() {
	for {
		var rs []Target
		err := db.Find(&rs).Error
		if err != nil {
			log.Error(err)
		}

		for _, t := range rs {
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
		time.Sleep(10 * time.Second)
	}
}
