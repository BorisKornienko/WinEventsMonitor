package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	_ "github.com/denisenkom/go-mssqldb"
	"database/sql"
	"context"

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

// this var's for database connections
var db *sql.DB
var DBPort = 1433

func getToStruct(jsonPath string) (ParsedEvents, error) {
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
		return eventFile, err
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
		return eventFile, err
	}

	err = json.Unmarshal(byteValue, &eventFile)
	if err != nil {
		log.Fatal(err)
		return eventFile, err
	}

	defer f.Close()

	return eventFile, nil
}

func selectProcessed(MachineFolder, fileDateName string) (int, error) {
	// This is for check if file with (may be) processed events is again adds to list

	ctx := context.Background()
	//check is database is alive
	err := db.PingContext(ctx)
	if err != nil{
		return -1, err
	}

	tsql := fmt.Sprintf("SELECT * FROM WsEventsMonitor.ProcessedFiles WHERE machineDir = @machineDir AND fileDateName = @fileDateName;")

	rows, err := db.QueryContext(
		ctx,
		tsql,
		sql.Named("machineDir", MachineFolder),
		sql.Named("fileDateName", fileDateName))
	if err != nil {
		return -1, err
	}
	defer rows.Close()

    var count int
    for rows.Next() {
        count++
	}
	return count, nil
}

func writeToDatabase(MachineFolder, DBUser, DBPassw, DBServerName, DBName string,  eventFile ParsedEvents) (int64, error) {
	JSONFiles, err := ioutil.ReadDir(MachineFolder)
	if err != nil {
		log.Fatal(err)
	}
	if JSONFiles == nil {
		return 0, nil
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", DBServerName, DBUser, DBPassw, DBPort, DBName)

	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err)
		return -1, err
	}

	ctx := context.Background()
    err = db.PingContext(ctx)
    if err != nil {
        log.Fatal(err)
    }

	
	for _, f := range JSONFiles {
		isProcessed, err := selectProcessed(MachineFolder, f.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		if isProcessed != 0{
			log.Println("this file is already: ", f.Name())
			
			continue
		}
		JSONPath := filepath.Join(MachineFolder, f.Name())
		eventStruct, err := getToStruct(JSONPath)
		if err != nil {
			log.Fatal(err)
			return -1, err
		}
	}
	return 1, err
}

func main() {
	// TEMP for testing
}
