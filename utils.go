package main

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func onNewServer(addr string) error {

	url := "http://" + addr + "/cmd/" + "jupyter notebook --ip='*' --NotebookApp.token='' --NotebookApp.password='' --allow-root"
	log.Println(url)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error(err)
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Remote resp not 200")
	}

	log.Info("New notebook server OK")
	return nil
}

func onCMD(input string) error {
	log.Println(input)
	cmd := exec.Command("sh", "-c", input)
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error(err)
	}
}
