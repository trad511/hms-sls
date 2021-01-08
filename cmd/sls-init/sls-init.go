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

package main

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
	"log"
	"os"
	"stash.us.cray.com/HMS/hms-sls/internal/database"
	"strconv"
)

var (
	forceStep = flag.Int("force_step", -1, "Force migration to step X")
	fresh     = flag.Bool("fresh", false, "Revert all schemas before installing (drops all data)")
	schema_version = flag.String("schema_version", "latest", "Version of schema to migrate to")
)

func main() {
	flag.Parse()

	log.Printf("Beginning migration...")

	err := database.NewDatabase()

	driver, err := postgres.WithInstance(database.DB, &postgres.Config{})
	if err != nil {
		log.Printf("Creating Postgres driver failed: '%s'", err)
		os.Exit(1)
	}
	log.Printf("Creating Postgres driver succeeded.")

	m, err := migrate.NewWithDatabaseInstance("file:///persistent_migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Creating migration failed: '%s'!", err)
	}
	if m == nil {
		log.Fatalf("Creating migration failed: nil pointer!")
	}
	defer m.Close()

	log.Printf("Creating migration succeeded.")

	// Drop all tables.
	if *fresh {
		log.Printf("Fresh requested...migrating down...")
		err = m.Drop()
		if err != nil {
			log.Printf("Failed to drop: '%s'!", err)
		} else {
			log.Printf("Drop succeeded.")
		}
	}

	// User-defined force, doesn't matter if dirty or not
	if *forceStep >= 0 {
		err = m.Force(*forceStep)
		if err != nil {
			log.Printf("Force to %d failed: '%s'", forceStep, err)
			os.Exit(1)
		}
		log.Printf("Force to %d succeeded!", forceStep)
		os.Exit(0)
	}

	// No schema version explicitly given or `latest` used, just go up.
	if *schema_version == "" || *schema_version == "latest" {
		log.Printf("Schema version not set or value is latest...migrating up.")

		err = m.Up()
	} else {
		log.Printf("Migrating schema to version: %s", *schema_version)

		schemaVersionUInt64, parseErr := strconv.ParseUint(*schema_version, 10, 64)
		if parseErr != nil {
			log.Printf("Failed to parse schema version (%s): %s", *schema_version, parseErr)
			os.Exit(1)
		}

		err = m.Migrate(uint(schemaVersionUInt64))
	}
	if err == migrate.ErrNoChange {
		log.Printf("Migration required no changes.")
	} else if err != nil {
		log.Printf("Failed to migrate: '%s'!", err)
		os.Exit(1)
	} else {
		log.Printf("Migration succeeded.")
	}
}
