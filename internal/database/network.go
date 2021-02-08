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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

func InsertNetwork(network sls_common.Network) (err error) {
	q := "INSERT INTO \n" +
		"    network (name, \n" +
		"             full_name, \n" +
		"             ip_ranges, \n" +
		"             type, \n" +
		"             extra_properties, \n" +
		"             last_updated_version) \n" +
		"VALUES \n" +
		"($1, \n" +
		" $2, \n" +
		" $3, \n" +
		" $4, \n" +
		" $5, \n" +
		" $6) "

	jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	version, err := IncrementVersion(trans, network.Name)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, network.Name, network.FullName, pq.Array(network.IPRanges), network.Type, string(jsonBytes), version)
	if transErr != nil {
		switch transErr.(type) {
		case *pq.Error:
			if transErr.(*pq.Error).Code.Name() == "unique_violation" {
				err = AlreadySuch
				return
			}
		}

		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("insert network failed: %s", rowsErr)
		_ = trans.Rollback()
		return
	}
	if counter < 1 {
		err = NoSuch
		_ = trans.Rollback()
		return
	}

	commitErr := trans.Commit()
	if commitErr != nil {
		err = errors.Errorf("unable to commit transaction: %s", commitErr)
		return
	}

	return
}

func DeleteNetwork(networkName string) (err error) {
	q := "DELETE \n" +
		"FROM \n" +
		"    network \n" +
		"WHERE \n" +
		"    name = $1 "

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	_, err = IncrementVersion(trans, networkName)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, networkName)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		_ = trans.Rollback()
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("delete generic component failed: %s", rowsErr)
		_ = trans.Rollback()
		return
	}
	if counter < 1 {
		err = NoSuch
		_ = trans.Rollback()
		return
	}

	commitErr := trans.Commit()
	if commitErr != nil {
		err = errors.Errorf("unable to commit transaction: %s", commitErr)
		return
	}

	return
}

func DeleteAllNetworks() (err error) {
	q := "TRUNCATE " +
		"    network "

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	_, err = IncrementVersion(trans, "delete all networks")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	_, transErr := trans.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	commitErr := trans.Commit()
	if commitErr != nil {
		err = errors.Errorf("unable to commit transaction: %s", commitErr)
		return
	}

	return
}

func UpdateNetwork(network sls_common.Network) (err error) {
	q := "UPDATE network \n" +
		"SET \n" +
		"    full_name        = $2, \n" +
		"    ip_ranges        = $3, \n" +
		"    type             = $4, \n" +
		"    extra_properties = $5, \n" +
		"    last_updated_version = $6 \n" +
		"WHERE \n" +
		"    name = $1 "

	jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	version, err := IncrementVersion(trans, network.Name)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, network.Name, network.FullName, pq.Array(network.IPRanges), network.Type, string(jsonBytes), version)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("update network failed: %s", rowsErr)
		_ = trans.Rollback()
		return
	}
	if counter < 1 {
		err = NoSuch
		_ = trans.Rollback()
		return
	}

	commitErr := trans.Commit()
	if commitErr != nil {
		err = errors.Errorf("unable to commit transaction: %s", commitErr)
		return
	}

	return
}

func GetAllNetworks() (networks []sls_common.Network, err error) {
	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n"

	rows, rowsErr := DB.Query(q)
	if rowsErr != nil {
		err = errors.Errorf("unable to query network: %s", rowsErr)
		return
	}

	for rows.Next() {
		var thisNetwork sls_common.Network
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		scanErr := rows.Scan(&thisNetwork.Name,
			&thisNetwork.FullName,
			pq.Array(&thisNetwork.IPRanges),
			&thisNetwork.Type,
			&lastUpdated,
			&extraPropertiesBytes)
		if scanErr != nil {
			err = errors.Errorf("unable to scan network row: %s", scanErr)
			return
		}

		thisNetwork.LastUpdated = lastUpdated.Unix()
		thisNetwork.LastUpdatedTime = lastUpdated.String()

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisNetwork.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}

		networks = append(networks, thisNetwork)
	}

	return
}

func GetNetworkForName(name string) (network sls_common.Network, err error) {
	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network  \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n" +
		"WHERE \n" +
		"    name = $1 "

	row := DB.QueryRow(q, name)

	var extraPropertiesBytes []byte
	var lastUpdated time.Time
	baseErr := row.Scan(&network.Name,
		&network.FullName,
		pq.Array(&network.IPRanges),
		&network.Type,
		&lastUpdated,
		&extraPropertiesBytes)
	network.LastUpdated = lastUpdated.Unix()
	network.LastUpdatedTime = lastUpdated.String()
	if baseErr == sql.ErrNoRows {
		err = NoSuch
	} else if baseErr != nil {
		err = errors.Errorf("unable to scan network row: %s", baseErr)
	} else {
		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &network.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}
	}

	return
}

