package main

import (
	"chilledoj/myreal/models"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/namsral/flag"
)

var (
	host  string
	port  int
	dbURL string
)

func init() {
	flag.StringVar(&host, "host", "", "WWW Host ip address for http connections")
	flag.IntVar(&port, "port", 8080, "WWW Port to listen for http connections")
	flag.StringVar(&dbURL, "dburl", "", "DB Connection URL")
	flag.Parse()
}

func main() {
	logger := log.New(os.Stdout, "[app] ", log.LstdFlags)

	logger.Printf("Opening connection to DB: %s", dbURL)
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Fatal(err)
	}

	appdb := &models.AppDB{DB: db}
	r := buildRouter(appdb, logger)

	addr := fmt.Sprintf("%s:%d", host, port)

	log.Fatal(http.ListenAndServe(addr, r))
}
