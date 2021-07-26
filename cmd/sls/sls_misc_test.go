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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"

	"github.com/Cray-HPE/hms-sls/internal/database"

	"github.com/gorilla/mux"
	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-sls/internal/datastore"
)

var glbRouter *mux.Router

func setupInit(t *testing.T) error {
	//Hook up the API routes
	if glbRouter == nil {
		routes := generateRoutes()
		glbRouter = newRouter(routes)
		datastore.ConfigureStorage("postgres", "", []string{})
	}
	return nil
}

//This just does a POST to create a component.

func updateDBVersion() error {
	req_payload := bytes.NewBufferString(`{"Parent":"x100c0s1b0","Xname":"x100c0s1b0n0","Type":"comptype_node","TypeString":"Node","Class":"River","ExtraProperties":{"NID":5555,"Role":"Management"}}`)
	req, rerr := http.NewRequest("POST", "http://localhost:8080"+API_HARDWARE,
		req_payload)
	if rerr != nil {
		return rerr
	}

	pw := httptest.NewRecorder()
	glbRouter.ServeHTTP(pw, req)

	if pw.Code != http.StatusOK {
		return fmt.Errorf("ERROR response in POST operation: %d/%s",
			pw.Code, http.StatusText(pw.Code))
	}

	return nil
}

func TestDoReadinessGet(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	// set up env vars
	os.Setenv("SLS_HTTP_LISTEN_ADDR", "8376")
	os.Setenv("SLS_DEBUG", "0")
	envVars()

	t.Log("==> Checking Initial /readiness API <==")
	req, rerr := http.NewRequest("GET", "http://localhost:8080"+API_READINESS, nil)
	if rerr != nil {
		t.Error("ERROR setting up /readiness request:", rerr)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doReadinessGet)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("ERROR in /readiness GET request, bad status: %d\n", rr.Code)
	}

	//Check for bad verb
	t.Log("==> Checking Bad HTTP Verb Usage <==")
	req2, rerr2 := http.NewRequest("POST", "http://localhost:8080"+API_READINESS, nil)
	if rerr2 != nil {
		t.Error("ERROR setting up /readiness POST request:", rerr2)
	}
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusMethodNotAllowed {
		t.Error("ERROR in /readiness POST request should have failed.")
	}
}

func TestDoLivenessGet(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	// set up env vars
	os.Setenv("SLS_HTTP_LISTEN_ADDR", "8376")
	os.Setenv("SLS_DEBUG", "0")
	envVars()

	t.Log("==> Checking Initial /liveness API <==")
	req, rerr := http.NewRequest("GET", "http://localhost:8080"+API_LIVENESS, nil)
	if rerr != nil {
		t.Error("ERROR setting up /liveness request:", rerr)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doLivenessGet)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("ERROR in /liveness GET request, bad status: %d\n", rr.Code)
	}

	//Check for bad verb
	t.Log("==> Checking Bad HTTP Verb Usage <==")
	req2, rerr2 := http.NewRequest("POST", "http://localhost:8080"+API_LIVENESS, nil)
	if rerr2 != nil {
		t.Error("ERROR setting up /liveness POST request:", rerr2)
	}
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusMethodNotAllowed {
		t.Error("ERROR in /liveness POST request should have failed.")
	}
}

func TestDoHealthGet(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	// set up env vars
	os.Setenv("SLS_HTTP_LISTEN_ADDR", "8376")
	os.Setenv("SLS_DEBUG", "0")
	envVars()

	t.Log("==> Checking Initial /health API <==")
	req, rerr := http.NewRequest("GET", "http://localhost:8080"+API_LIVENESS, nil)
	if rerr != nil {
		t.Error("ERROR setting up /health request:", rerr)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doHealthGet)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("ERROR in /health GET request, bad status: %d\n", rr.Code)
	}

	body, berr := ioutil.ReadAll(rr.Body)
	if berr != nil {
		t.Error("ERROR reading /health GET response body:", berr)
	}

	var jdata HealthResponse
	jerr := json.Unmarshal(body, &jdata)
	if jerr != nil {
		t.Error("ERROR unmarshaling /health GET data:", jerr)
	}

	//Initial should be "Not enabled", "Ready"
	expVal := HealthResponse{Vault: "Not enabled", DBConnection: "Ready"}
	if jdata != expVal {
		t.Errorf("ERROR, mismatch in initial /health data, exp:\n%v\ngot:\n%v\n",
			expVal, jdata)
	}

	//Check for bad verb
	t.Log("==> Checking Bad HTTP Verb Usage <==")
	req2, rerr2 := http.NewRequest("POST", "http://localhost:8080"+API_LIVENESS, nil)
	if rerr2 != nil {
		t.Error("ERROR setting up /health POST request:", rerr2)
	}
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusMethodNotAllowed {
		t.Error("ERROR in /health POST request should have failed.")
	}
}

