package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ParsedEvents struct {
	ApplicationsCritical []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Critical"`
	SystemError []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Error"`
	IP                  string `json:"ip"`
	ApplicationsWarning []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Warning"`
	SystemCritical []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Syste_Critical"`
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

func getToStruct(jsonPath string) ParsedEvents {
	var eventFile ParsedEvents
	// var fMachineName string
	// var fDate string

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

	defer f.Close()

	return eventFile
}

func writeToDatabase(MachineFolder string, eventFile ParsedEvents) (int64, error) {
	JSONFiles, err := ioutil.ReadDir(MachineFolder)
	if err != nil {
		log.Fatal(err)
	}
	if JSONFiles == nil {
		return 0, nil
	}
	for _, f := range JSONFiles {
		JSONPath := filepath.Join(MachineFolder, f.Name())
		eventStruct := getToStruct(JSONPath)
	}
}

func main() {
	// TEMP for testing
}
