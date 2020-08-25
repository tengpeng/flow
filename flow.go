package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Flow struct {
	gorm.Model
	FlowName string `gorm:"unique;not null" json:"FlowName"`
	Target   string `gorm:"not null" json:"Target"`
	Schedule string `gorm:"not null" json:"Schedule"`
	Status   string //used by db
	Tasks    []Task
}

type Task struct {
	gorm.Model
	FlowID   uint
	FlowName string `gorm:"not null" json:"FlowName"` //TODO: use flowID instead
	Name     string `gorm:"not null" json:"Name"`
	Path     string `gorm:"not null" json:"Path"`
	Next     string `json:"Next"`
}

type dep struct {
	gorm.Model
	FlowName string `gorm:"not null"`
	Parent   string `gorm:"not null"`
	Child    string `gorm:"not null"`
}

type FlowRun struct {
	gorm.Model
	FlowID   uint
	FlowName string
	Time     time.Time
	Status   int
	tasks    []TaskRun
}

type TaskRun struct {
	FlowRunID uint
	Name      string
	Path      string
	Status    int
	retryCnt  int
	Notebook  string
}

const (
	READY = iota
	RUNNING
	OK
	FAIL
)

//TODO: refactor
func (f *Flow) start() {
	f.generateDep() //TODO: not necessary if no change in Flow
	f.run()
}

func (f Flow) generateDep() {
	var tasks []Task
	db.Find(&tasks, "flow_name = ?", f.FlowName)

	for _, t := range tasks {
		if len(t.Next) > 0 {
			db.Create(&dep{FlowName: t.FlowName, Parent: t.Name, Child: t.Next})
		}
	}
}

//cron trigger run: flow -> flow run
func (f *Flow) run() {
	done := make(chan struct{})

	db.Create(&FlowRun{FlowID: f.ID, FlowName: f.FlowName, Time: time.Now(), Status: READY})
	log.Info("Flow run created")

	var r FlowRun
	db.First(&r, "status = ?", READY)

	var tasks []Task
	db.Find(&tasks, "flow_id = ?", f.ID)
	r.setTasks(tasks) //Move this

	r.start()
	go r.status(done)

	<-done
}

// task -> taskrun
func (r *FlowRun) setTasks(tasks []Task) {
	for _, t := range tasks {
		tr := TaskRun{FlowRunID: r.ID, Name: t.Name, Path: t.Path, retryCnt: 2, Status: READY}
		r.tasks = append(r.tasks, tr)
		db.Create(&tr)
	}

	log.Info("Flow tasks set")
}

func (r *FlowRun) start() {
	db.Model(r).Update("Status", RUNNING)
	log.Info("Flow running")

	for i := range r.tasks {
		r.tasks[i].start()
	}
}

//TODO: tests
func (r *FlowRun) status(done chan struct{}) {
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

func (r FlowRun) done() bool {
	for _, t := range r.tasks {
		if t.Status == READY || t.Status == RUNNING {
			return false
		}
	}
	return true
}

func (t *TaskRun) start() {
	for {
		if t.checkParent() {
			t.run()
			break
		}
		time.Sleep(time.Second)
	}
}

//TODO: ID
func (t TaskRun) checkParent() bool {
	var deps []dep
	db.Find(&deps, "child = ?", t.Name)
	if len(deps) == 0 {
		return true
	}
	return false
}

func (t *TaskRun) delParent() {
	var deps []dep
	db.Find(&deps, "parent = ?", t.Name)
	db.Delete(&deps)
}

//TODO: refactor
//TODO: check if jupyter installed
func (t *TaskRun) run() {
	if t.retryCnt == 0 {
		t.Status = FAIL
		db.Model(t).Update("status", FAIL)
		return
	}

	t.retryCnt--
	t.Status = RUNNING
	db.Model(t).Update("status", RUNNING) //TODO: failed ?

	out := "temp" + "-" + t.Name
	oPath := filepath.Join("data", out+".ipynb")
	cmd := exec.Command("x", "nbconvert", "--to", "notebook",
		"--output", out, "--execute", t.Path, "--ExecutePreprocessor.timeout=3600")
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		t.run()
		return
	}

	t.Status = OK
	log.WithFields(logrus.Fields{
		"task": t.Name,
	}).Info("Task run OK")

	notebook, err := ioutil.ReadFile(oPath)
	if err != nil {
		log.Error(err)
	}

	db.Model(t).Update("status", OK, "notebook", string(notebook))
	t.delParent()
}
