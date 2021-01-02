package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	ConnHost = "localhost"
	ConnPort = "8080"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}
type Routes []Route

var routes = Routes{
	Route{
		"getEvents",
		"GET",
		"/events",
		getEvents,
	},
	Route{
		"getEventsID",
		"GET",
		"/event/{id}",
		getEventsID,
	},
	Route{
		"getComputer",
		"GET",
		"/computer/{name}",
		getComputer,
	},
	Route{
		"getEventsSeverity",
		"GET",
		"/events/{severity}",
		getEventsSeverity,
	},
	Route{
		"getEventsDateNtime",
		"GET",
		"/events/{datentime}",
		getEventsDateNtime,
	},
	Route{
		"addEventsPack",
		"POST",
		"/eventspack/add",
		addEventsPack,
	},
	Route{
		"getDbNames",
		"GET",
		"/getdbnames",
		getDbNames,
	},
	Route{
		"findDateMark",
		"GET",
		"/findDateMark/{name}/{datemark}",
		findDateMark,
	},
}

type EventGen []struct {
	Source      string    `json:"Source"`
	Description string    `json:"description"`
	ID          string    `json:"id"`
	Count       int       `json:"count"`
	DateNtime   time.Time `json:"dateNtime"`
	User        string    `json:"user"`
}
type EventsPack struct {
	ApplicationsCritical []struct {
		EventGen
	} `json:"Applications_Critical"`
	SystemError []struct {
		EventGen
	} `json:"System_Error"`
	IP                  string `json:"ip"`
	ApplicationsWarning []struct {
		EventGen
	} `json:"Applications_Warning"`
	SystemCritical []struct {
		EventGen
	} `json:"System_Critical"`
	Computer          string `json:"computer"`
	Datemark          string `json:"dateMark"`
	ApplicationsError []struct {
		EventGen
	} `json:"Applications_Error"`
	SystemWarning []struct {
		EventGen
	} `json:"System_Warning"`
}

var session *mgo.Session
var connectionError error

func init() {
	var dbUser string
	var dbPassword string
	mongoDbUrl := "mongodb://" + dbUser + ":" + dbPassword + "@colorpi.mongo.cosmos.azure.com:10255/?ssl=true&replicaSet=globaldb&maxIdleTimeMS=120000&appName=@colorpi@"
	session, connectionError = mgo.Dial(mongoDbUrl)
	if connectionError != nil {
		log.Fatal("error connecting to databaase :: ", connectionError)
	}
	session.SetMode(mgo.Monotonic, true)
}

func findDateMark(w http.ResponseWriter, r *http.Request) {
	eventsPack := EventsPack{}
	vars := mux.Vars(r)
	name := vars["name"]
	datemark := vars["datemark"]
	log.Print("finding datemark for computer ", name)
	collection := session.DB("allWsMonitor").C("WinEvents")
	// fmt
	findErr := collection.Find(bson.M{"computer": name, "datemark": datemark}).One(&eventsPack)
	if findErr != nil {
		fmt.Println("Error occured whil reading from DB ", findErr)
		return
	}
	json.NewEncoder(w).Encode(eventsPack)
}

func addEventsPack(w http.ResponseWriter, r *http.Request) {
	eventsPack := EventsPack{}
	// err := json.NewDecoder(r.Body).Decode(&eventsPack)
	b, err := ioutil.ReadAll(r.Body)
	b = []byte(b)
	defer r.Body.Close()
	err = json.Unmarshal(b, &eventsPack)
	// DEBUG INFO
	fmt.Println(eventsPack.Computer, eventsPack.Datemark)
	if err != nil {
		log.Print("error occured while decoding events data :: ", err)
		return
	}
	collection := session.DB("allWsMonitor").C("WinEvents")
	err = collection.Insert(eventsPack)
	if err != nil {
		log.Print("error occured while inserting document in database :: ", err)
		return
	}
	fmt.Fprintf(w, "last created document computer is :: ", eventsPack.Computer)
}

func getDbNames(w http.ResponseWriter, r *http.Request) {
	db, err := session.DatabaseNames()
	if err != nil {
		log.Print("error getting database names :: ", err)
		return
	}
	fmt.Fprintf(w, "Databases names are :: %s", strings.Join(db, ", "))
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Events placeholder")
}
func getEventsID(w http.ResponseWriter, r *http.Request) {

}
func getComputer(w http.ResponseWriter, r *http.Request) {

}
func getEventsSeverity(w http.ResponseWriter, r *http.Request) {

}
func getEventsDateNtime(w http.ResponseWriter, r *http.Request) {

}

func addRoutes(router *mux.Router) *mux.Router {
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return router
}

// Invoke-WebRequest http://localhost:8080/eventspack/add -Method POST -Body $xxx  -ContentType "text/plain; charset=utf-8"

func main() {
	muxRouter := mux.NewRouter().StrictSlash(true)
	router := addRoutes(muxRouter)
	defer session.Close()
	err := http.ListenAndServe(ConnHost+":"+ConnPort, router)
	if err != nil {
		log.Fatal("error starting http server :: ", err)
		return
	}
}
