package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"time"
)

type NetworkResponse struct {
	gorm.Model
	Host     string
	Value    string
	HasError bool
	Error    error `sql:"-"`
	Time     time.Duration
}

var router_ip = "192.168.0.1:80"
var host_list = []string{
	"google.com:80",
	"facebook.com:80",
	"cloudflare.com:80"}

func main() {
	dbSetup()
	channel := make(chan NetworkResponse)
	go responseHandler(channel)

	for z := 0; z < len(host_list); z++ {
		go dialTester(host_list[z], channel)
	}
	go dialTester(router_ip, channel)

	var i int
	_, err := fmt.Scanf("%d", &i)
	println(err)
}

func dialTester(host string, c chan NetworkResponse) {
	for {
		startTime := time.Now()
		conn, err := net.DialTimeout("tcp", host, time.Second*20)
		if err != nil {
			c <- NetworkResponse{HasError: true, Error: err, Time: time.Since(startTime), Host: host}
		} else {
			c <- NetworkResponse{HasError: false, Time: time.Since(startTime), Host: host}
			conn.Close()
		}
		time.Sleep(time.Second * 10)
	}
}

func dbSetup() {
	db, err := gorm.Open("sqlite3", "./foo.db")
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	db.CreateTable(&NetworkResponse{})
}

func responseHandler(c chan NetworkResponse) error {
	db, err := gorm.Open("sqlite3", "./foo.db")
	defer db.Close()

	if err != nil {
		return error(err)
	}

	db.Exec("")

	for {
		resp := <-c
		fmt.Println(resp)
		if resp.HasError {
			db.Create(&resp)
			fmt.Println(resp)
		}
	}

	return nil
}
