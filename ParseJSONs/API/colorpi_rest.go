package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	database   string
	collection string
)

const (
	mongoDBConnectionStringEnvVarName = "MONGODB_CONNECTION_STRING"
	mongoDBDatabaseEnvVarName         = "MONGODB_DATABASE"
	mongoDBCollectionEnvVarName       = "MONGODB_COLLECTION"
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

func connect() *mongo.Client {
	mongoDBConnectionString := os.Getenv(mongoDBConnectionStringEnvVarName)
	if mongoDBConnectionString == "" {
		log.Fatal("missing environment variable: ", mongoDBCollectionEnvVarName)
	}

	database = os.Getenv(mongoDBDatabaseEnvVarName)
	if database == "" {
		log.Fatal("missing environment variable: ", mongoDBDatabaseEnvVarName)
	}

	collection = os.Getenv(mongoDBCollectionEnvVarName)
	if collection == "" {
		log.Fatal("missing environment variable: ", mongoDBCollectionEnvVarName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoDBConnectionString).SetDirect(true)
	c, err := mongo.NewClient(clientOptions)

	err = c.Connect(ctx)

	if err != nil {
		log.Fatalf("unable to initializa connection %v", err)
	}
	err = c.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("unable to connect %v", err)
	}

	return c

}

func findDateMark(w http.ResponseWriter, r *http.Request) {
	c := connect()
	ctx := context.Background()
	defer c.Disconnect(ctx)

	eventsPack := EventsPack{}
	vars := mux.Vars(r)
	name := vars["name"]
	datemark := vars["datemark"]
	log.Print("finding datemark for computer ", name)

	collection := c.Database(database).Collection(collection)
	// fmt
	rs, err := collection.Find(ctx, bson.M{"computer": name, "datemark": datemark})
	if err != nil {
		fmt.Println("Error occured whil reading from DB ", err)
		return
	}
	err = rs.All(ctx, &eventsPack)
	if err != nil {
		log.Fatalf("failed to list datemark(s) %v", err)
	}

	json.NewEncoder(w).Encode(eventsPack)
}

func addEventsPack(w http.ResponseWriter, r *http.Request) {
	c := connect()
	ctx := context.Background()
	defer c.Disconnect(ctx)

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

	collection := c.Database(database).Collection(collection)
	resp, err := collection.InsertOne(ctx, eventsPack)
	if err != nil {
		log.Print("error occured while inserting document in database :: ", err)
		return
	}
	fmt.Fprintf(w, "last created document computer is :: ", eventsPack.Computer)
	fmt.Println(w, "last created document computer is :: ", resp.InsertedID)
}

func getDbNames(w http.ResponseWriter, r *http.Request) {
	c := connect()
	ctx := context.Background()
	defer c.Disconnect(ctx)

	db, err := c.ListDatabaseNames(ctx, bson.D{})
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
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("error starting http server :: ", err)
		return
	}
}
