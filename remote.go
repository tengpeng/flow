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

	"github.com/cheggaaa/pb"
	"github.com/jinzhu/gorm"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Host struct {
	gorm.Model
	Name     string `gorm:"unique;not null" json:"Name"`
	User     string `gorm:"not null" json:"User"`
	IP       string `gorm:"not null" json:"IP"`
	Password string `json:"Password"` //TODO: Add check constraint
	Pem      string `json:"Pem"`
	Tunnels  []Tunnel

	client *ssh.Client
}

type Tunnel struct {
	gorm.Model
	HostID     uint
	LocalAddr  string `gorm:unique;not null"`
	ServerAddr string `gorm:"not null"`
	RemoteAddr string `gorm:"not null"`
	// Deployed   bool
	Forwarded bool
}

func (t *Host) connect() (*ssh.Client, error) {
	err := t.checkPort("32768")
	if err != nil {
		return nil, err
	}

	config := t.newConfig()
	client, err := t.dial(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	t.client = client

	return client, nil
}

//TODO:
func (t *Tunnel) Forward() {
	var h Host
	if db.First(&h, "id = ?", t.HostID).RecordNotFound() {
		log.Error("Host not found")
		return
	}

	client, err := h.connect()
	if err != nil {
		log.Error(err)
	}

	go t.forward(client)
}

func (t *Host) deployBinary() error {
	fileName := "flow"
	remoteHome := t.getRemoteHome()
	srcPath := filepath.Join(".", fileName)
	destPath := filepath.Join(remoteHome, fileName)

	//TODO:
	// err := t.isJupyterOK()
	// if err != nil {
	// 	return err
	// }

	t.runCommand("rm "+destPath, false)
	t.copyFile(srcPath, destPath)
	go t.runCommand(destPath+" -worker", true)
	//TODO: check if running
	db.Model(t).Update("deployed = ?", true)

	return nil
}

//TODO:
func (t *Host) isJupyterOK() error {
	_, err := t.runCommand("shell -c jupyter", false)
	if err != nil {
		log.Error("Jupyter is not installed")
		return err
	}
	log.Info("Jupyter OK")
	return nil
}

func (t *Host) getRemoteHome() string {
	homeDir, err := t.runCommand("eval echo ~$USER", false)
	if err != nil {
		log.Error(err)
	}
	remoteHome := filepath.Join(homeDir)
	log.Info("Remote home dir: ", homeDir)
	return remoteHome
}

// func (t *Target) getRemotePort() {
// 	p, err := t.runCommand("env | grep FLOW_PORT", false)
// 	if err != nil {
// 		log.Error(err)
// 	}

// 	t.remotePort = p
// 	log.WithFields(logrus.Fields{
// 		"port": p,
// 	}).Info("Get remote work port")
// }

func (t *Tunnel) forward(sshClient *ssh.Client) {
	localListener, err := net.Listen("tcp", t.LocalAddr)
	if err != nil {
		log.Error("net.Listen failed: %v", err)
	}

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Error("listen.Accept failed: %v", err)
		}

		go t.copy(sshClient, localConn)
	}
}

//TODO: ssh: rejected: connect failed (Connection refused)
//TODO: ssh: unexpected packet in response to channel open: <nil>
//TODO: ERRO[0181] io.Copy failed: %vreadfrom tcp 127.0.0.1:8001->127.0.0.1:53145: write tcp 127.0.0.1:8001->127.0.0.1:53145: write: broken pipe
func (t *Tunnel) copy(sshClient *ssh.Client, localConn net.Conn) {
	sshConn, err := sshClient.Dial("tcp", t.RemoteAddr)
	if err != nil {
		log.Error(err)
	}
	time.Sleep(1 * time.Second)

	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Error("io.Copy failed: %v", err)
		}
	}()

	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Error("io.Copy failed: %v", err)
		}
	}()
}

func (t *Host) createDstFile(dstPath string) *sftp.File {
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
func (t *Host) copyFile(srcPath string, dstPath string) {
	dstFile := t.createDstFile(dstPath)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Error(err, srcPath)
	}

	//TODO:
	go func() {
		srcInfo, err := srcFile.Stat()
		if err != nil {
			log.Error(err)
		}

		count := int(srcInfo.Size())
		bar := pb.StartNew(count)
		for {
			if int(bar.Get()) < count {
				dstInfo, err := dstFile.Stat()
				if err != nil {
					log.Error(err)
				}

				bar.Set64(dstInfo.Size())
			} else {
				bar.Finish()
				break
			}
		}
	}()

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

func (t *Host) runCommand(cmd string, start bool) (string, error) {
	session, err := t.client.NewSession()
	if err != nil {
		log.Error(err)
	}
	defer session.Close()

	b, err := session.CombinedOutput(cmd)
	if err != nil {
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
}

func (t Host) newClientPemConfig() *ssh.ClientConfig {
	err := os.Chmod(t.Pem, 0400)
	if err != nil {
		log.Error(err)
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

func (t Host) newClientPasswordConfig() *ssh.ClientConfig {
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

func (t Host) newConfig() *ssh.ClientConfig {
	var config *ssh.ClientConfig
	if len(t.Pem) > 0 {
		config = t.newClientPemConfig()
	} else {
		config = t.newClientPasswordConfig()
	}
	return config
}

func (t Host) dial(config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", t.IP+":"+"32768", config)
	if err != nil {
		log.Error(err, t.IP, config.Config)
		return nil, errors.New("SSH failed")
	} else {
		log.Info("SSH connected: ", t.IP)
	}
	return client, nil
}

func (t Host) checkPort(port string) error {
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
	return errors.New("Port 32768 is not OK")
}

func getFreePort() string {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	l, _ := net.ListenTCP("tcp", addr)
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}