func GetNetworksContainingIP(addr string) (networks []sls_common.Network, err error) {
	return SearchNetworks(map[string]string{
		"ip_ranges": addr,
	}, map[string]interface{}{})
}

func SearchNetworks(conditions map[string]string, properties map[string]interface{}) (networks []sls_common.Network, err error) {
	if len(conditions) == 0 && len(properties) == 0 {
		err = errors.Errorf("no properties with which to search")
		return
	}

	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network  \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n" +
		"WHERE \n     "

	// Now build up the WHERE clause with the given conditions.
	index := 0
	for key, value := range conditions {
		if index != 0 {
			q = q + "  AND"
		}

		if key == "ip_ranges" {
			q = q + fmt.Sprintf(" '%s' <<= ANY(ip_ranges) \n", value)
		} else {
			q = q + fmt.Sprintf(" %s = '%s' \n", key, value)
		}

		index++
	}

	// Build the conditions for the extra properties JSON column.
	for key, value := range properties {
		if index != 0 {
			q = q + "  AND"
		}

		// Some day I want to come back around and make this work with infinite levels of depth, but for now just
		// investigate the type of the value interface. If it's a string then use one syntax, if it's an array use
		// another. The rational being that nested types need different syntax to query.
		valueString, ok := value.(string)
		if ok {
			q = q + fmt.Sprintf(" extra_properties ->> '%s' = '%s' \n", key, valueString)
		} else if valueArray, ok := value.([]string); ok {
			q = q + fmt.Sprintf(" extra_properties -> '%s' ?| array['%v'] \n", key,
				strings.Join(valueArray, "','"))
		} else {
			err = fmt.Errorf("Unable to query on parameter %s: %v", key, value)
			return
		}

		index++
	}

	rows, rowsErr := DB.Query(q)
	if rowsErr != nil {
		err = errors.Errorf("unable to query network: %s", rowsErr)
		return
	}

	for rows.Next() {
		var thisNetwork sls_common.Network
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		scanErr := rows.Scan(&thisNetwork.Name,
			&thisNetwork.FullName,
			pq.Array(&thisNetwork.IPRanges),
			&thisNetwork.Type,
			&lastUpdated,
			&extraPropertiesBytes)
		if scanErr != nil {
			err = errors.Errorf("unable to scan network row: %s", scanErr)
			return
		}

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisNetwork.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}

		thisNetwork.LastUpdated = lastUpdated.Unix()
		thisNetwork.LastUpdatedTime = lastUpdated.String()

		networks = append(networks, thisNetwork)
	}

	if len(networks) == 0 {
		err = NoSuch
	}

	return
}

func ReplaceAllNetworks(networks []sls_common.Network) (err error) {
	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	version, err := IncrementVersion(trans, "replaced all networks")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	// Start by deleting all the networks currently there.
	q := "TRUNCATE " +
		"    network "

	_, transErr := trans.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		_ = trans.Rollback()
		return
	}

	// Now bulk load the passed in hardware into the database using a prepared statement.
	statement, prepareErr := trans.Prepare(pq.CopyIn("network",
		"name", "full_name", "ip_ranges", "type", "last_updated_version", "extra_properties"))
	if prepareErr != nil {
		err = errors.Errorf("unable to prepare statement: %s", prepareErr)
		_ = trans.Rollback()
		return
	}

	for _, network := range networks {
		jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
		if jsonErr != nil {
			err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
			return
		}

		_, execErr := statement.Exec(network.Name, network.FullName, pq.Array(network.IPRanges), network.Type,
			version, string(jsonBytes))
		if execErr != nil {
			err = errors.Errorf("unable to exec statement: %s", execErr)
			_ = trans.Rollback()
			return
		}
	}

	_, statementErr := statement.Exec()
	if statementErr != nil {
		err = errors.Errorf("unable to exec statement: %s", statementErr)
		_ = trans.Rollback()
		return
	}

	statementErr = statement.Close()
	if statementErr != nil {
		err = errors.Errorf("unable to close statement: %s", statementErr)
		_ = trans.Rollback()
		return
	}

	// Now finally we can commit the entire transaction. Assuming this works, we're done here.
	commitErr := trans.Commit()
	if commitErr != nil {
		err = errors.Errorf("unable to commit transaction: %s", commitErr)
		return
	}

	return
}
