package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chilledoj/realworld-starter-kit/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/namsral/flag"
)

var (
	https                    bool
	origins, host, cert, key string
	port                     int
	dbURL                    string
)

func init() {
	flag.String(flag.DefaultConfigFlagname, "./config.ini", "path to config file")
	flag.BoolVar(&https, "https", false, "If https is provided -cert and -key must be defined")
	flag.StringVar(&cert, "cert", "./cert.pem", "HTTPS certificate filepath")
	flag.StringVar(&key, "key", "./key.pem", "HTTPS key filepath")
	flag.StringVar(&host, "host", "", "WWW Host ip address for http connections")
	flag.IntVar(&port, "port", 8080, "WWW Port to listen for http connections")
	flag.StringVar(&origins, "origins", "", "list of allowed origins for CORS")
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
	var protocol string
	if https {
		protocol = "https"
	} else {
		protocol = "http"
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	fulladdr := fmt.Sprintf("%s://%s", protocol, addr)

	r := buildRouter(appdb, logger, fulladdr)

	logger.Printf("Starting %s listener on %s", protocol, addr)
	logger.Printf("Navigate to %s", fulladdr)
	if https {
		err = http.ListenAndServeTLS(addr, cert, key, r)
	} else {
		err = http.ListenAndServe(addr, r)
	}
	log.Fatal(err)

}
