package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func startFlowRun() {
	var f flow
	db.First(&f, 1)
	f.generateDep()
	f.run()
}

func deploy() {
	t := target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
	r := newRemote(t)
	r.deployBinary()
}

func exportFlow() {
	var f flow
	db.First(&f, 1)
	b, err := json.Marshal(f)
	if err != nil {
		log.Error(err)
	}
	ioutil.WriteFile("temp.json", b, 0666)
}

func testForward() {
	t := target{Name: "aws", User: "ubuntu", Pem: "/Users/pt/Downloads/hk.pem", IP: "ec2-18-166-71-228.ap-east-1.compute.amazonaws.com"}
	r := newRemote(t)

	r.serverAddr = "ec2-18-166-71-228.ap-east-1.compute.amazonaws.com:22"
	r.localAddr = "localhost:8000"
	r.remoteAddr = "localhost:9000"

	log.Info("forwarding")
	r.forward()
}

func localTestForward() {
	t := target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
	r := newRemote(t)

	r.localAddr = "0.0.0.0:8000"
	r.remoteAddr = "0.0.0.0:9000"
	r.serverAddr = "0.0.0.0:22"

	log.Info("forwarding")

	r.forward()
}
