package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mxk/go-sqlite/sqlite3"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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
	EndTime   time.Time     `sql:"-"`
	MonthInt  int           `sql:"-"`
}
type Configuration struct {
	Port      string
	Gateway   string
	TestHosts []string
	DbName    string
}

var defaults = Configuration{
	Port:    ":8080",
	Gateway: "192.168.0.1:80",
	TestHosts: []string{
		"google.com:80",
		"facebook.com:80",
		"cloudflare.com:80"},
	DbName: "./failures.db",
}

var configFile = flag.String("json", "", "Config file")
var config Configuration
var chartTemplate *template.Template

var dbInstance gorm.DB

func main() {

	flag.Parse()

	if len(*configFile) > 0 {
		f, err := os.Open(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = json.NewDecoder(f).Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		config = defaults
	}

	fmt.Println(config)

	dbSetup()
	templateSetup()
	channel := make(chan NetworkResponse)
	go netTestHandler(channel)

	for z := 0; z < len(config.TestHosts); z++ {
		go dialTester(config.TestHosts[z], false, channel)
	}
	go dialTester(config.Gateway, true, channel)

	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(config.Port, nil)

	var i int
	_, err := fmt.Scanf("%d", &i)
	println(err)

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	db := dbInstance

	results := []NetworkResponse{}
	db.Where(&NetworkResponse{}).Find(&results)

	for i := 0; i < len(results); i++ {
		results[i].Time = 10 * time.Second
		results[i].EndTime = results[i].CreatedAt.Add(results[i].Time)
		results[i].MonthInt = int(results[i].CreatedAt.Month()) - 1

	}
	chartTemplate.Execute(w, results)
	//fmt.Fprintf(w, string(jsonVal))
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

func templateSetup() {
	dat, err := ioutil.ReadFile("chart.go.html")

	if err != nil {
		fmt.Println(err)
	}

	temp := string(dat)

	t := template.New("Chart")
	t, err = t.Parse(temp)

	if err != nil {
		fmt.Println(err)
	}
	chartTemplate = t
}

func dbSetup() {
	db, err := gorm.Open("sqlite3", config.DbName)
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
