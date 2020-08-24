package main

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/ssh"
)

type target struct {
	gorm.Model
	Name     string `gorm:"unique;not null"`
	User     string `gorm:"not null"`
	IP       string `gorm:"not null"`
	Password string
	Pem      string

	client     *ssh.Client
	remotePath string
}
