package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
	Name     string `gorm:"unique;not null" json:"Name"`
	User     string `gorm:"not null" json:"User"`
	IP       string `gorm:"not null" json:"IP"`
	Password string `json:"Password"` //TODO: Add check constraint
	Pem      string `json:"Pem"`
}

type remote struct {
	gorm.Model
	Name       string `gorm:"unique;not null"`
	serverAddr string `gorm:"not null"`
	localAddr  string `gorm:"not null"`
	remoteAddr string `gorm:"not null"`
	remoteHome string `gorm:"not null"`

	client *ssh.Client
	config *ssh.ClientConfig
}

//TODO: SSH error
func (t target) isSSHOK() error {
	err := t.checkPort("32772")
	if err != nil {
		return err
	}

	config := t.newConfig()
	_, err = t.dial(config)
	if err != nil {
		return err
	}
	return nil
}

//TODO: check jupyter installation
//TODO: error handling
func newRemote(t target) remote {
	config := t.newConfig()
	client, err := t.dial(config)
	if err != nil {
		log.Error(err)
	}

	r := remote{client: client, config: config}

	r.getHome()

	r.serverAddr = t.IP + ":22"
	r.localAddr = "0.0.0.0:8000"  //TODO: getFreePort()
	r.remoteAddr = "0.0.0.0:9000" //TODO: getFreePort()

	go r.forward()
	db.Create(&r)
	return r
}

//TODO: websocket

func (r remote) deployBinary() {
	fileName := "flow"
	srcPath := filepath.Join(".", fileName)
	destPath := filepath.Join(r.remoteHome, fileName)

	//TODO: stop run
	r.runCommand("rm " + destPath)
	//TODO: progress bar
	r.copyFile(srcPath, destPath)
	r.runCommand(destPath + " &")
	//TODO: check if running
}

func (r *remote) getHome() {
	homeDir, err := r.runCommand("eval echo ~$USER")
	if err != nil {
		log.Error(err)
	}
	r.remoteHome = filepath.Join(homeDir)
	log.Info("Remote home dir: ", homeDir)
}

func (r *remote) forward() {
	localListener, err := net.Listen("tcp", r.localAddr)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}

		go r.connect(localConn)
	}
}

func (r *remote) connect(localConn net.Conn) {
	sshConn, err := r.client.Dial("tcp", r.remoteAddr)
	if err != nil {
		log.Error(err)
	}

	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()

	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()
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

	if err != nil {
		log.Error(err)
	}

	return dstFile
}

//TODO: compare size of files
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

	err = dstFile.Chmod(0755)
	if err != nil {
		log.Error(err)
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
	//TODO: Check permission 400

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

func (t target) dial(config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", t.IP+":"+"32772", config)
	if err != nil {
		log.Error(err, t.IP, config.Config)
		return nil, errors.New("SSH failed")
	} else {
		log.Info("SSH connected: ", t.IP)
	}
	return client, nil
}

func (t target) checkPort(port string) error {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(t.IP, port), time.Second)
		if err != nil {
			log.Error("CheckPorts failed:", err)
		} else {
			defer conn.Close()
			log.Info("Opened ", net.JoinHostPort(t.IP, port))
			return nil
		}
	}
	return errors.New("Port 22 is not OK")
}

func getFreePort() string {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	l, _ := net.ListenTCP("tcp", addr)
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}
