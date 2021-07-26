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
	"encoding/json"
	"github.com/Cray-HPE/hms-sls/pkg/sls-common"
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
