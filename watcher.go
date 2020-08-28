package main

import (
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func tunnelWatcher() {
	//delete all tunnels
	var ts []Tunnel
	db.Delete(&ts)

	//set up core + jupyter tunnel
	go watchCoreTunnel()
	go Forward()

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
		ServerAddr: h.IP + ":32768",
		LocalAddr:  localAddr,
		RemoteAddr: RemoteAddr,
	}
	db.Save(&t)
}

//TODO: check if remote running
//TODO: add status update check heartbeat for remote
func Forward() {
	for {
		time.Sleep(time.Second)

		var t Tunnel
		if db.First(&t, "forwarded = ?", false).RecordNotFound() {
			continue
		}

		t.Forward()

		db.Model(&t).Update("forwarded", true)

		log.WithFields(logrus.Fields{
			"LocalAddr": t.LocalAddr,
		}).Info("Start forwarding")

	}
}
