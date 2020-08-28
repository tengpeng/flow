package main

import (
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//TODO: remove json tag ?
type Flow struct {
	gorm.Model
	FlowName string `gorm:"unique;not null" json:"FlowName"`
	HostID   uint   //TODO
	Schedule string `gorm:"not null" json:"Schedule"`
	Status   string //used by db
	Tasks    []Task //`gorm:"ForeignKey:FlowID"`
}

type Task struct {
	gorm.Model
	FlowID   uint
	FlowName string `gorm:"not null" json:"FlowName"`
	Name     string `gorm:"not null" json:"Name"`
	Path     string `gorm:"not null" json:"Path"`
	Content  string
	Next     string `json:"Next"`
}

type dep struct {
	gorm.Model
	FlowRunID uint
	FlowName  string `gorm:"not null"`
	Parent    string `gorm:"not null"`
	Child     string `gorm:"not null"`
}

type FlowRun struct {
	gorm.Model
	FlowID   uint
	HostID   uint
	FlowName string
	Time     time.Time
	Status   int
	TaskRuns []TaskRun `gorm:"ForeignKey:FlowRunID"` //TODO: why null
}

type TaskRun struct {
	FlowRunID uint
	Name      string
	Path      string
	Status    int
	runCnt    int
	Notebook  string
}

const (
	READY = iota
	RUNNING
	OK
	FAIL
)

func (f *Flow) run() {
	done := make(chan struct{})

	//create flow run
	db.Create(&FlowRun{FlowID: f.ID, HostID: f.HostID, FlowName: f.FlowName, Time: time.Now(), Status: READY})
	log.Info("Flow run created")

	//get flow run
	var r FlowRun
	db.First(&r, "status = ?", READY)

	//get tasks for flow run
	var tasks []Task
	db.Find(&tasks, "flow_id = ?", f.ID)
	r.setTasks(tasks) //Move this

	//start
	go r.start()
	go r.status(done)

	//exist
	<-done
}

// task -> taskrun
func (r *FlowRun) setTasks(tasks []Task) {
	for _, t := range tasks {
		tr := TaskRun{FlowRunID: r.ID, Name: t.Name, Path: t.Path, runCnt: 2, Status: READY}
		db.Create(&tr)

		r.TaskRuns = append(r.TaskRuns, tr)

		//generate dep
		if len(t.Next) > 0 {
			db.Create(&dep{FlowRunID: r.ID, FlowName: t.FlowName, Parent: t.Name, Child: t.Next})
		}

		err := db.Save(&r).Error
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Flow tasks set")
}

func (r *FlowRun) start() {
	db.Model(r).Update("Status", RUNNING)
	log.Info("Flow running")

	for i := range r.TaskRuns {
		r.TaskRuns[i].start()
	}
}

func (r *FlowRun) status(done chan struct{}) {
out:
	for {
		for i := range r.TaskRuns {
			if r.TaskRuns[i].Status == FAIL {
				db.Model(r).Update("Status", FAIL)
				log.Info("Flow run failed")
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
	for _, t := range r.TaskRuns {
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

func (t TaskRun) checkParent() bool {
	var deps []dep
	db.Find(&deps, "flow_run_id = ? and child = ?", t.FlowRunID, t.Name)
	if len(deps) == 0 {
		return true
	}
	return false
}

func (t *TaskRun) delParent() {
	var deps []dep
	db.Find(&deps, "flow_run_id = ? and parent = ?", t.FlowRunID, t.Name)
	db.Delete(&deps)
}

//TODO: refactor
func (t *TaskRun) run() {
	if t.runCnt == 0 {
		t.Status = FAIL
		db.Model(t).Update("status", FAIL)
		log.WithFields(logrus.Fields{
			"task": t.Name,
		}).Info("Task run failed")
		return
	}

	t.runCnt--
	t.Status = RUNNING
	db.Model(t).Update("status", RUNNING)

	out := "temp" + "-" + t.Name
	oPath := out + ".ipynb"
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

	//"status", OK,
	err = db.Model(t).Update("notebook", string(notebook)).Error
	if err != nil {
		log.Error(err)
	}
	//TODO: remove temp notebook
	t.delParent()
}

//TODO: add all existing flows to cron when restarted
func watchNewFlow() {
	for {
		var f Flow
		db.Find(&f, "status = ?", "")

		if f.Schedule != "" {
			db.Model(&f).Update("Status", "STARTED")

			log.WithFields(logrus.Fields{
				"flow": f.FlowName,
			}).Info("Get new flow")

			c := cron.New()
			c.Start()
			c.AddFunc(f.Schedule, func() { f.run() })

			log.WithFields(logrus.Fields{
				"schedule": f.Schedule,
			}).Info("Add cron job")
		}
		time.Sleep(time.Second)
	}
}
