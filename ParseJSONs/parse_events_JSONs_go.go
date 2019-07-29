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
	"time"
	"strings"
	"bytes"
)

//ParsedEvents is for Event JSONs unmarshal
type ParsedEvents struct {
	ApplicationsCritical []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Critical"`
	SystemError []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Error"`
	IP                  string `json:"ip"`
	ApplicationsWarning []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Warning"`
	SystemCritical []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Critical"`
	Computer          string `json:"computer"`
	ApplicationsError []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"Applications_Error"`
	SystemWarning []struct {
		Source      string `json:"Source"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Count       int    `json:"count"`
		// MachineName string `json:"machineName"`
		DateNtime   string `json:"dateNtime"`
		User        string `json:"user"`
	} `json:"System_Warning"`
}

// this var's for database connections
var db *sql.DB

//DBPort is a standart MS SQL port
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
		log.Fatal("File open: ", err)
		return eventFile, err
	}
	// println("os opened: ", f)
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("File read: ", err)
		return eventFile, err
	}
	// println(string(byteValue))
	byteValue = bytes.TrimPrefix(byteValue, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(byteValue, &eventFile)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
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

	tsql := fmt.Sprintf("SELECT * FROM WinEventsMonitor.dbo.ProcessedFiles WHERE machineDir = @machineDir AND fileDateName = @fileDateName;")

	rows, err := db.QueryContext(
		ctx,
		tsql,
		sql.Named("machineDir", MachineFolder),
		sql.Named("fileDateName", fileDateName))
	if err != nil {
		// err = errors.New("cant read from ProcessedFiles table")
		err = fmt.Errorf("ProcessedFiles table read error: %q", err)
		return -1, err
	}
	defer rows.Close()

    var count int
    for rows.Next() {
        count++
	}
	return count, nil
}


func writeEvent(tableName, eventID, eventSource, eventDescription, eventDateNtime, eventUser string, eventCount int, eventFile ParsedEvents) (int64, error) {
	// write each once event in events list
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

	tsql := "INSERT INTO WinEventsMonitor.dbo."+tableName+"(machine, eventid, source, description, count, datentime, subnet2, subnet3, event_user) VALUES (@machine, @eventid, @source, @description, @count, @datentime, @subnet2, @subnet3, @event_user); select convert(bigint, SCOPE_IDENTITY());"
	tsql = fmt.Sprintf(tsql)
	stmt, err := db.Prepare(tsql)
    if err != nil {
       return -1, err
	}
	defer stmt.Close()
	// "dateNtime":  "2019:7:4-9:32:29",	'2004-05-23T14:25:10'
	eDate := strings.Split(eventDateNtime, "-")[0]
	eTime := strings.Split(eventDateNtime, "-")[1]
	
	// split date to get T-SQL standart date
	eYear := strings.Split(eDate, ":")[0]
	eMonth := strings.Split(eDate, ":")[1]
	if len(eMonth) < 2{
		eMonth = "0" + eMonth
	}
	eDay := strings.Split(eDate, ":")[2]
	if len(eDay) < 2 {
		eDay = "0"+eDay
	}
	// split time to get T-SQL standart time
	eHour := strings.Split(eTime, ":")[0]
	if len(eHour) < 2 {
		eHour = "0" + eHour
	}
	eMinute := strings.Split(eTime, ":")[1]
	if len(eMinute) < 2 {
		eMinute = "0" + eMinute
	}
	eSecond := strings.Split(eTime, ":")[2]
	if len(eSecond) < 2 {
		eSecond = "0" + eSecond
	}
	dateNtime := eYear+"-"+eMonth+"-"+eDay+"T"+eHour+":"+eMinute+":"+eSecond
	subnet2 := strings.Split(eventFile.IP, ".")[1]
	subnet3 := strings.Split(eventFile.IP, ".")[2]

	row := stmt.QueryRowContext(
        ctx,
        sql.Named("machine", eventFile.Computer),
		sql.Named("eventid", eventID),
		sql.Named("source", eventSource),
		sql.Named("description", eventDescription),
		sql.Named("count", eventCount),
		sql.Named("datentime", dateNtime),
		sql.Named("subnet2", subnet2),
		sql.Named("subnet3", subnet3),
		sql.Named("event_user", eventUser))
    var newID int64
    err = row.Scan(&newID)
    if err != nil {
        return -1, err
    }

	return newID, nil
}


