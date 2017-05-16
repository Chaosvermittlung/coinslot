package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

type donation struct {
	Name    string
	Amount  float64
	Message string
}

type project struct {
	ProjectID   int
	Name        string
	Explanation string
	Goal        float64
	Amount      float64
	Percentage  float64
	Difference  float64
	Diffpercent float64
	Donations   []donation
}

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func (p *project) calcPercentage() float64 {
	return p.Amount / p.Goal * 100
}

func (p *project) calcDifference() float64 {
	return p.Goal - p.Amount
}

func (p *project) calcAmount() float64 {
	var res float64
	for _, d := range p.Donations {
		res = res + d.Amount
	}
	return Round(res, 0.5, 2)
}

type mainpage struct {
	Projects []project
}

type config struct {
	Port int
	Con  dbConnection
}

func (c *config) Load() error {
	f, err := os.Open(filepath.FromSlash("config.json"))
	if err != nil {
		return err
	}
	d := json.NewDecoder(f)
	err = d.Decode(c)
	return err
}

var conf config

func mainhandler(w http.ResponseWriter, r *http.Request) {
	var mp mainpage

	pp, err := GetProjects()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range pp {
		p.Amount = p.calcAmount()
		p.Percentage = p.calcPercentage()
		p.Difference = p.calcDifference()
		p.Diffpercent = 100 - p.Percentage
		mp.Projects = append(mp.Projects, p)
	}

	t, err := template.ParseFiles("templates/main.html")
	if err != nil {
		fmt.Fprintln(w, "Error parsing template:", err)
		return
	}
	err = t.Execute(w, mp)
	if err != nil {
		fmt.Fprintln(w, "Error executing template:", err)
		return
	}

}

func main() {
	err := conf.Load()
	if err != nil {
		log.Fatal("Error loadin config:", err)
	}
	initialisation(conf.Con)
	r := mux.NewRouter()
	r.HandleFunc("/", mainhandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(conf.Port), r))
}
