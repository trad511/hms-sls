// Copyright 2019 Cray Inc. All Rights Reserved.

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"

	"github.com/pkg/errors"
)

func InsertGenericHardware(hardware sls_common.GenericHardware) (err error) {
	q := "INSERT INTO \n" +
		"    components (xname, \n" +
		"                parent, \n" +
		"                comp_type, \n" +
		"                comp_class, \n" +
		"                extra_properties, \n" +
		"				 last_updated_version) \n" +
		"VALUES \n" +
		"($1, \n" +
		" $2, \n" +
		" $3, \n" +
		" $4, \n" +
		" $5, \n" +
		" $6) "

	jsonBytes, jsonErr := json.Marshal(hardware.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return err
	}

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return err
	}

	version, err := IncrementVersion(trans, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, hardware.Xname, hardware.Parent, hardware.Type, hardware.Class, string(jsonBytes), version)
	if transErr != nil {
		switch transErr.(type) {
		case *pq.Error:
			if transErr.(*pq.Error).Code.Name() == "unique_violation" {
				err = AlreadySuch
				_ = trans.Rollback()
				return
			}
		}

		// I bet we're getting back the wrong (or no) ID
		println(transErr)
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		_ = trans.Rollback()
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("insert generic component failed: %s", rowsErr)
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

func DeleteGenericHardware(hardware sls_common.GenericHardware) (err error) {
	q := "DELETE \n" +
		"FROM \n" +
		"    components \n" +
		"WHERE \n" +
		"    xname = $1 "

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	_, err = IncrementVersion(trans, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, hardware.Xname)
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

func DeleteAllGenericHardware() (err error) {
	q := "TRUNCATE " +
		"    components "

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	_, err = IncrementVersion(trans, "delete all hardware")
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

func UpdateGenericHardware(hardware sls_common.GenericHardware) (err error) {
	q := "UPDATE components \n" +
		"SET \n" +
		"    parent           = $2, \n" +
		"    comp_type        = $3, \n" +
		"    comp_class       = $4, \n" +
		"    extra_properties = $5, \n" +
		"    last_updated_version = $6 \n" +
		"WHERE \n" +
		"    xname = $1 "

	jsonBytes, jsonErr := json.Marshal(hardware.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	version, err := IncrementVersion(trans, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	result, transErr := trans.Exec(q, hardware.Xname, hardware.Parent, hardware.Type, hardware.Class, string(jsonBytes), version)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		_ = trans.Rollback()
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("update generic component failed: %s", rowsErr)
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

func getChildrenForXname(xname string) (children []string, err error) {
	// Now we find all the children for this object and add them to the base object.
	parentQ := "SELECT \n" +
		"    xname \n" +
		"FROM \n" +
		"    components \n" +
		"WHERE \n" +
		"    parent = $1 "
	childrenRows, parentErr := DB.Query(parentQ, xname)
	if parentErr != nil {
		err = errors.Errorf("unable to query children: %s", parentErr)
		return
	}

	for childrenRows.Next() {
		var thisChildXname string

		childErr := childrenRows.Scan(&thisChildXname)
		if childErr != nil {
			err = errors.Errorf("unable to scan child row: %s", childErr)
			return
		}

		children = append(children, thisChildXname)
	}

	return
}

func GetAllGenericHardware() (hardware []sls_common.GenericHardware, err error) {
	// First, get the base object and all its associated data
	baseQ := "SELECT \n" +
		"    xname, \n" +
		"    parent, \n" +
		"    comp_type, \n" +
		"    comp_class, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    components \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON components.last_updated_version = version_history.version"
	baseRows, baseErr := DB.Query(baseQ)
	if baseErr != nil {
		err = errors.Errorf("unable to query generic hardware: %s", baseErr)
		return
	}

	for baseRows.Next() {
		var thisGenericHardware sls_common.GenericHardware
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		baseErr := baseRows.Scan(&thisGenericHardware.Xname,
			&thisGenericHardware.Parent,
			&thisGenericHardware.Type,
			&thisGenericHardware.Class,
			&lastUpdated,
			&extraPropertiesBytes)
		if baseErr != nil {
			err = errors.Errorf("unable to scan base row: %s", baseErr)
			return
		}

		thisGenericHardware.LastUpdated = lastUpdated.Unix()
		thisGenericHardware.LastUpdatedTime = lastUpdated.String()

		thisGenericHardware.TypeString = sls_common.HMSStringTypeToHMSType(thisGenericHardware.Type)

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisGenericHardware.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
			return
		}

		var children []string
		children, err = getChildrenForXname(thisGenericHardware.Xname)
		if err != nil {
			return
		}

		thisGenericHardware.Children = children

		hardware = append(hardware, thisGenericHardware)
	}

	return
}

func GetGenericHardwareFromXname(xname string) (hardware sls_common.GenericHardware, err error) {
	// First, get the base object and all its associated data
	baseQ := "SELECT \n" +
		"    xname, \n" +
		"    parent, \n" +
		"    comp_type, \n" +
		"    comp_class, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    components \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON components.last_updated_version = version_history.version \n" +
		"WHERE \n" +
		"    xname = $1 "
	baseRow := DB.QueryRow(baseQ, xname)

	var extraPropertiesBytes []byte
	var lastUpdated time.Time
	baseErr := baseRow.Scan(&hardware.Xname,
		&hardware.Parent,
		&hardware.Type,
		&hardware.Class,
		&lastUpdated,
		&extraPropertiesBytes)
	if baseErr == sql.ErrNoRows {
		err = NoSuch
		return
	} else if baseErr != nil {
		err = errors.Errorf("unable to scan generic hardware row: %s", baseErr)
		return
	}

	hardware.TypeString = sls_common.HMSStringTypeToHMSType(hardware.Type)
	hardware.LastUpdated = lastUpdated.Unix()
	hardware.LastUpdatedTime = lastUpdated.String()

	unmarshalErr := json.Unmarshal(extraPropertiesBytes, &hardware.ExtraPropertiesRaw)
	if unmarshalErr != nil {
		err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
		return
	}

	children, err := getChildrenForXname(hardware.Xname)
	if err != nil {
		return
	}

	hardware.Children = children

	return
}

func GetGenericHardwareForExtraProperties(properties map[string]interface{}) (hardware []sls_common.GenericHardware,
	err error) {
	return SearchGenericHardware(nil, properties)
}

func SearchGenericHardware(conditions map[string]string, properties map[string]interface{}) (
	hardware []sls_common.GenericHardware, err error) {
	if len(conditions) == 0 && len(properties) == 0 {
		err = errors.Errorf("no conditions/properties with which to search")
		return
	}

	q := "SELECT \n" +
		"    xname, \n" +
		"    parent, \n" +
		"    comp_type, \n" +
		"    comp_class, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    components  \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON components.last_updated_version = version_history.version \n" +
		"WHERE \n     "

	// Build the conditions for the regular columns.
	index := 0
	for key, value := range conditions {
		if index != 0 {
			q = q + "  AND"
		}

		q = q + fmt.Sprintf(" %s = '%s' \n", key, value)

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
		}

		index++
	}

	rows, queryErr := DB.Query(q)
	if queryErr != nil {
		err = errors.Errorf("unable to query extra properties: %s", queryErr)
		return
	}

	for rows.Next() {
		newGenericHardware := sls_common.GenericHardware{}
		var extraPropertiesBytes []byte
		var lastUpdated time.Time

		scanErr := rows.Scan(&newGenericHardware.Xname,
			&newGenericHardware.Parent,
			&newGenericHardware.Type,
			&newGenericHardware.Class,
			&lastUpdated,
			&extraPropertiesBytes)
		if scanErr != nil {
			err = errors.Errorf("unable to scan row: %s", scanErr)
			return
		}

		newGenericHardware.TypeString = sls_common.HMSStringTypeToHMSType(newGenericHardware.Type)
		newGenericHardware.LastUpdated = lastUpdated.Unix()
		newGenericHardware.LastUpdatedTime = lastUpdated.String()

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &newGenericHardware.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
			return
		}

		var children []string
		children, err = getChildrenForXname(newGenericHardware.Xname)
		if err != nil {
			return
		}

		newGenericHardware.Children = children

		hardware = append(hardware, newGenericHardware)
	}

	return
}

func ReplaceAllGenericHardware(hardware []sls_common.GenericHardware) (err error) {
	trans, beginErr := DB.Begin()
	if beginErr != nil {
		err = errors.Errorf("unable to begin transaction: %s", beginErr)
		return
	}

	version, err := IncrementVersion(trans, "replaced all components")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = trans.Rollback()
		return err
	}

	// Start by deleting all the components currently there.
	q := "TRUNCATE " +
		"    components "

	_, transErr := trans.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		_ = trans.Rollback()
		return
	}

	// Now bulk load the passed in hardware into the database using a prepared statement.
	statement, prepareErr := trans.Prepare(pq.CopyIn("components",
		"xname", "parent", "comp_type", "comp_class", "last_updated_version", "extra_properties"))
	if prepareErr != nil {
		err = errors.Errorf("unable to prepare statement: %s", prepareErr)
		_ = trans.Rollback()
		return
	}

	for _, component := range hardware {
		jsonBytes, jsonErr := json.Marshal(component.ExtraPropertiesRaw)
		if jsonErr != nil {
			err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
			_ = trans.Rollback()
			return
		}

		_, execErr := statement.Exec(component.Xname, component.Parent, component.Type, component.Class,
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
