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
	FlowName string `gorm:"not null"`
	Parent   string `gorm:"not null"`
	Child    string `gorm:"not null"`
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
	Notebook string
}

const (
	READY = iota
	RUNNING
	OK
	FAIL
)

func (f flow) generateDep() {
	var tasks []task
	db.Find(&tasks, "flow_name = ?", f.FlowName)

	for _, t := range tasks {
		if len(t.Next) > 0 {
			db.Create(&dep{FlowName: t.FlowName, Parent: t.Name, Child: t.Next})
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

//TODO: tests
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
		if t.checkParent() {
			t.run()
			break
		}
		time.Sleep(time.Second)
	}
}

func (t taskRun) checkParent() bool {
	var deps []dep
	db.Find(&deps, "child = ?", t.Name)
	if len(deps) == 0 {
		return true
	}
	return false
}

func (t *taskRun) delParent() {
	var deps []dep
	db.Find(&deps, "parent = ?", t.Name)
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

	out := "temp" + "-" + t.Name
	oPath := filepath.Join("data", out+".ipynb")
	cmd := exec.Command("jupyter", "nbconvert", "--to", "notebook",
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
