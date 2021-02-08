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

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
	"strings"

	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-sls/internal/database"

	"github.com/gorilla/mux"
	"stash.us.cray.com/HMS/hms-sls/internal/datastore"
)

//  /hardware POST API

func doHardwarePost(w http.ResponseWriter, r *http.Request) {
	var jdata sls_common.GenericHardware
	var tstr string

	// Decode the JSON to see what we are to post

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR reading request body:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "error reading REST request")
		return
	}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		log.Println("ERROR unmarshalling request body:", err)
		sendJsonRsp(w, http.StatusBadRequest, "error decoding JSON")
		return
	}

	if jdata.Xname == "" {
		log.Printf("ERROR, request JSON has empty Xname field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing required Xname field")
		return
	}
	if !base.IsHMSCompIDValid(jdata.Xname) {
		log.Printf("ERROR, request JSON has invalid Xname field: '%s'.\n",
			jdata.Xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid Xname field")
		return
	}
	if base.GetHMSCompParent(jdata.Xname) != "" {
		if jdata.Parent == "" {
			log.Printf("ERROR, request JSON has empty Parent field.\n")
			sendJsonRsp(w, http.StatusBadRequest, "missing required Parent field")
			return
		}
		if !base.IsHMSCompIDValid(jdata.Parent) {
			log.Printf("ERROR, request JSON has invalid Parent field: '%s'.\n",
				jdata.Parent)
			sendJsonRsp(w, http.StatusBadRequest, "invalid Parent field")
			return
		}
	}

	if jdata.Class == "" {
		log.Printf("ERROR, request JSON has empty Class field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing required Class field")
		return
	}
	if !sls_common.IsCabinetTypeValid(jdata.Class) {
		log.Printf("ERROR, request JSON has invalid Class field: '%s'.\n",
			string(jdata.Class))
		sendJsonRsp(w, http.StatusBadRequest, "invalid Class field")
		return
	}
	if jdata.Type == "" {
		log.Printf("ERROR, request JSON has empty Type field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing Type field")
		return
	}
	tstr = string(sls_common.HMSStringTypeToHMSType(jdata.Type))
	if tstr == string(sls_common.HMSTypeInvalid) {
		log.Printf("ERROR, request JSON has invalid Type field: '%s'.\n",
			string(jdata.Type))
		sendJsonRsp(w, http.StatusBadRequest, "invalid Type field")
		return
	}
	if jdata.TypeString == "" {
		log.Printf("ERROR, request JSON has empty TypeString field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing TypeString field")
		return
	}
	tstr = string(sls_common.HMSTypeToHMSStringType(jdata.TypeString))
	if string(tstr) == string(sls_common.HMSTypeInvalid) {
		log.Printf("ERROR, request JSON has invalid TypeString field: '%s'.\n",
			string(jdata.TypeString))
		sendJsonRsp(w, http.StatusBadRequest, "invalid TypeString field")
		return
	}

	// Check if the component already exists.  If so, it's an error.

	cname, cerr := datastore.GetXname(jdata.Xname)
	if cerr != nil {
		log.Printf("ERROR accessing DB for '%s': %v", jdata.Xname, cerr)
		sendJsonRsp(w, http.StatusInternalServerError, "DB lookup error")
		return
	}
	if cname != nil {
		log.Printf("ERROR, DB object exists for '%s'.\n", jdata.Xname)
		sendJsonRsp(w, http.StatusConflict, "object already exists")
		return
	}

	// Write these into the DB

	err = datastore.SetXname(jdata.Xname, jdata)
	if err != nil {
		log.Printf("ERROR inserting component '%s' into DB: %s\n", jdata.Xname, err)
		sendJsonRsp(w, http.StatusInternalServerError, "error inserting object into DB")
		return
	}

	sendJsonRsp(w, http.StatusOK, "inserted new entry")
}

//  /hardware GET API

