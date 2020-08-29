package main

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func tunnelWatcher() {
	//delete all tunnels
	var ts []Tunnel
	db.Delete(&ts)

	//set up core + jupyter tunnel
	go watchCoreTunnel()

	forward()
}

//TODO: add all existing flows to cron when restarted
func flowWatcher() {
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

func watchCoreTunnel() {
	for {
		time.Sleep(time.Second)

		var h Host //TODO: deployed & running
		if db.First(&h).RecordNotFound() {
			continue
		}

		newTunnel(h, true)
		break
	}
}

//TODO: called by server to add Jupyter. Core fixed. Jupyter flexible
func newTunnel(h Host, core bool) {
	var localAddr string
	var RemoteAddr string
	if core {
		localAddr = "127.0.0.1:8000"
		RemoteAddr = "127.0.0.1:9000"

	} else {
		localAddr = "127.0.0.1:" + getFreePort()
		RemoteAddr = "127.0.0.1:8888"
	}

	t := Tunnel{
		HostID:     h.ID,
		ServerAddr: h.IP + ":" + p,
		LocalAddr:  localAddr,
		RemoteAddr: RemoteAddr,
	}
	db.Save(&t)
}

//TODO: check if remote running
//TODO: add status update check heartbeat for remote
//TODO: it does not pick up tunnel
func forward() {
	for {
		time.Sleep(time.Second)
		var t Tunnel
		if db.First(&t, "forwarded = ?", false).RecordNotFound() {
			continue
		}
		go t.forward()

		if t.LocalAddr != "127.0.0.1:8000" {
			openBrowser(t.LocalAddr)

			log.WithFields(logrus.Fields{
				"IP": t.LocalAddr,
			}).Info("Open browser")
		}

		db.Model(&t).Update("forwarded", true)

		log.WithFields(logrus.Fields{
			"LocalAddr": t.LocalAddr,
		}).Info("Start forwarding")

	}
}
