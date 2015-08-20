package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"net/http"
	"time"
)

type NetworkResponse struct {
	gorm.Model
	Host      string
	IsGateway bool
	Value     string
	HasError  bool
	Error     error         `sql:"-"`
	Time      time.Duration `sql:"-"`
}

var router_ip = "192.168.0.1:80"
var host_list = []string{
	"google.com:80",
	"facebook.com:80",
	"cloudflare.com:80"}

var dbInstance gorm.DB

func main() {
	dbSetup()
	channel := make(chan NetworkResponse)
	go netTestHandler(channel)

	for z := 0; z < len(host_list); z++ {
		go dialTester(host_list[z], false, channel)
	}
	go dialTester(router_ip, true, channel)

	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8080", nil)

	var i int
	_, err := fmt.Scanf("%d", &i)
	println(err)

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	db := dbInstance

	results := []NetworkResponse{}
	db.Where(&NetworkResponse{}).Find(&results)

	jsonVal, jsonerror := json.Marshal(results)
	if jsonerror != nil {
		return
	}
	fmt.Fprintf(w, string(jsonVal))
}

func dialTester(host string, isGateway bool, c chan NetworkResponse) {
	for {
		startTime := time.Now()
		conn, err := net.DialTimeout("tcp", host, time.Second*5)
		if err != nil {
			val := fmt.Sprintf("%s", err)
			c <- NetworkResponse{HasError: true, Error: err, Time: time.Since(startTime), Host: host, IsGateway: isGateway, Value: val}
		} else {
			c <- NetworkResponse{HasError: false, Time: time.Since(startTime), Host: host, IsGateway: isGateway}
			conn.Close()
		}
		time.Sleep(time.Second * 10)
	}
}

func dbSetup() {
	db, err := gorm.Open("sqlite3", "./failures.db")
	if err == nil {
		fmt.Println(err)
	}
	dbInstance = db

	db.CreateTable(&NetworkResponse{})
}

func netTestHandler(c chan NetworkResponse) error {
	db := dbInstance

	for {
		resp := <-c
		if resp.HasError {
			fmt.Println(resp)
			db.Create(&resp)
		}
	}

	return nil
}
