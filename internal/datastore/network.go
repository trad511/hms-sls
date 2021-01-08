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

package datastore

import (
	"errors"
	"strings"

	"stash.us.cray.com/HMS/hms-sls/internal/database"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
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
