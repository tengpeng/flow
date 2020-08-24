package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type flow struct {
	gorm.Model
	FlowName string `gorm:"unique;not null"`
	Target   string `gorm:"not null"`
	Schedule string `gorm:"not null"`
}

type task struct {
	gorm.Model
	FlowName string `gorm:"not null"`
	Name     string `gorm:"not null"`
	Path     string `gorm:"not null"`
	Next     string
}

type dep struct {
	gorm.Model
	FlowName   string `gorm:"not null"`
	Upstream   string `gorm:"not null"`
	Downstream string `gorm:"not null"`
}

type flowRun struct {
	gorm.Model
	RunID    string `gorm:"AUTO_INCREMENT"`
	FlowName string
	Time     time.Time
	Status   int

	tasks []taskRun
}

type taskRun struct {
	RunID    string
	Name     string
	Path     string
	Next     string
	Status   int
	retryCnt int
}

const (
	READY = iota
	RUNNING
	OK
	FAIL
)

type vertex struct {
	gorm.Model
	Name       string
	Status     string
	RetryCnt   int
	Upstream   []string
	Downstream []string
}

func (f flow) generateDep() {
	var tasks []task
	db.Find(&tasks, "flow_name = ?", f.FlowName)

	for _, t := range tasks {
		if len(t.Next) > 0 {
			db.Create(&dep{FlowName: t.FlowName, Upstream: t.Name, Downstream: t.Next})
		}
	}
}

//cron trigger run: flow -> flow run
func (f *flow) run() {
	done := make(chan struct{})

	db.Create(&flowRun{FlowName: f.FlowName, Time: time.Now(), Status: READY})
	log.Info("Flow run created")

	var r flowRun
	db.First(&r, "status = ?", READY)

	r.setTasks()
	r.start()
	go r.status(done)

	<-done
}

// task -> taskrun
func (r *flowRun) setTasks() {
	var tasks []task
	db.Find(&tasks, "flow_name = ?", r.FlowName)

	for _, t := range tasks {
		tr := taskRun{RunID: r.RunID, Name: t.Name, Path: t.Path, Next: t.Next, retryCnt: 2, Status: READY}
		r.tasks = append(r.tasks, tr)
		db.Create(&tr)
	}

	log.Info("Flow tasks set")
}

func (r *flowRun) start() {
	db.Model(r).Update("Status", RUNNING)
	for i := range r.tasks {
		r.tasks[i].start()
	}

	log.Info("Flow run started")
}

func (r *flowRun) status(done chan struct{}) {
out:
	for {
		for _, t := range r.tasks {
			if t.Status == FAIL {
				db.Model(r).Update("Status", FAIL)
				log.Info("Flow run Failed")
				break out
			}
		}

		if r.done() {
			db.Model(r).Update("Status", OK)
			log.Info("Flow run OK")
			break
		}

		time.Sleep(time.Second)
	}

	log.Info("Flow run exist")

	done <- struct{}{}
}

func (r flowRun) done() bool {
	for _, t := range r.tasks {
		if t.Status == READY || t.Status == RUNNING {
			return false
		}
	}
	return true
}

func (t *taskRun) start() {
	t.Status = READY
	db.Model(t).Update("Status", READY)

	for {
		if t.checkUpstream() {
			t.run()
			break
		}
		time.Sleep(time.Second)
	}
}

func (t taskRun) checkUpstream() bool {
	var deps []dep
	db.Find(&deps, "downstream = ?", t.Name)

	if len(deps) == 0 {
		return true
	}
	return false
}

func (t *taskRun) delUpstream() {
	var deps []dep
	db.Find(&deps, "upstream = ?", t.Name)
	db.Delete(&deps)
}

func (t *taskRun) run() {
	if t.retryCnt == 0 {
		t.Status = FAIL
		db.Model(t).Update("status", FAIL)
		return
	}

	t.retryCnt--
	t.Status = RUNNING
	db.Model(t).Update("status", RUNNING)

	cmd := exec.Command("pwd")
	fmt.Println(t.Name + " pwd")
	err := cmd.Run()
	if err != nil {
		t.run()
		return
	}

	t.Status = OK
	db.Model(t).Update("status", OK)
	t.delUpstream()
}