func TestDoVersionGet(t *testing.T) {
	var jdata sls_common.SLSVersion

	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	req_payload := bytes.NewBufferString("")

	//Update the DB version

	t.Log("==> Updating DB <==")
	verr := updateDBVersion()
	if verr != nil {
		t.Error("ERROR, can't update DB version:", verr)
	}

	//GET version, version should have changed

	t.Log("==> Checking New /version Value <==")
	req_payload2 := bytes.NewBufferString("")
	req2, rerr2 := http.NewRequest("GET", "http://localhost:8080"+API_VERSION,
		req_payload2)
	if rerr2 != nil {
		t.Error("ERROR setting up /version request:", rerr2)
	}
	rr2 := httptest.NewRecorder()
	handler2 := http.HandlerFunc(doVersionGet)

	handler2.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("ERROR in /version GET request, bad status: %d\n", rr2.Code)
	}

	body, berr := ioutil.ReadAll(rr2.Body)
	if berr != nil {
		t.Error("ERROR reading /version GET response body:", berr)
	}

	jerr := json.Unmarshal(body, &jdata)
	if jerr != nil {
		t.Error("ERROR unmarshaling /version GET data:", jerr)
	}

	//Should be 1, current time.  Can't guarantee a perfect match,
	//so we'll decode the time and check at least the year.

	toks := strings.Split(jdata.LastUpdated, "-")
	actYear := toks[0]
	expYear := fmt.Sprintf("%d", time.Now().Year())
	if actYear != expYear {
		t.Errorf("ERROR in version year compare, exp '%s', got '%s'\n",
			expYear, actYear)
	}

	//Check for bad verb

	t.Log("==> Checking Bad HTTP Verb Usage <==")
	req3, rerr3 := http.NewRequest("POST", "http://localhost:8080"+API_VERSION,
		req_payload)
	if rerr3 != nil {
		t.Error("ERROR setting up /version POST request:", rerr3)
	}
	rr3 := httptest.NewRecorder()
	handlerBV := http.HandlerFunc(doVersionGet)
	handlerBV.ServeHTTP(rr3, req3)
	if rr3.Code == http.StatusOK {
		t.Error("ERROR in /version POST request should have failed.")
	}
}

