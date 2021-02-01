// MIT License
//
// (C) Copyright [2019, 2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

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
