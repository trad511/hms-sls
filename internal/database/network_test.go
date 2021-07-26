// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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
	"testing"

	"github.com/stretchr/testify/suite"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

type NetworkTestSuite struct {
	suite.Suite
}

func (suite *NetworkTestSuite) SetupSuite() {
	err := NewDatabase()
	if err != nil {
		suite.FailNowf("Unable create database", "err: %s", err)
	}
}

func (suite *NetworkTestSuite) TestCUDNetwork_HappyPath() {
	previousVersion, versionErr := GetCurrentVersion()
	suite.NoError(versionErr)

	network := sls_common.Network{
		Name:     "hmn",
		FullName: "Hardware Man Network",
		IPRanges: []string{"192.168.1.0/24"},
		Type:     "ethernet",
	}

	err := InsertNetwork(network)
	suite.NoError(err)

	err = InsertNetwork(network)
	suite.EqualError(err, AlreadySuch.Error())

	newVersion, versionErr := GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion

	// Update all the fields except for the name
	network.FullName = "Hardware Management Network"
	network.IPRanges = append(network.IPRanges, "176.16.0.0/16")
	network.Type = "mixed"

	err = UpdateNetwork(network)
	suite.NoError(err)

	newVersion, versionErr = GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion

	err = DeleteNetwork(network.Name)
	suite.NoError(err)

	newVersion, versionErr = GetCurrentVersion()
	suite.NoError(versionErr)
	suite.Greater(newVersion, previousVersion)
	previousVersion = newVersion
}

func (suite *NetworkTestSuite) TestRNetwork_HappyPath() {
	// Put in a network
	network := sls_common.Network{
		Name:     "nmn",
		FullName: "Node Man Network",
		IPRanges: []string{"192.168.1.0/24"},
		Type:     "ethernet",
		ExtraPropertiesRaw: sls_common.NetworkExtraProperties{
			CIDR: "192.168.1.0/24",
			MTU:  9000,
		},
	}

	err := InsertNetwork(network)
	suite.NoError(err)

	// Get the data back out
	returnedNetwork, err := GetNetworkForName("nmn")
	suite.NoError(err)

	_, err = json.MarshalIndent(returnedNetwork, "\t", "\t")
	suite.NoError(err)

	// Search for a network that contains an IP address
	returnedNetworks, err := GetNetworksContainingIP("192.168.1.5")
	suite.NoError(err)
	suite.GreaterOrEqual(len(returnedNetworks), 1)

	_, err = json.MarshalIndent(returnedNetworks, "\t", "\t")
	suite.NoError(err)

	// Search for a network that doesn't exist
	returnedNetworks, err = GetNetworksContainingIP("1.1.1.1")
	suite.EqualError(err, NoSuch.Error())
	suite.Equal(len(returnedNetworks), 0)

	// Do a free form search
	returnedNetworks, err = SearchNetworks(map[string]string{
		"name":      network.Name,
		"full_name": network.FullName,
		"ip_ranges": "192.168.1.2",
	}, map[string]interface{}{})
	suite.NoError(err)
	suite.GreaterOrEqual(len(returnedNetworks), 1)

	// Do a free form search using the networks extra properties
	returnedNetworks, err = SearchNetworks(
		map[string]string{},
		map[string]interface{}{
			"MTU": "9000",
		})
	suite.NoError(err)
	suite.Len(returnedNetworks, 1)

	// Make sure we don't get anything for this.
	returnedNetworks, err = SearchNetworks(map[string]string{
		"name":      network.Name,
		"full_name": "not a real name",
	}, map[string]interface{}{})
	suite.EqualError(err, NoSuch.Error())
	suite.Equal(len(returnedNetworks), 0)

	// Make sure we don't get anything for this search using the extra properties:
	returnedNetworks, err = SearchNetworks(
		map[string]string{},
		map[string]interface{}{
			"MTU": "9001",
		})
	suite.EqualError(err, NoSuch.Error())
	suite.Equal(len(returnedNetworks), 0)
}

func TestNetworkSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
