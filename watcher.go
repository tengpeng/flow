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

		newTunnel(h, "127.0.0.1:9000")
		break
	}
}

//TODO: called by server to add Jupyter
func newTunnel(h Host, RemoteAddr string) {
	t := Tunnel{
		HostID:     h.ID,
		ServerAddr: h.IP + ":32768",
		LocalAddr:  "127.0.0.1:8000", //TODO: get free port
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
			"RemoteAddr": t.RemoteAddr,
		}).Info("Start forwarding")

	}
}
