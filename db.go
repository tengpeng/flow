package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

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

	db.AutoMigrate(&Flow{})
	db.AutoMigrate(&Target{})
	db.AutoMigrate(&Remote{})
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&dep{})
	db.AutoMigrate(&FlowRun{})
	db.AutoMigrate(&TaskRun{})
}

func setUpTestDB() {
	t := &Target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}

	t1 := Task{FlowName: "wf1", Name: "nb1", Path: "data/nb1.ipynb", Next: "nb3"}
	t2 := Task{FlowName: "wf1", Name: "nb2", Path: "data/nb2.ipynb", Next: "nb3"}
	t3 := Task{FlowName: "wf1", Name: "nb3", Path: "data/nb3.ipynb"}

	f := &Flow{FlowName: "wf1", Schedule: "* * * * *", Tasks: []Task{t1, t2, t3}}

	db.Create(f)
	db.Create(t)
}

//different entry point
func pollData() {
	for {
		var rs []Remote
		err := db.Find(&rs).Error
		if err != nil {
			log.Error(err)
		}

		for _, t := range rs {
			//find local runs
			var flowRuns []FlowRun
			db.Delete(&flowRuns, "target_id = ?", t.TargetID)

			url := "http://127.0.0.1:8000/sync"
			method := "GET"

			client := &http.Client{}
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				fmt.Println(err)
			}

			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
			} else {
				defer res.Body.Close()

				body, _ := ioutil.ReadAll(res.Body)

				fmt.Println(string(body))
			}

			// url := "http://" + t.LocalAddr + "/ping"

			// resp, err := http.Get(url)
			// if err != nil {
			// 	log.Error(err)
			// } else {
			// 	b, err := ioutil.ReadAll(resp.Body)
			// 	if err != nil {
			// 		log.Error(err)
			// 	}

			// 	var frs []FlowRun
			// 	err = json.Unmarshal(b, &frs)
			// 	if err != nil {
			// 		log.Error(err)
			// 	}

			// 	log.WithFields(logrus.Fields{
			// 		"remote": t.Name,
			// 		"count":  len(frs),
			// 	}).Info("Poll data OK")

			// 	db.Create(&frs)
			// }
		}

		time.Sleep(10 * time.Second)
	}
}
