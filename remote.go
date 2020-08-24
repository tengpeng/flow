package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type target struct {
	gorm.Model
	Name     string `gorm:"unique;not null"`
	User     string `gorm:"not null"`
	IP       string `gorm:"not null"`
	Password string
	Pem      string
}

type remote struct {
	client *ssh.Client
	home   string
}

//TODO: error handling
func newRemote(t target) remote {
	t.checkPort("32772")
	config := t.newConfig()
	r := remote{client: t.dial(config)}
	r.getHome()
	return r
}

func (r *remote) getHome() {
	homeDir, err := r.runCommand("eval echo ~$USER")
	if err != nil {
		log.Error(err)
	}
	r.home = filepath.Join(homeDir)
	log.Info("Remote home dir: ", homeDir)
}

//TODO: stop run
func (r remote) deployBinary() {
	fileName := "flow"
	srcPath := filepath.Join(".", fileName)
	destPath := filepath.Join(r.home, fileName)

	r.runCommand("rm " + destPath)
	r.copyFile(srcPath, destPath)
	r.runCommand(destPath + " &")
}

func (r remote) dispatchFlow() {

}

func (r remote) createDstFile(dstPath string) *sftp.File {
	sftp, err := sftp.NewClient(r.client)
	if err != nil {
		log.Error(err)
	}

	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		log.Error(err, dstPath)
	}

	dstFile.Chmod(777)
	if err != nil {
		log.Error(err)
	}

	return dstFile
}

func (r remote) copyFile(srcPath string, dstPath string) {
	dstFile := r.createDstFile(dstPath)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Error(err, srcPath)
	}

	bytes, err := dstFile.ReadFrom(srcFile)
	if err != nil {
		log.Error(err, srcPath, dstPath)
	}

	dstFile.Close()

	log.WithFields(logrus.Fields{
		"src":   srcPath,
		"dst":   dstPath,
		"bytes": bytes,
	}).Info("Copy file ")
}

func (r remote) runCommand(cmd string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		log.Error(err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	err = session.Run(cmd)
	if err != nil {
		log.WithFields(logrus.Fields{
			"cmd": cmd,
			"out": session.Stderr, //TODO
		}).Info("Run command failed")

		return "", err
	}

	log.WithFields(logrus.Fields{
		"cmd": cmd,
		"out": b.String(),
	}).Info("Run command OK")

	return strings.TrimSuffix(b.String(), "\n"), err
}

func (t target) newClientPemConfig() *ssh.ClientConfig {
	err := os.Chmod(t.Pem, 400)
	if err != nil {
		log.Fatal(err)
	}

	pemBytes, err := ioutil.ReadFile(t.Pem)
	if err != nil {
		log.Error(err)
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Error("parse key failed: ", err)
	}

	config := &ssh.ClientConfig{
		User: t.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Info("SSH pem config generated")
	return config
}

func (t target) newClientPasswordConfig() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: t.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(t.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Info("SSH password config generated")
	return config
}

func (t target) newConfig() *ssh.ClientConfig {
	var config *ssh.ClientConfig
	if len(t.Pem) > 0 {
		config = t.newClientPemConfig()
	} else {
		config = t.newClientPasswordConfig()
	}
	return config
}

func (t target) dial(config *ssh.ClientConfig) *ssh.Client {
	client, err := ssh.Dial("tcp", t.IP+":"+"32772", config)
	if err != nil {
		log.Error(err, t.IP, config.Config)
	} else {
		log.Info("SSH connected: ", t.IP)
	}
	return client
}

func (t target) checkPort(port string) {
	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(t.IP, port), time.Second)
		if err != nil {
			log.Error("CheckPorts failed:", err)
			time.Sleep(time.Second)
		} else {
			defer conn.Close()
			log.Info("Opened ", net.JoinHostPort(t.IP, port))
			break
		}
	}
}

//deploy to remote

//dispatch flow

//get flow run + notebook