func TestDoLoadstateWithKey(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	// Preload the database with some data so after we make the request we can make sure it's gone
	sampleObj := sls_common.GenericHardware{"x0", []string{}, "x0c0", sls_common.Chassis, sls_common.ClassRiver, base.Chassis, 0, "2014-07-16 20:55:46 +0000 UTC", nil, nil}
	datastore.SetXname(sampleObj.Xname, sampleObj)
	sampleNw := sls_common.Network{"DUMMY", "Sample dummy network", []string{}, sls_common.NetworkTypeEthernet, 0, "2014-07-16 20:55:46 +0000 UTC", nil}
	datastore.SetNetwork(sampleNw)

	// Build the multipart form files necessary for the public key and dump.
	const privateKeyPEM = `
-----BEGIN PRIVATE KEY-----
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
OrdRstweaH84Wxjriy/PGWbxUxgPRu0=
-----END PRIVATE KEY-----
`
	privateKeyPEMReader := strings.NewReader(privateKeyPEM)

	const slsDump = `
{
  "Hardware": {
    "x0c0s0b0": {
      "Parent": "x0c0s0",
      "Xname": "x0c0s0b0",
      "Type": "comptype_ncard",
      "Class": "River",
      "TypeString": "NodeBMC",
      "ExtraProperties": {
        "IP4addr": "10.1.1.1",
        "IP6addr": "DHCPv6"
      },
      "VaultData": "lSuvAWeK5Gh6o7DIpWIMfegh85NvyAbqBk41YRIATi9JZQJcgf5H+rna6xGmHMwNu3/z4/ClP8UAkfWuH6xWsvux+XzddjrOdziWD8SNOF4MvxUz5AVHaDqG6E/JoHbGU9kjEHfxD5Hqr2SeqY0j4mXMjXWrvh/TmidXCpnfJ4CFA4IgBjnSR5Tbs/8k0vACo2JTOo9+g7NBK6nnZhOqcQ69UzBAHQkdhlazMqdk82MJv3ygC5Vt5lR9asy5TRHiTucmu49x8AVoiea5OCxS8E/KWcimXvPw07H3RPdf/0gHLNh1ztr6wUuwLOBTs6DYIaTGgab/YckregsKOrnpYkPbpzsZsIa4TULjMpZjLXUohc+J94dFWazTIeD8JGBpMLS39QFM2zZCuKJHI2vS5KYLSjw4iqvr670vEWaAbekxdZkwzJe+7loN/BRY1l059i+ES2OJ92t6yzLJVvXxISlKxORZW7VtZbBQh0aXsApsGW826SY8sNo2M3mjU/oJgY7mO5eom0ckrV+EWK2sIpZ0P9GSO3j+XGAs3M7RyEfF9sIQ9X6JcEhTxBLGrBEyZ8RtIBUVDKphvnfF8pJlvH9s5ZYT9ldxAjy2GP3oOxyklxtiRgNT/xk8oV04LB8SdVB+kS3IGY1vntJ+rlrii3wRz8r4bJ0yT/qFJ7US7PU="
    },
    "x1000c3": {
      "Parent": "x1000",
      "Children": [
        "x1000c3s2"
      ],
      "Xname": "x1000c3",
      "Type": "comptype_chassis",
      "Class": "",
      "TypeString": "Chassis"
    },
    "x1000c3s2": {
      "Parent": "x1000c3",
      "Xname": "x1000c3s2",
      "Type": "comptype_compmod",
      "Class": "",
      "TypeString": "ComputeModule"
    }
  },
  "Networks": {
    "HSN": {
      "Name": "HSN",
      "FullName": "High Speed Network",
      "IPRanges": [
        "192.168.1.0/28",
        "192.168.2.0/28"
      ],
      "Type": "slingshot10"
    },
    "NMN": {
      "Name": "NMN",
      "FullName": "Node Management Network",
      "IPRanges": [
        "10.100.1.0/28",
        "10.100.2.0/28"
      ],
      "Type": "ethernet"
    }
  }
}
`
	slsDumpReader := strings.NewReader(slsDump)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fw, err := writer.CreateFormFile("sls_dump", "sls_test_config.json")
	if err != nil {
		t.Error("Failed to create form file for dump:", err)
	}
	_, err = io.Copy(fw, slsDumpReader)
	if err != nil {
		t.Error("Failed to copy form file for dump:", err)
	}

	fw, err = writer.CreateFormFile("private_key", "private_key.pem")
	if err != nil {
		t.Error("Failed to create form file for private key:", err)
	}
	_, err = io.Copy(fw, privateKeyPEMReader)
	if err != nil {
		t.Error("Failed to copy form file for private key:", err)
	}

	writer.Close()

	t.Log("Making request to /loadstate")
	req, rerr := http.NewRequest("POST", "http://localhost:8080"+API_LOADSTATE, &buf)
	if rerr != nil {
		t.Error("ERROR setting up /loadstate request:", rerr)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doLoadState)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("ERROR in /loadstate GET request, bad status: %d\n", rr.Code)
	}

	// Now validate that the initial data isn't in there...
	r, err := datastore.GetXname("x0c0")
	if err != nil {
		t.Errorf("Error retrieving old data: %s", err)
	}
	if r != nil {
		t.Errorf("Old data remains in the database!")
	}

	// And that the new data went in
	currentXnames, err := datastore.GetAllXnames()
	if err != nil {
		t.Errorf("Error retrieving database contents: %s", err)
	}
	if len(currentXnames) != 3 {
		t.Errorf("Datastore has the wrong number of xnames: %d", len(currentXnames))
	}

	for _, r2 := range currentXnames {
		if r2 != "x1000c3" && r2 != "x1000c3s2" && r2 != "x0c0s0b0" {
			t.Errorf("Unexpected xname in results: %s", r2)
		}
	}

	currentNetworks, err := datastore.GetAllNetworks()
	if err != nil {
		t.Errorf("Error retrieving database network contents: %s", err)
	}
	if len(currentNetworks) != 2 {
		t.Errorf("Datastore has the wrong number of networks: %d", len(currentNetworks))
	}
}