func doHardwareGet(w http.ResponseWriter, r *http.Request) {
	hwList, err := datastore.GetAllXnameObjects()
	if err != nil {
		log.Println("ERROR getting all /hardware objects from DB:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed hardware DB query")
		return
	}
	ba, baerr := json.Marshal(hwList)
	if baerr != nil {
		log.Println("ERROR: JSON marshal of /hardware failed:", baerr)
		sendJsonRsp(w, http.StatusInternalServerError, "JSON marshal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

//  /hardware/{xname} GET API

func doHardwareObjGet(w http.ResponseWriter, r *http.Request) {
	// Decode the URL to get the XName
	vars := mux.Vars(r)
	xname := base.NormalizeHMSCompID(vars["xname"])

	if !base.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, invalid xname in request URL: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	// Fetch the item and all of its descendants from the database.  If
	// the item does not exist, error.

	cmp, err := datastore.GetXname(xname)
	if cmp == nil {
		log.Printf("ERROR, requested component not found in DB: '%s'\n",
			xname)
		sendJsonRsp(w, http.StatusNotFound, "no such component not in DB")
		return
	}
	if err != nil {
		log.Println("ERROR, DB query failed:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}

	// Return the HW component.

	sendJsonCompRsp(w, *cmp)
}

//  /hardware/{xname} PUT API

func doHardwareObjPut(w http.ResponseWriter, r *http.Request) {
	var cmp, jdata sls_common.GenericHardware

	// Decode the URL to get the XName

	vars := mux.Vars(r)
	xname := base.NormalizeHMSCompID(vars["xname"])

	if !base.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, PUT request with xname: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	//Unmarshal the payload.

	body, berr := ioutil.ReadAll(r.Body)
	if berr != nil {
		log.Println("ERROR reading request body:", berr)
		sendJsonRsp(w, http.StatusInternalServerError, "unable to read request body")
		return
	}

	berr = json.Unmarshal(body, &jdata)
	if berr != nil {
		log.Println("ERROR unmarshalling request body:", berr)
		sendJsonRsp(w, http.StatusBadRequest, "unable to unmarshal JSON payload")
		return
	}

	//Insure mandatory fields are present

	errstr := ""
	if jdata.Xname == "" {
		errstr = errstr + "Xname "
	}
	if jdata.Parent == "" {
		errstr = errstr + "Parent "
	}
	if jdata.Type == "" {
		errstr = errstr + "Type "
	}
	if jdata.TypeString == "" {
		errstr = errstr + "TypeString "
	}
	if jdata.Class == "" {
		errstr = errstr + "Class "
	}
	if errstr != "" {
		errstr = "missing fields " + errstr
		log.Printf("ERROR in PUT JSON: '%s'\n", errstr)
		sendJsonRsp(w, http.StatusBadRequest, errstr)
		return
	}

	//If the payload GenericComponent has an Xname, be sure it
	//matches the one in the URL.

	if strings.ToLower(xname) != strings.ToLower(jdata.Xname) {
		log.Printf("ERROR, PUT request JSON Xname != URL xname (%s/%s)\n",
			strings.ToLower(jdata.Xname), strings.ToLower(xname))
		sendJsonRsp(w, http.StatusBadRequest, "JSON payload xname != request URL xname")
		return
	}

	// Check if the component exists in the DB.

	cmpPtr, err := datastore.GetXname(xname)
	if err != nil {
		log.Println("ERROR getting component from DB:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}
	if cmpPtr != nil {
		cmp = *cmpPtr
	} else {
		cmp = jdata //DB object doesn't exist, so just use the inbound one
	}

	// Replace the specified data.  Make sure to avoid verboten ones.

	var lxname string
	var ltype sls_common.HMSStringType
	var ltypestr base.HMSType

	lxname = base.VerifyNormalizeCompID(jdata.Parent)
	if lxname == "" {
		log.Printf("ERROR, Invalid parent name '%s'\n", jdata.Parent)
		sendJsonRsp(w, http.StatusBadRequest, "invalid parent xname")
		return
	}
	cmp.Parent = lxname

	// jdata.Type is really a common.HMSStringType
	ltypestr = sls_common.HMSStringTypeToHMSType(jdata.Type)
	if string(ltypestr) == string(sls_common.HMSTypeInvalid) {
		log.Printf("ERROR, Invalid Type field: '%s'\n", string(jdata.Type))
		sendJsonRsp(w, http.StatusBadRequest, "invalid component type")
		return
	}
	cmp.Type = jdata.Type

	// jdata.TypeString is really a base.HMSType
	ltype = sls_common.HMSTypeToHMSStringType(jdata.TypeString)
	if ltype == sls_common.HMSTypeInvalid {
		log.Printf("ERROR, Invalid TypeString: '%s'\n",
			string(jdata.TypeString))
		sendJsonRsp(w, http.StatusBadRequest, "invalid component type string")
		return
	}
	if ltype != jdata.Type {
		log.Printf("ERROR, Mismatched Type and TypeString: '%s'/'%s'\n",
			string(jdata.Type), string(jdata.TypeString))
		sendJsonRsp(w, http.StatusBadRequest, "invalid component type string")
		return
	}
	cmp.TypeString = jdata.TypeString

	if !sls_common.IsCabinetTypeValid(jdata.Class) {
		log.Printf("ERROR, invalid component class '%s'\n",
			string(jdata.Class))
		sendJsonRsp(w, http.StatusBadRequest, "invalid component class")
		return
	}
	cmp.Class = jdata.Class

	if jdata.ExtraPropertiesRaw != nil {
		cmp.ExtraPropertiesRaw = jdata.ExtraPropertiesRaw
	}

	// Write back to the DB

	err = datastore.SetXname(cmp.Xname, cmp)
	if err != nil {
		log.Println("ERROR updating DB:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "DB update failed")
		return
	}

	sendJsonCompRsp(w, cmp)
}

// Recursive function used to get all components of a component
// tree and put them into a linear slice.

func getCompTree(gcomp sls_common.GenericHardware, compList *[]sls_common.GenericHardware) error {
	for _, cxname := range gcomp.Children {
		cmp, err := datastore.GetXname(cxname)
		if cmp == nil {
			log.Printf("WARNING: child component '%s' not found in DB.\n",
				cxname)
			continue
		}
		if err != nil {
			return err
		}
		err = getCompTree(*cmp, compList)
		if err != nil {
			return err
		}
	}

	*compList = append(*compList, gcomp)
	return nil
}

//  /hardware/{xname} DELETE API

func doHardwareObjDelete(w http.ResponseWriter, r *http.Request) {
	var compList []sls_common.GenericHardware

	// Decode the URL to get the XName
	vars := mux.Vars(r)
	xname := base.NormalizeHMSCompID(vars["xname"])

	if !base.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, invalid Xname in request URL: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	// Fetch the item and all of its descendants from the database.  If
	// the item does not exist, error.

	cmp, err := datastore.GetXname(xname)
	if err != nil {
		log.Println("ERROR, error in DB query:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}
	if cmp == nil {
		log.Printf("ERROR, no '%s' component in DB.\n", xname)
		sendJsonRsp(w, http.StatusNotFound, "no such component not in DB")
		return
	}

	err = getCompTree(*cmp, &compList)
	if err != nil {
		log.Println("ERROR, error in comp tree DB query:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}

	// Delete the item(s) from the database

	ok := true
	for _, component := range compList {
		log.Printf("INFO: Deleting: '%s'\n", component.Xname)
		err = datastore.DeleteXname(component.Xname)
		if err != nil {
			ok = false
		}
	}

	if !ok {
		sendJsonRsp(w, http.StatusInternalServerError, "failed to delete entry in DB")
		return
	}

	sendJsonRsp(w, http.StatusOK, "deleted entry and its descendants")
}

//  /search/hardware GET API

func doHardwareSearch(w http.ResponseWriter, r *http.Request) {
	hardware := sls_common.GenericHardware{
		Parent:             r.FormValue("parent"),
		Children:           nil,
		Xname:              r.FormValue("xname"),
		Type:               sls_common.HMSStringType(r.FormValue("type")),
		Class:              sls_common.CabinetType(r.FormValue("class")),
		ExtraPropertiesRaw: nil,
	}

	// Build up the extra properties section by gathering the various possible query object and adding them.
	properties := make(map[string]interface{})

	powerConnector := r.FormValue("power_connector")
	if powerConnector != "" {
		properties["PoweredBy"] = powerConnector
	}

	object := r.FormValue("object")
	if object != "" {
		properties["Object"] = object
	}

	nodeNics := r.FormValue("node_nics")
	if nodeNics != "" {
		properties["NodeNics"] = []string{nodeNics}
	}

	networks := r.FormValue("networks")
	if networks != "" {
		properties["Networks"] = []string{networks}
	}

	peers := r.FormValue("peers")
	if peers != "" {
		properties["Peers"] = []string{peers}
	}

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

	hardware.ExtraPropertiesRaw = properties

	returnedNetworks, err := datastore.SearchGenericHardware(hardware)
	if err == database.NoSuch {
		log.Println("ERROR: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Not Found",
			"Hardware not found in DB",
			r.URL.Path, http.StatusNotFound)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if err != nil {
		log.Println("ERROR: Failed to search for hardware:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to search hardware in DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(returnedNetworks)
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
