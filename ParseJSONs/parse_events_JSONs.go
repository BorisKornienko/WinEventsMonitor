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
	"errors"
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
	} `json:"System_Critical"`
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

func writeEvent(tableName string, eventID, eventSource, eventDescription, eventDateNtime, eventUser, string, eventCount int, eventFile ParsedEvents) (int64, error) {
	if eventCount == 0{
		return 0, nil
	}
	ctx := context.Background()
    var err error

    if db == nil {
        err = errors.New("Write event: db is null")
        return -1, err
	}
	
	// Check if database is alive.
	err = db.PingContext(ctx)
	if err != nil {
		return -1, err
	}

	tsql := "INSERT INTO WsEventsMonitor."+ tableName +" (machine, eventid, source, description, count, datentime, ip_v4, event_user) VALUES (@machine, @eventid, @source, @description, @count, @datentime, @ip_v4, @event_user);"
	tsql = fmt.Sprintf(tsql)
	stmt, err := db.Prepare(tsql)
    if err != nil {
       return -1, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(
        ctx,
        sql.Named("machine", eventFile.Computer),
		sql.Named("eventid", eventID),
		sql.Named("source", eventSource),
		sql.Named("description", eventDescription),
		sql.Named("count", eventCount),
		// In this place we string, it's must be a MSSQL DateTime
		// sql.Named("datentime", eventDateNtime),
		// No :) IP must be writed by SQlfunction or with the Go string to binary convert
		// sql.Named("ip_v4", eventFile.IP),
		sql.Named("event_user", eventUser))
    var newID int64
    err = row.Scan(&newID)
    if err != nil {
        return -1, err
    }

	return newID, nil
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
		JSONPath := filepath.Join(MachineFolder, f.Name())
		isProcessed, err := selectProcessed(MachineFolder, f.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		if isProcessed != 0{
			log.Println("this file is already: ", f.Name())
			err = os.Remove(JSONPath)
			if err != nil {
				log.Println("Cant delete ", JSONPath)
			}
			continue
		}
		
		eventStruct, err := getToStruct(JSONPath)
		if err != nil {
			log.Fatal(err)
			return -1, err
		}
		

		// for _, systemCrit := range(eventStruct.SystemCritical){

		// }
		
	}

	// NO! Not is 1. It is temp!
	return 1, err
}

func main() {
	// TEMP for testing
}