func TestDoLoadstateWithoutKey(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	// Preload the database with some data so after we make the request we can make sure it's gone
	sampleObj := sls_common.GenericHardware{"x0", []string{}, "x0c0", sls_common.Chassis, sls_common.ClassRiver, base.Chassis, 0, "2014-07-16 20:55:46 +0000 UTC", nil, nil}
	datastore.SetXname(sampleObj.Xname, sampleObj)
	sampleNw := sls_common.Network{"DUMMY", "Sample dummy network", []string{}, sls_common.NetworkTypeEthernet, 0, "2014-07-16 20:55:46 +0000 UTC", nil}
	datastore.SetNetwork(sampleNw)

	const slsDump = `
{
  "Hardware": {
    "x0c0s0b0": {
      "Parent": "x0c0s0",
      "Xname": "x0c0s0b0",
      "Type": "comptype_ncard",
      "Class": "River",
      "TypeString": "NodeBMC",
      "ExtraProperties": {
        "IP4addr": "10.1.1.1",
        "IP6addr": "DHCPv6"
      },
      "VaultData": "lSuvAWeK5Gh6o7DIpWIMfegh85NvyAbqBk41YRIATi9JZQJcgf5H+rna6xGmHMwNu3/z4/ClP8UAkfWuH6xWsvux+XzddjrOdziWD8SNOF4MvxUz5AVHaDqG6E/JoHbGU9kjEHfxD5Hqr2SeqY0j4mXMjXWrvh/TmidXCpnfJ4CFA4IgBjnSR5Tbs/8k0vACo2JTOo9+g7NBK6nnZhOqcQ69UzBAHQkdhlazMqdk82MJv3ygC5Vt5lR9asy5TRHiTucmu49x8AVoiea5OCxS8E/KWcimXvPw07H3RPdf/0gHLNh1ztr6wUuwLOBTs6DYIaTGgab/YckregsKOrnpYkPbpzsZsIa4TULjMpZjLXUohc+J94dFWazTIeD8JGBpMLS39QFM2zZCuKJHI2vS5KYLSjw4iqvr670vEWaAbekxdZkwzJe+7loN/BRY1l059i+ES2OJ92t6yzLJVvXxISlKxORZW7VtZbBQh0aXsApsGW826SY8sNo2M3mjU/oJgY7mO5eom0ckrV+EWK2sIpZ0P9GSO3j+XGAs3M7RyEfF9sIQ9X6JcEhTxBLGrBEyZ8RtIBUVDKphvnfF8pJlvH9s5ZYT9ldxAjy2GP3oOxyklxtiRgNT/xk8oV04LB8SdVB+kS3IGY1vntJ+rlrii3wRz8r4bJ0yT/qFJ7US7PU="
    },
    "x1000c3": {
      "Parent": "x1000",
      "Children": [
        "x1000c3s2"
      ],
      "Xname": "x1000c3",
      "Type": "comptype_chassis",
      "Class": "",
      "TypeString": "Chassis"
    },
    "x1000c3s2": {
      "Parent": "x1000c3",
      "Xname": "x1000c3s2",
      "Type": "comptype_compmod",
      "Class": "",
      "TypeString": "ComputeModule"
    }
  },
  "Networks": {
    "HSN": {
      "Name": "HSN",
      "FullName": "High Speed Network",
      "IPRanges": [
        "192.168.1.0/28",
        "192.168.2.0/28"
      ],
      "Type": "slingshot10"
    },
    "NMN": {
      "Name": "NMN",
      "FullName": "Node Management Network",
      "IPRanges": [
        "10.100.1.0/28",
        "10.100.2.0/28"
      ],
      "Type": "ethernet"
    }
  }
}
`
	slsDumpReader := strings.NewReader(slsDump)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fw, err := writer.CreateFormFile("sls_dump", "sls_test_config.json")
	if err != nil {
		t.Error("Failed to create form file for dump:", err)
	}
	_, err = io.Copy(fw, slsDumpReader)
	if err != nil {
		t.Error("Failed to copy form file for dump:", err)
	}

	writer.Close()

	t.Log("Making request to /loadstate")
	req, rerr := http.NewRequest("POST", "http://localhost:8080"+API_LOADSTATE, &buf)
	if rerr != nil {
		t.Error("ERROR setting up /loadstate request:", rerr)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doLoadState)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("ERROR in /loadstate GET request, bad status: %d\n", rr.Code)
	}

	// Now validate that the initial data isn't in there...
	r, err := datastore.GetXname("x0c0")
	if err != nil {
		t.Errorf("Error retrieving old data: %s", err)
	}
	if r != nil {
		t.Errorf("Old data remains in the database!")
	}

	// And that the new data went in
	currentXnames, err := datastore.GetAllXnames()
	if err != nil {
		t.Errorf("Error retrieving database contents: %s", err)
	}
	if len(currentXnames) != 3 {
		t.Errorf("Datastore has the wrong number of xnames: %d", len(currentXnames))
	}

	for _, r2 := range currentXnames {
		if r2 != "x1000c3" && r2 != "x1000c3s2" && r2 != "x0c0s0b0" {
			t.Errorf("Unexpected xname in results: %s", r2)
		}
	}

	currentNetworks, err := datastore.GetAllNetworks()
	if err != nil {
		t.Errorf("Error retrieving database network contents: %s", err)
	}
	if len(currentNetworks) != 2 {
		t.Errorf("Datastore has the wrong number of networks: %d", len(currentNetworks))
	}
}

