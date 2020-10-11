package spring

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type SpringApplication struct {
	muxRouter *mux.Router
	db        *sql.DB
}

var app SpringApplication

func init() {
	viper.SetConfigName("application")     // name of config file (without extension)
	viper.SetConfigType("json")            // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("../../resources") // optionally look for config in the working directory
	err := viper.ReadInConfig()            // Find and read the config file
	if err != nil {                        // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	//
	databaseName := viper.Get("spring.gpa.datasource.name").(string)
	user := viper.Get("spring.gpa.datasource.username").(string)
	password := viper.Get("spring.gpa.datasource.password").(string)
	url := viper.Get("spring.gpa.datasource.address").(string)
	platform := viper.Get("spring.gpa.platform").(string)
	//
	connStr := platform + "://" + user + ":" + password + "@" + url + "/" + databaseName + "?sslmode=disable"
	log.Println("Connecting to database: " + databaseName)
	log.Println("Database address: " + url)
	db, err := sql.Open(platform, connStr)
	if err != nil {
		panic(fmt.Errorf("Fatal error database connection: %s \n", err))
	}
	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("Fatal error testing database connection: %s \n", err))
	}
	app.db = db
	app.muxRouter = mux.NewRouter().StrictSlash(true)
	log.Println("STARTING SPRING APP")
}

func Run() {
	srv := &http.Server{
		Handler: app.muxRouter,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("App started on port 8000")
	log.Fatal(srv.ListenAndServe())
}
