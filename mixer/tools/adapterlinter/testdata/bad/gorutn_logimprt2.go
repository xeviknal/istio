package testdata

import (
	"fmt"
	"log")

func f1() {
	//var v int
}

func M1(){
	log.Println("")
	go f()
	fmt.Println("M")
}