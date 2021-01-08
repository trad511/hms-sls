/* Copyright 2019 Cray Inc. All Rights Reserved.
 *
 * Except as permitted by contract or express written permission of Cray Inc.,
 * no part of this work or its content may be modified, used, reproduced or
 * disclosed in any form. Modifications made without express permission of
 * Cray Inc. may damage the system the software is installed within, may
 * disqualify the user from receiving support from Cray Inc. under support or
 * maintenance contracts, or require additional support services outside the
 * scope of those contracts to repair the software or system.
 */

package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
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
