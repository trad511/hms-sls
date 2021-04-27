/*
 * MIT License
 *
 * (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var DB *sql.DB

var mutex = &sync.Mutex{}
var DBInitialized bool

var NoSuch = errors.New("nothing found by that name")
var AlreadySuch = errors.New("entity already exists by that name")

func NewDatabase() (err error) {
	mutex.Lock()
	if !DBInitialized {
		connStr := getConnectionString()

		var openErr error
		DB, openErr = sql.Open("postgres", connStr)
		if openErr != nil {
			err = errors.Errorf("unable to open connection to Postgres: %s", openErr)
			return
		}

		maxDatabaseConnections := 25; // By deafult the postgres database has 100 connections MAX. There are usually 3 replicas for SLS, so each one can get 25
		if maxDatabaseConnectionsRaw, ok := os.LookupEnv("SLS_MAX_DATABASE_CONNECTIONS"); ok {
			maxDatabaseConnections, err = strconv.Atoi(maxDatabaseConnectionsRaw)
			if err != nil {
				log.Printf("Unable to parse SLS_MAX_DATABASE_CONNECTIONS environment variable: %v", err)
				return err
			}
		}

		log.Printf("Max database connections: %d", maxDatabaseConnections)
		DB.SetMaxOpenConns(maxDatabaseConnections)  
		DB.SetConnMaxLifetime(time.Minute)

		for {
			pingErr := DB.Ping()
			if pingErr != nil {
				log.Printf("DB connection failed ('%s'), retrying after 1 second...", pingErr)
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
		log.Printf("Connected to Postgres successfully")

		DBInitialized = true
	}
	mutex.Unlock()

	return
}

func CloseDatabase() (err error) {
	mutex.Lock()
	if DBInitialized {
		err = DB.Close()
	}
	mutex.Unlock()

	return
}

func getConnectionString() string {
	connStr := ""

	if database, ok := os.LookupEnv("DBNAME"); ok {
		connStr = fmt.Sprintf("dbname=%s", database)
	} else {
		connStr = fmt.Sprintf("dbname=sls")
	}

	if user, ok := os.LookupEnv("DBUSER"); ok {
		connStr = fmt.Sprintf("%s user=%s", connStr, user)
	} else {
		connStr = fmt.Sprintf("%s user=slsuser", connStr)
	}

	if pass, ok := os.LookupEnv("DBPASS"); ok {
		connStr = fmt.Sprintf("%s password=%s", connStr, pass)
	}

	if host, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		connStr = fmt.Sprintf("%s host=%s", connStr, host)
	} else {
		connStr = fmt.Sprintf("%s host=localhost", connStr)
	}

	if port, ok := os.LookupEnv("POSTGRES_PORT"); ok {
		portInt, _ := strconv.Atoi(port)
		connStr = fmt.Sprintf("%s port=%d", connStr, portInt)
	} else {
		connStr = fmt.Sprintf("%s port=5432", connStr)
	}

	if opts, ok := os.LookupEnv("DBOPTS"); ok {
		connStr = fmt.Sprintf("%s %s", connStr, opts)
	}
	return connStr
}
