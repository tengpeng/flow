package main

import "github.com/jinzhu/gorm"

type Cmd struct {
	gorm.Model
	TargetID string
	Input    string
	Output   string
	Status   int
}
