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

package datastore

import (
	"errors"
	"strings"

	"github.com/Cray-HPE/hms-sls/internal/database"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

var InvalidNetworkType = errors.New("invalid network type")
var InvalidNetworkName = errors.New("invalid network name")

func verifyNetworkType(networkType sls_common.NetworkType) error {
	networkTypeLower := strings.ToLower(string(networkType))
	if networkTypeLower != sls_common.NetworkTypeCassini.String() &&
		networkTypeLower != sls_common.NetworkTypeEthernet.String() &&
		networkTypeLower != sls_common.NetworkTypeInfiniband.String() &&
		networkTypeLower != sls_common.NetworkTypeMixed.String() &&
		networkTypeLower != sls_common.NetworkTypeOPA.String() &&
		networkTypeLower != sls_common.NetworkTypeSS10.String() {
		return InvalidNetworkType
	}

	return nil
}

func verifyNetworkName(networkName string) error {
	if strings.Contains(networkName, " ") {
		return InvalidNetworkName
	}

	return nil
}

// Helper function to verify network is of a correct type and name.
func verifyNetwork(nw sls_common.Network) error {
	typeErr := verifyNetworkType(nw.Type)
	if typeErr != nil {
		return typeErr
	}

	nameErr := verifyNetworkName(nw.Name)
	if nameErr != nil {
		return nameErr
	}

	return nil
}

// GetNetwork returns the network object matching the given name.
func GetNetwork(name string) (sls_common.Network, error) {
	return database.GetNetworkForName(name)
}

// InsertNetwork adds a given network into the database assuming it passes validation.
func InsertNetwork(network sls_common.Network) (err error) {
	err = verifyNetwork(network)
	if err != nil {
		return
	}

	err = database.InsertNetwork(network)

	return
}

// UpdateNetwork updates all of the fields for a given network in the DB *except* for the name which is read-only.
// Therefore, this function does no validation on network name.
func UpdateNetwork(network sls_common.Network) error {
	return database.UpdateNetwork(network)
}

// Insert or update a network
func SetNetwork(network sls_common.Network) error {
	err := verifyNetwork(network)
	if err != nil {
		return err
	}
	_, nwerr := GetNetwork(network.Name)
	if (nwerr != nil) && (nwerr != database.NoSuch) {
		return nwerr
	}

	if (nwerr != nil) && (nwerr == database.NoSuch) {
		inserr := database.InsertNetwork(network)
		if inserr != nil {
			return inserr
		}
	} else {
		upderr := database.UpdateNetwork(network)
		if upderr != nil {
			return upderr
		}
	}
	return nil
}

// DeleteNetwork removes a network from the DB.
func DeleteNetwork(networkName string) error {
	return database.DeleteNetwork(networkName)
}

// GetAllNetworks returns all the network objects in the DB.
func GetAllNetworks() ([]sls_common.Network, error) {
	return database.GetAllNetworks()
}

func SearchNetworks(network sls_common.Network) (networks []sls_common.Network, err error) {
	conditions := make(map[string]string)

	if network.Name != "" {
		err = verifyNetworkName(network.Name)
		if err != nil {
			return
		}

		conditions["name"] = network.Name
	}
	if network.FullName != "" {
		conditions["full_name"] = network.FullName
	}
	if len(network.IPRanges) == 1 && network.IPRanges[0] != "" {
		conditions["ip_ranges"] = network.IPRanges[0]
	}
	if network.Type != "" {
		err = verifyNetworkType(network.Type)
		if err != nil {
			return
		}

		conditions["type"] = string(network.Type)
	}

	propertiesMap, ok := network.ExtraPropertiesRaw.(map[string]interface{})
	if !ok {
		err = InvalidExtraProperties
		return
	}

	networks, err = database.SearchNetworks(conditions, propertiesMap)

	return
}

func ReplaceAllNetworks(networks []sls_common.Network) error {
	return database.ReplaceAllNetworks(networks)
}
