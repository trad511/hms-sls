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
	"encoding/json"
	"stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenericHardwareTestSuite struct {
	suite.Suite
}

func (suite *GenericHardwareTestSuite) SetupSuite() {
	err := NewDatabase()
	if err != nil {
		suite.FailNowf("Unable create database", "err: %s", err)
	}
}

func (suite *GenericHardwareTestSuite) TestCUDGenericHardware_HappyPath() {
	previousVersion, versionErr := GetCurrentVersion()
	suite.NoError(versionErr)

	genericHardware := sls_common.GenericHardware{
		Parent: "x0c0s0b0",
		Xname:  "x0c0s0b0n0",
		Type:   sls_common.Node,
		Class:  sls_common.ClassRiver,
		ExtraPropertiesRaw: sls_common.ComptypeNode{
			NID:  12,
			Role: "Compute",
		},
	}

	err := InsertGenericHardware(genericHardware)
	suite.NoError(err)

	err = InsertGenericHardware(genericHardware)
	suite.EqualError(err, AlreadySuch.Error())

	newVersion, versionErr := GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion

	// Update all the fields except for the xname
	genericHardware.Parent = "x0c0s0b1"
	genericHardware.Type = sls_common.Chassis
	genericHardware.Class = sls_common.ClassMountain
	genericHardware.ExtraPropertiesRaw = sls_common.ComptypeRtrMod{PowerConnector: "foo"}

	err = UpdateGenericHardware(genericHardware)
	suite.NoError(err)

	newVersion, versionErr = GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion

	err = DeleteGenericHardware(genericHardware)
	suite.NoError(err)

	newVersion, versionErr = GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion
}

func (suite *GenericHardwareTestSuite) TestRGetGenericHardware_HappyPath() {
	// Insert some children
	for i := 0; i < 5; i++ {
		genericHardware := sls_common.GenericHardware{
			Parent: "x1c0s0b0",
			Xname:  "x0c0s1b0n" + strconv.Itoa(i),
			Type:   sls_common.Node,
			Class:  sls_common.ClassRiver,
			ExtraPropertiesRaw: sls_common.ComptypeNodeHsnNic{
				Networks: []string{"NMN"},
				Peers:    []string{"x0c0r0i" + strconv.Itoa(i)},
			},
		}

		_ = InsertGenericHardware(genericHardware)
	}

	// Now put a parent in there
	genericHardware := sls_common.GenericHardware{
		Parent: "x1c0s0",
		Xname:  "x1c0s0b0",
		Type:   sls_common.Node,
		Class:  sls_common.ClassRiver,
		ExtraPropertiesRaw: sls_common.ComptypeNode{
			NID:  13,
			Role: "Compute",
		},
	}

	_ = InsertGenericHardware(genericHardware)

	// Now get the data back out.
	returnedHardware, err := GetGenericHardwareFromXname(genericHardware.Xname)
	suite.NoError(err)

	_, err = json.MarshalIndent(returnedHardware, "\t", "\t")
	suite.NoError(err)

	properties := make(map[string]interface{})
	properties["NID"] = "13"
	properties["Role"] = "Compute"

	// Try a search
	searchResults, err := GetGenericHardwareForExtraProperties(properties)
	suite.NoError(err)
	suite.GreaterOrEqual(len(searchResults), 1)

	// Try a more stringent search
	properties = make(map[string]interface{})
	properties["Networks"] = []string{"NMN"}

	searchResults, err = SearchGenericHardware(
		map[string]string{
			"Xname": "x0c0s1b0n0",
		},
		properties)
	suite.NoError(err)
	suite.Equal(len(searchResults), 1)

	_, err = json.MarshalIndent(searchResults, "\t", "\t")
	suite.NoError(err)

	// Try getting everything
	allGenericHardware, err := GetAllGenericHardware()
	suite.NoError(err)
	suite.GreaterOrEqual(len(allGenericHardware), 1)

	_, err = json.MarshalIndent(allGenericHardware, "\t", "\t")
	suite.NoError(err)
}

func TestGenericHardwareSuite(t *testing.T) {
	suite.Run(t, new(GenericHardwareTestSuite))
}
