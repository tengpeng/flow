package main

import (
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
	Name       string `gorm:"unique;not null" json:"Name"`
	User       string `gorm:"not null" json:"User"`
	IP         string `gorm:"not null" json:"IP"`
	Password   string `json:"Password"` //TODO: Add check constraint
	Pem        string `json:"Pem"`
	ServerAddr string `gorm:"not null"`
	LocalAddr  string `gorm:"not null"`
	RemoteAddr string `gorm:"not null"`
	RemoteHome string
	Forwarded  bool

	client *ssh.Client
	config *ssh.ClientConfig
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

func (t *Target) connect() {
	config := t.newConfig()
	var err error
	t.client, err = t.dial(config)
	if err != nil {
		log.Error(err)
	}
}

func (t *Target) Forward() {
	t.connect()
	go t.forward()
}

func (t *Target) deployBinary() {
	fileName := "flow"
	srcPath := filepath.Join(".", fileName)
	destPath := filepath.Join(t.RemoteHome, fileName)

	t.isJupyterOK()
	//TODO: stop run
	t.runCommand("rm "+destPath, false)
	//TODO: progress bar
	t.copyFile(srcPath, destPath)
	//TODO: tests
	go t.runCommand(destPath, true)
	//TODO: check if running
}

func (t *Target) isJupyterOK() {
	_, err := t.runCommand("jupyter --version", false)
	if err != nil {
		log.Error("Jupyter is not installed")
		return
	}
	log.Info("Jupyter OK")
}

func (t *Target) getHome() {
	homeDir, err := t.runCommand("eval echo ~$USER", false)
	if err != nil {
		log.Error(err)
	}
	t.RemoteHome = filepath.Join(homeDir)
	log.Info("Remote home dir: ", homeDir)
}

func (t *Target) forward() {
	localListener, err := net.Listen("tcp", t.LocalAddr)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}

		go t.copy(localConn)
	}
}

//TODO
func (t *Target) copy(localConn net.Conn) {
	sshConn, err := t.client.Dial("tcp", t.RemoteAddr)
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

func (t *Target) createDstFile(dstPath string) *sftp.File {
	sftp, err := sftp.NewClient(t.client)
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
func (t *Target) copyFile(srcPath string, dstPath string) {
	dstFile := t.createDstFile(dstPath)

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

func (t *Target) runCommand(cmd string, start bool) (string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		log.Error(err)
	}
	defer session.Close()

	b, err := session.CombinedOutput(cmd)
	if err != nil {
		// log.Error(err)
		log.WithFields(logrus.Fields{
			"cmd": cmd,
			"err": err,
		}).Error("Run command failed")
		return "", err
	}

	log.WithFields(logrus.Fields{
		"cmd": cmd,
		"out": strings.TrimSuffix(string(b), "\n"), //TODO
	}).Info("Run command OK")

	return strings.TrimSuffix(string(b), "\n"), nil

	// var b bytes.Buffer
	// session.Stdout = &b

	// if !start {
	// 	err = session.Run(cmd)
	// } else {
	// 	err = session.Start(cmd)
	// }
	// if err != nil {
	// log.WithFields(logrus.Fields{
	// 	"cmd": cmd,
	// 	"out": session.Stderr, //TODO
	// }).Info("Run command failed")

	// 	return "", err
	// }

	// log.WithFields(logrus.Fields{
	// 	"cmd": cmd,
	// 	"out": b.String(),
	// }).Info("Run command OK")

	// return strings.TrimSuffix(b.String(), "\n"), err
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
