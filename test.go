package main

func startFlowRun() {
	var f Flow
	db.First(&f, 1)
	f.run()
}

// func deploy() {
// 	t := Target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
// 	r := newRemote(t)
// 	r.deployBinary()
// }

// func exportFlow() {
// 	var f Flow
// 	db.First(&f, 1)
// 	b, err := json.Marshal(f)
// 	if err != nil {
// 		log.Error(err)
// 	}
// 	ioutil.WriteFile("temp.json", b, 0666)
// }

// func testForward() {
// 	t := Target{Name: "aws", User: "ubuntu", Pem: "/Users/pt/Downloads/hk.pem", IP: "ec2-18-166-71-228.ap-east-1.compute.amazonaws.com"}
// 	r := newRemote(t)

// 	r.ServerAddr = "ec2-18-166-71-228.ap-east-1.compute.amazonaws.com:22"
// 	r.LocalAddr = "localhost:8000"
// 	r.RemoteAddr = "localhost:9000"

// 	log.Info("forwarding")
// 	r.forward()
// }

// func localTestForward() {
// 	t := Target{Name: "test", User: "root", Password: "z", IP: "0.0.0.0"}
// 	r := newRemote(t)

// 	r.LocalAddr = "0.0.0.0:8000"
// 	r.RemoteAddr = "0.0.0.0:9000"
// 	r.ServerAddr = "0.0.0.0:22"

// 	log.Info("forwarding")

// 	r.forward()
// }
