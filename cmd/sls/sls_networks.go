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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-sls/internal/database"
	"stash.us.cray.com/HMS/hms-sls/internal/datastore"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

//  /networks GET API

func doNetworksGet(w http.ResponseWriter, r *http.Request) {
	// Get the networks from the database
	networks, err := datastore.GetAllNetworks()
	if err != nil {
		log.Println("ERROR: Can't get networks from DB:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Unable to get networks from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(networks)
	if err != nil {
		log.Println("ERROR: JSON marshal of networks failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

//  /networks POST API

func doNetworksPost(w http.ResponseWriter, r *http.Request) {
	var network sls_common.Network

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR: Failed to read body: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Failed to read body",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	err = json.Unmarshal(bodyBytes, &network)
	if err != nil {
		log.Println("ERROR: Failed to unmarshal body: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Failed to unmarshal body",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	// Now add it to the database.
	err = datastore.InsertNetwork(network)
	if err == database.AlreadySuch {
		log.Println("ERROR: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Conflict",
			"A network with that name already exists in the database",
			r.URL.Path, http.StatusConflict)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if err != nil {
		log.Println("ERROR: Failed to insert network into DB: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to insert network into DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(network)
	if err != nil {
		log.Println("ERROR: JSON marshal of network failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(ba)
}

//  /networks/{name} GET API

func doNetworkObjGet(w http.ResponseWriter, r *http.Request) {
	// Figure out what the requested network is
	networkName := mux.Vars(r)["network"]

	// Get the networks from the database
	network, err := datastore.GetNetwork(networkName)
	if err == database.NoSuch {
		log.Println("ERROR: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Not Found",
			"Network not found in DB",
			r.URL.Path, http.StatusNotFound)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if err != nil {
		log.Println("ERROR: Failed to get network from DB:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to get network from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(network)
	if err != nil {
		log.Println("ERROR: JSON marshal of network failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

//  /networks/{network} PUT API

func doNetworkObjPut(w http.ResponseWriter, r *http.Request) {
	// Figure out what the requested network is
	networkName := mux.Vars(r)["network"]
	var network sls_common.Network

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR: Failed to read body: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Failed to read body",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	err = json.Unmarshal(bodyBytes, &network)
	if err != nil {
		log.Println("ERROR: Failed to unmarshal body: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Failed to unmarshal body",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	// Make sure that the name is set to what the URL says and *not* what the payload coming from the user has.
	// This won't update the network name regardless, the database doesn't have that logic, it's only used as a
	// reference for which row to update.
	network.Name = networkName

	// Now do the update.
	err = datastore.SetNetwork(network)

	if err != nil {
		log.Println("ERROR: Failed to update network in DB:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to update network in DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(network)
	if err != nil {
		log.Println("ERROR: JSON marshal of network failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

//  /networks/{network} PATCH API

func doNetworkObjPatch(w http.ResponseWriter, r *http.Request) {
	log.Printf("doNetworkObjPatch: not implemented yet.\n")
	w.WriteHeader(http.StatusNotImplemented)
}

//  /networks/{network} DELETE API

func doNetworkObjDelete(w http.ResponseWriter, r *http.Request) {
	// Figure out what the requested network is
	networkName := mux.Vars(r)["network"]

	// Delete the network from the DB
	err := datastore.DeleteNetwork(networkName)
	if err == database.NoSuch {
		log.Println("ERROR: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Not Found",
			"Network not found in DB",
			r.URL.Path, http.StatusNotFound)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if err != nil {
		log.Println("ERROR: Failed to delete network from DB:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to delete network from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//  /search/networks GET API

func doNetworksSearch(w http.ResponseWriter, r *http.Request) {
	network := sls_common.Network{
		Name:               r.FormValue("name"),
		FullName:           r.FormValue("full_name"),
		IPRanges:           []string{r.FormValue("ip_address")},
		Type:               sls_common.NetworkType(r.FormValue("type")),
		ExtraPropertiesRaw: nil,
	}

	// Build up the extra properties section by gathering the various possible query object and adding them.
	properties := make(map[string]interface{})

	// The ExtraProperties section of SLS is probably the most powerful concept it has. To support generic queries
	// WITHOUT having to code in conditions for each possible field, look for everything that begins with:
	//   `extra_properties.`
	// And add each of them to the map for searching.
	for key, value := range r.Form {
		if strings.HasPrefix(key, "extra_properties.") {
			// What comes after the period is the name of the property.
			keyParts := strings.SplitN(key, ".", 2)
			if len(keyParts) != 2 || keyParts[1] == "" {
				log.Println("ERROR: ExtraProperties search does not include field")
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to search hardware in DB. ExtraProperties search does not include field.",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			// Support multiple values if they're provided.
			if len(value) == 1 {
				properties[keyParts[1]] = value[0]
			} else {
				properties[keyParts[1]] = value
			}
		}
	}

	network.ExtraPropertiesRaw = properties

	networks, err := datastore.SearchNetworks(network)
	if err == database.NoSuch {
		log.Println("ERROR: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Not Found",
			"Network not found in DB",
			r.URL.Path, http.StatusNotFound)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if err != nil {
		log.Println("ERROR: Failed to search for network:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to search network from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(networks)
	if err != nil {
		log.Println("ERROR: JSON marshal of networks failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}
