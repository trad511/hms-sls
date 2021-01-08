/*
 * Copyright 2019 Cray Inc. All Rights Reserved.
 *
 * Except as permitted by contract or express written permission of Cray Inc.,
 * no part of this work or its content may be modified, used, reproduced or
 * disclosed in any form. Modifications made without express permission of
 * Cray Inc. may damage the system the software is installed within, may
 * disqualify the user from receiving support from Cray Inc. under support or
 * maintenance contracts, or require additional support services outside the
 * scope of those contracts to repair the software or system.
 *
 */

package database

import (
	"database/sql"

	"github.com/pkg/errors"
)

func IncrementVersion(trans *sql.Tx, updatedEntity string) (id int64, err error) {
	var version int64

	q := "INSERT INTO " +
		"    version_history (updated_entity) " +
		"VALUES " +
		"    ($1) " +
		"RETURNING (version)"

	result := trans.QueryRow(q, updatedEntity)

	transErr := result.Scan(&version)

	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	result.Scan()

	return version, err
}

func GetCurrentVersion() (version int, err error) {
	q := "SELECT " +
		"    max(version) " +
		"FROM " +
		"    version_history "

	row := DB.QueryRow(q)
	err = row.Scan(&version)

	return
}

func GetLastModified() (lastModified string, err error) {
	q := "SELECT " +
		"    max(timestamp) " +
		"FROM " +
		"    version_history "

	row := DB.QueryRow(q)
	err = row.Scan(&lastModified)

	return
}
