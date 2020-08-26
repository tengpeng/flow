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

type Target struct {
	gorm.Model
	Name     string `gorm:"unique;not null" json:"Name"`
	User     string `gorm:"not null" json:"User"`
	IP       string `gorm:"not null" json:"IP"`
	Password string `json:"Password"` //TODO: Add check constraint
	Pem      string `json:"Pem"`
}

type Remote struct {
	gorm.Model
	TargetID   uint
	Name       string `gorm:"unique;not null"`
	ServerAddr string `gorm:"not null"`
	LocalAddr  string `gorm:"not null"`
	RemoteAddr string `gorm:"not null"`

	remoteHome string
	client     *ssh.Client
	config     *ssh.ClientConfig
}

//TODO: SSH error
func (t Target) isSSHOK() error {
	err := t.checkPort("32768")
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
func newRemote(t Target) Remote {
	config := t.newConfig()
	client, err := t.dial(config)
	if err != nil {
		log.Error(err)
	}

	r := Remote{TargetID: t.ID, Name: t.Name, client: client, config: config}

	r.getHome()

	r.ServerAddr = t.IP + ":22"
	r.LocalAddr = "0.0.0.0:8000"  //TODO: getFreePort()
	r.RemoteAddr = "0.0.0.0:9000" //TODO: getFreePort()

	go r.forward()
	//	db.Create(&r)
	return r
}

//TODO: websocket

func (r Remote) deployBinary() {
	fileName := "flow"
	srcPath := filepath.Join(".", fileName)
	destPath := filepath.Join(r.remoteHome, fileName)

	r.isJupyterOK()
	//TODO: stop run
	r.runCommand("rm "+destPath, false)
	//TODO: progress bar
	r.copyFile(srcPath, destPath)
	//TODO: tests
	r.runCommand(destPath, true)
	//TODO: check if running
}

func (r Remote) isJupyterOK() {
	_, err := r.runCommand("jupyter --version", false)
	if err != nil {
		log.Error("Jupyter is not installed")
		return
	}
	log.Info("Jupyter OK")
}

func (r *Remote) getHome() {
	homeDir, err := r.runCommand("eval echo ~$USER", false)
	if err != nil {
		log.Error(err)
	}
	r.remoteHome = filepath.Join(homeDir)
	log.Info("Remote home dir: ", homeDir)
}

func (r *Remote) forward() {
	localListener, err := net.Listen("tcp", r.LocalAddr)
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

func (r *Remote) connect(localConn net.Conn) {
	sshConn, err := r.client.Dial("tcp", r.RemoteAddr)
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

func (r Remote) createDstFile(dstPath string) *sftp.File {
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
func (r Remote) copyFile(srcPath string, dstPath string) {
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

func (r Remote) runCommand(cmd string, start bool) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		log.Error(err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if !start {
		err = session.Run(cmd)
	} else {
		err = session.Start(cmd)
	}
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

func (t Target) newClientPemConfig() *ssh.ClientConfig {
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

func (t Target) newClientPasswordConfig() *ssh.ClientConfig {
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

func (t Target) newConfig() *ssh.ClientConfig {
	var config *ssh.ClientConfig
	if len(t.Pem) > 0 {
		config = t.newClientPemConfig()
	} else {
		config = t.newClientPasswordConfig()
	}
	return config
}

func (t Target) dial(config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", t.IP+":"+"32768", config)
	if err != nil {
		log.Error(err, t.IP, config.Config)
		return nil, errors.New("SSH failed")
	} else {
		log.Info("SSH connected: ", t.IP)
	}
	return client, nil
}

func (t Target) checkPort(port string) error {
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