func writeProcessed(machineFolder, jsonName, result string) (int64, error) {
	ctx := context.Background()
	var err error

	if db == nil {
		err = errors.New("Write event: db is null")
		return -1, err
	}
	today := time.Now()
	tsql := "INSERT INTO WinEventsMonitor.dbo.ProcessedFiles (machineDir, fileDateName, processedDate, result) VALUES (@machineDir, @fileDateName, @processedDate, @result); select convert(bigint, SCOPE_IDENTITY());"
	tsql = fmt.Sprintf(tsql)
	stmt, err := db.Prepare(tsql)
	if err != nil{
		return -1, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(
		ctx,
		sql.Named("machineDir", machineFolder),
		sql.Named("fileDateName", jsonName), 
		sql.Named("processedDate", today.Format("01-02-2006 15:04:05")),
		sql.Named("result", result))
	
	var newID int64
	err = row.Scan(&newID)
	if err != nil {
		return -1, err
	}

	return newID, nil
}

// Main database write function
func writeToDatabase(rootPath, machineFolder, DBUser, DBPassw, DBServerName, DBName string) (int, error) {
	machineFolderPath := rootPath+"\\"+machineFolder
	JSONFiles, err := ioutil.ReadDir(machineFolderPath)
	if err != nil {
		log.Fatal("Cant open Machine Folder: ", err)
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

	var succesDB int
	var failDB int
	
	succesDB = 0
	failDB = 0
	
	
	for _, f := range JSONFiles {
		JSONPath := filepath.Join(rootPath, machineFolder, f.Name())
		isProcessed, err := selectProcessed(machineFolder, f.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		if isProcessed != 0{
			log.Printf("this file is already processed: %s,%s", machineFolder, f.Name())
			err = os.Remove(JSONPath)
			if err != nil {
				log.Printf("Cant delete %s", JSONPath)
			}

			_, err := writeProcessed(machineFolder, f.Name(), "DENY")
			if err != nil{
				fmt.Println("Cant write already processed file: ", err)
			}
			// fmt.Println(id)
			
			continue
		}
		
		eventFile, err := getToStruct(JSONPath)
		if err != nil {
			log.Fatal(err)
			return -1, err
		}
		
		// System Criticals
		for _, systemCrit := range(eventFile.SystemCritical){
			_, err := writeEvent("SystemCriticals", systemCrit.ID, systemCrit.Source, systemCrit.Description, systemCrit.DateNtime, systemCrit.User, systemCrit.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		// System Errors
		for _, systemErr := range(eventFile.SystemError){
			_, err := writeEvent("SystemErrors", systemErr.ID, systemErr.Source, systemErr.Description, systemErr.DateNtime, systemErr.User, systemErr.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		// System Warnings
		for _, systemWarn := range(eventFile.SystemWarning){
			_, err := writeEvent("SystemWarnings", systemWarn.ID, systemWarn.Source, systemWarn.Description, systemWarn.DateNtime, systemWarn.User, systemWarn.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		// Applications Criticals
		for _, appsCrit := range(eventFile.ApplicationsCritical){
			_, err := writeEvent("ApplicationsCriticals", appsCrit.ID, appsCrit.Source, appsCrit.Description, appsCrit.DateNtime, appsCrit.User, appsCrit.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		// Applications Errors
		for _, appsErr := range(eventFile.ApplicationsError){
			_, err := writeEvent("ApplicationsErrors", appsErr.ID, appsErr.Source, appsErr.Description, appsErr.DateNtime, appsErr.User, appsErr.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		// Applications Warnings
		for _, appsWarn := range(eventFile.ApplicationsWarning){
			_, err := writeEvent("ApplicationsWarnings", appsWarn.ID, appsWarn.Source, appsWarn.Description, appsWarn.DateNtime, appsWarn.User, appsWarn.Count, eventFile)
			if err != nil{
				failDB ++
			}
			succesDB ++
			// fmt.Println("sucess DB writed: ", succesDB)
		}

		if failDB != 0{
			fmt.Printf("DB fails: %s, %s", machineFolder, f.Name())
			writeProcessed(machineFolder, f.Name(), "WARN")
		}else{
			writeProcessed(machineFolder, f.Name(), "ALLOW")
			os.Remove(JSONPath)
		}
	}

	if failDB != 0{
		err := fmt.Errorf("DB writes errors: %d", failDB)
		return succesDB, err
	}
	
	return succesDB, nil
}

func main() {
	rootPath := os.Args[1]
	DBUser := os.Args[2]
	DBPassw := os.Args[3]
	DBServerName := os.Args[4]
	DBName := os.Args[5]


	listMachineFolders, err := ioutil.ReadDir(rootPath)
	if err != nil{
		log.Fatal("cant read root folder ", rootPath)
	}
	f := func (machineFolder os.FileInfo) {
		if machineFolder.IsDir() != true {
			fmt.Println("it is not a directory: ", machineFolder.Name())
			// continue
		}
		
		_, err := writeToDatabase(rootPath, machineFolder.Name(), DBUser, DBPassw, DBServerName, DBName)
		if err != nil{
			log.Println(err)
			// continue
		}
	}
	for _, machineFolder := range listMachineFolders {

		go f(machineFolder)
	// fmt.Println(succesWrites)
	} 
	
	// go f(listMachineFolders)
}
