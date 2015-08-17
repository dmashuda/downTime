package main

import (
	"fmt"
	"net"
	"time"
)

type NetworkResponse struct {
	Value    string
	HasError bool
	Error    error
	Time     time.Duration
}

var router_ip = "192.168.0.1:80"
var host_list = []string{
	"google.com:80",
	"facebook.com:80",
	"cloudflare.com:80"}

func main() {
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
			c <- NetworkResponse{HasError: true, Error: err, Time: time.Since(startTime)}
		} else {
			c <- NetworkResponse{HasError: false, Time: time.Since(startTime)}
			conn.Close()
		}
		time.Sleep(time.Second * 10)
	}
}

func responseHandler(c chan NetworkResponse) {
	for {
		resp := <-c
		if resp.HasError {
			fmt.Println(resp)
		}
	}
}