func TestDoDumpstate(t *testing.T) {
	kerr := setupInit(t)
	if kerr != nil {
		t.Error("Error with test setup:", kerr)
	}

	err := database.DeleteAllGenericHardware()
	if err != nil {
		t.Errorf("Error deleting all hardware: %s", err)
	}

	inputObjs := make(map[string]sls_common.GenericHardware, 0)
	inputObjs["x1000c3"] = sls_common.GenericHardware{
		Parent:     "x1000",
		Xname:      "x1000c3",
		Type:       sls_common.Chassis,
		TypeString: base.Chassis,
		Children:   []string{},
	}
	inputObjs["x1000c3c2"] = sls_common.GenericHardware{
		Parent:     "x1000c3",
		Xname:      "x1000c3s2",
		Type:       sls_common.ComputeModule,
		TypeString: base.ComputeModule,
		Children:   []string{},
	}

	for _, obj := range inputObjs {
		t.Logf("Inserting test data for %s: %v", obj.Xname, obj)
		err = datastore.SetXname(obj.Xname, obj)
		if err != nil {
			t.Fatalf("Failed ot insert %s: %s", obj.Xname, err)
		}
	}

	err = database.DeleteAllNetworks()
	if err != nil {
		t.Fatalf("Error deleting all networks: %s", err)
	}

	sampleNw := sls_common.Network{
		"HSN",
		"Rosetta High Speed Network",
		[]string{"10.1.1.0/24", "10.1.2.0/24"},
		"Slingshot10",
		0,
		"2014-07-16 20:55:46 +0000 UTC",
		nil,
	}
	err = datastore.SetNetwork(sampleNw)
	if err != nil {
		t.Fatalf("Failed to set network: %s", err)
	}

	t.Log("Making request to /dumpstate")
	req_payload := bytes.NewBufferString("")
	req, rerr := http.NewRequest("GET", "http://localhost:8080"+API_DUMPSTATE,
		req_payload)
	if rerr != nil {
		t.Error("ERROR setting up /dumpstate request:", rerr)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doDumpState)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("ERROR in /dumpstate GET request, bad status: %d\n", rr.Code)
	}

	body, berr := ioutil.ReadAll(rr.Body)
	if berr != nil {
		t.Error("ERROR reading /loadstate GET response body:", berr)
	}

	t.Log("Unmarshalling JSON result")
	result := new(sls_common.SLSState)
	jerr := json.Unmarshal(body, result)
	if jerr != nil {
		t.Error("ERROR unmarshaling /loadstate GET data:", jerr)
	}

	t.Log("Starting to check results")
	t.Logf("Result is: %v", result)
	if len((*result).Hardware) != 2 {
		t.Errorf("Result is the wrong length; expected 2, got %d", len((*result).Hardware))
	}

	expXnames := []string{"x1000c3", "x1000c3s2"}
	for _, name := range expXnames {
		_, ok := (*result).Hardware[name]
		if !ok {
			t.Errorf("Missing expected xname %s!", name)
		}
	}

	if len((*result).Networks) != 1 {
		t.Errorf("Wrong number of networks.  Expected 1, got %d", len((*result).Networks))
	}

	nwNames := []string{"HSN"}
	for _, name := range nwNames {
		_, ok := (*result).Networks[name]
		if !ok {
			t.Errorf("Missing expected  network name %s!", name)
		}
	}
}
