package main

import (
	"os/exec"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Cmd struct {
	gorm.Model
	TargetID string
	Input    string
	Output   string
	Wait     bool //TODO: add to server
	Status   int
}

//TODO: error handling
func watchCmd() {
	for {
		time.Sleep(time.Second)

		var c Cmd
		if db.First(&c, "status = ?", 0).RecordNotFound() {
			continue
		}

		log.WithFields(logrus.Fields{
			"cmd": c.Input,
		}).Info("Start command")

		cmd := exec.Command("sh", "-c", c.Input)
		err := cmd.Start()
		if err != nil {
			c.Status = 3
			db.Save(&c)
			log.Error(err)
			continue
		}
		c.Status = 2 //TODO: this failed for long running command
		db.Save(&c)
		//TODO: write to core using Put

	}
}
