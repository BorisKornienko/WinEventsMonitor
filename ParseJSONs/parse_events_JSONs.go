package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type ParsedEvents struct {
	ApplicationsCritical int `json:"Applications_Critical"`
	SystemError          []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Error"`
	IP                  string `json:"ip"`
	ApplicationsWarning struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Warning"`
	SystemCritical    int    `json:"System_Critical"`
	Computer          string `json:"computer"`
	ApplicationsError []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Error"`
	SystemWarning []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Warning"`
}

func getToStruct(jsonPath string) (ParsedEvents, string, string) {
	var eventFile ParsedEvents
	var fMachineName string
	var fDate string

	// For each dir get path to JSON and unmarshal data to struct
	// also get machine name from directory name and date from JSON name for next compr
	// with struct data

	///////////////////// parse to struct block
	f, err := os.Open(jsonPath)
	if err != nil {
		log.Fatal(err)
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(byteValue, &eventFile)
	if err != nil {
		log.Fatal(err)
	}
	/////////////////////////////////////

	/////////////////////get machine name and fDate
	pathParts := strings.Split(jsonPath, "/")
	fMachineName = strings.Split(pathParts[0], ".")[0]
	fDate = pathParts[1]
	/////////////////////////////////////

	defer f.Close()

	return eventFile, fMachineName, fDate
}

// func writeToDatabase(fMachineName, fDate string, eventFile ParsedEvents) {

// }

func main() {
	_, fName, _ := getToStruct("MMK-W-11271/2019_6_25.json")
	println(fName)
}
