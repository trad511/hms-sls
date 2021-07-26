// MIT License
//
// (C) Copyright [2021] Hewlett Packard Enterprise Development LP
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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	base "github.com/Cray-HPE/hms-base"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

const (
	loadStateURL    = "http://localhost:8080/v1/loadstate"
	hwSearchURLBase = "http://localhost:8376/v1/search/hardware"
)

func loadTestSLSState() error {
	// Load in the sls_input_file.json
	slsState, err := ioutil.ReadFile("testdata/sls_input_file.json")
	if err != nil {
		return err
	}
	slsStateReader := bytes.NewReader(slsState)

	// Create payload
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fw, err := writer.CreateFormFile("sls_dump", "sls_test_config.json")
	if err != nil {
		return err
	}

	_, err = io.Copy(fw, slsStateReader)
	if err != nil {
		return err
	}
	writer.Close()

	// Create HTTP Request
	req, err := http.NewRequest("POST", loadStateURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", response.Code)
	}

	return nil
}

type HardwareSearchTestSuite struct {
	suite.Suite
}

func (suite *HardwareSearchTestSuite) SetupSuite() {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}

	dbInit()
	hwDBClear()

	err := loadTestSLSState()
	suite.NoError(err)
}

func (suite *HardwareSearchTestSuite) doSearch(searchURL string, expectedStatus int) ([]sls_common.GenericHardware, *base.ProblemDetails) {
	req, reqerr := http.NewRequest(http.MethodGet, searchURL, nil)
	suite.NoError(reqerr)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	suite.T().Logf("Response: %d", response.Code)
	suite.Equal(expectedStatus, response.Code)

	if response.Code == http.StatusOK {
		// On success we get a collection of hardware
		returnedHardware := []sls_common.GenericHardware{}
		err := json.Unmarshal(response.Body.Bytes(), &returnedHardware)
		suite.NoError(err, "Response: %s", response.Body.String())

		return returnedHardware, nil
	}

	// On failure we get a Problem detail (hopefully)
	pd := base.ProblemDetails{}
	err := json.Unmarshal(response.Body.Bytes(), &pd)
	suite.NoError(err, "Response: %s", response.Body.String())

	return nil, &pd
}

func (suite *HardwareSearchTestSuite) verifyReturnedHardware(returnedHardware []sls_common.GenericHardware, expectedXnames []string) {
	suite.Len(returnedHardware, len(expectedXnames))

	returnedXnames := []string{}
	for _, hardware := range returnedHardware {
		returnedXnames = append(returnedXnames, hardware.Xname)
	}

	suite.T().Log("Expected xnames:", expectedXnames)
	suite.T().Log("Returned xnames:", returnedXnames)

	suite.ElementsMatch(expectedXnames, returnedXnames)
}

func (suite *HardwareSearchTestSuite) TestSearchValidTypes() {
	// Verify the hardware search endpoint accepts the following SLS types via the type query param
	slsTypes := []string{
		"comptype_cdu",                   // dD
		"comptype_cdu_mgmt_switch",       // dDwW
		"comptype_cab_cdu",               // xXdD
		"comptype_cabinet",               // xX
		"comptype_cab_pdu_controller",    // xXmM
		"comptype_cab_pdu",               // xXmMpP
		"comptype_cab_pdu_nic",           // xXmMiI
		"comptype_cab_pdu_outlet",        // xXmMpPjJ DEPRECATED
		"comptype_cab_pdu_pwr_connector", // xXmMpPvV

		"comptype_chassis",                 // xXcC
		"comptype_chassis_bmc",             // xXcCbB
		"comptype_cmm_rectifier",           // xXcCtT
		"comptype_cmm_fpga",                // xXcCfF
		"comptype_cec",                     // xXeE
		"comptype_compmod",                 // xXcCsS
		"comptype_rtrmod",                  // xXcCrR
		"comptype_ncard",                   // xXcCsSbB
		"comptype_bmc_nic",                 // xXcCsSbBiI
		"comptype_node_enclosure",          // xXcCsSeE
		"comptype_compmod_power_connector", // xXcCsSvV
		"comptype_node",                    // xXcCsSbBnN
		"comptype_node_processor",          // xXcCsSbBnNpP
		"comptype_node_nic",                // xXcCsSbBnNiI
		"comptype_node_hsn_nic",            // xXcCsSbBnNhH
		"comptype_dimm",                    // xXcCsSbBnNdD
		"comptype_node_accel",              // xXcCsSbBnNaA
		"comptype_node_fpga",               // xXcCsSbBfF
		"comptype_hsn_asic",                // xXcCrRaA
		"comptype_rtr_fpga",                // xXcCrRfF
		"comptype_rtr_tor_fpga",            // xXcCrRtTfF
		"comptype_rtr_bmc",                 // xXcCrRbB
		"comptype_rtr_bmc_nic",             // xXcCrRbBiI

		"comptype_hsn_board",             // xXcCrReE
		"comptype_hsn_link",              // xXcCrRaAlL
		"comptype_hsn_connector",         // xXcCrRjJ
		"comptype_hsn_connector_port",    // xXcCrRjJpP
		"comptype_mgmt_switch",           // xXcCwW
		"comptype_mgmt_switch_connector", // xXcCwWjJ
		"comptype_hl_switch",             // xXcChHsS

		// Special types and wildcards
		"comptype_ncn_box", // smsN
		"any",              // s0
	}

	for _, slsType := range slsTypes {
		searchURL := fmt.Sprintf("%s?type=%s", hwSearchURLBase, slsType)
		suite.T().Logf("Search URL: %s", searchURL)

		req, reqerr := http.NewRequest(http.MethodGet, searchURL, nil)
		suite.NoError(reqerr)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		suite.T().Logf("Response: %d", response.Code)

		suite.Equal(http.StatusOK, response.Code, "Bad response in GET op")
	}
}

func (suite *HardwareSearchTestSuite) TestSearchSpecialTypes() {
	// Verify the hardware search endpoint rejects the following SLS types via the type query param
	slsTypes := []string{
		// Special types and wildcards
		"comptype_partition", // pH.S
		"comptype_all",       // all
		"comptype_all_comp",  // all_comp
		"comptype_all_svc",   // all_svc
		"INVALID",            // Not a valid type/xname
	}

	for _, slsType := range slsTypes {
		searchURL := fmt.Sprintf("%s?type=%s", hwSearchURLBase, slsType)
		suite.T().Logf("Search URL: %s", searchURL)

		req, reqerr := http.NewRequest(http.MethodGet, searchURL, nil)
		suite.NoError(reqerr)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		suite.T().Logf("Response: %d", response.Code)

		// TODO CASMHMS-4671 - This feels like it should be a 400 Bad Request instead
		suite.Equal(http.StatusInternalServerError, response.Code, "Bad response in GET op")
	}
}

func (suite *HardwareSearchTestSuite) TestSearchNodeNIC() {
	tests := []struct {
		nodeNIC string
		xnames  []string
	}{{
		nodeNIC: "x3000c0s2b0",
		xnames:  []string{"x3000c0w31j29"},
	}, {
		nodeNIC: "x3000c0s3b0",
		xnames:  []string{"x3000c0w32j30"},
	}, {
		nodeNIC: "x3000c0s4b0",
		xnames:  []string{"x3000c0w31j30"},
	}, {
		nodeNIC: "x3000c0s5b0",
		xnames:  []string{"x3000c0w31j31"},
	}, {
		nodeNIC: "x1000c0s2b0",
	}, {
		nodeNIC: "d0w1",
	}, {
		nodeNIC: "foo",
	}, {
		nodeNIC: "x3000c0w31j30",
	}}

	for _, test := range tests {
		searchURL := fmt.Sprintf("%s?node_nics=%s", hwSearchURLBase, test.nodeNIC)
		suite.T().Logf("Search URL: %s", searchURL)

		returnedHardware, pd := suite.doSearch(searchURL, http.StatusOK)
		suite.Nil(pd)

		suite.verifyReturnedHardware(returnedHardware, test.xnames)
	}
}

func (suite *HardwareSearchTestSuite) TestSearchParent() {
	// Peers are for NICs (BMC/Node), not a to find the siblings of a xnames
	tests := []struct {
		parent string
		xnames []string
	}{{
		parent: "d0",
		xnames: []string{"d0w1", "d0w2"},
	}, {
		parent: "s0",
		xnames: []string{"x3000", "x1000"},
	}, {
		parent: "x3000c0s1b0",
		xnames: []string{"x3000c0s1b0n0"},
	}, {
		parent: "x3000c0w32",
		xnames: []string{
			"x3000c0w32j48",
			"x3000c0w32j32",
			"x3000c0w32j47",
			"x3000c0w32j31",
			"x3000c0w32j34",
			"x3000c0w32j27",
			"x3000c0w32j26",
			"x3000c0w32j43",
			"x3000c0w32j41",
			"x3000c0w32j42",
			"x3000c0w32j33",
			"x3000c0w32j28",
			"x3000c0w32j30",
			"x3000c0w32j25",
		},
	}, {
		parent: "x3000c0s1b0n0",
		xnames: []string{},
	}}

	for _, test := range tests {
		searchURL := fmt.Sprintf("%s?parent=%s", hwSearchURLBase, test.parent)
		suite.T().Logf("Search URL: %s", searchURL)

		returnedHardware, pd := suite.doSearch(searchURL, http.StatusOK)
		suite.Nil(pd)

		suite.verifyReturnedHardware(returnedHardware, test.xnames)
	}
}

func (suite *HardwareSearchTestSuite) TestSearchXname() {
	// Peers are for NICs (BMC/Node), not a to find the siblings of a xnames
	tests := []struct {
		parent         string
		xnames         []string
		expectedStatus int
	}{{
		parent:         "d0w1",
		xnames:         []string{"d0w1"},
		expectedStatus: http.StatusOK,
	}, {
		parent:         "x3000c0s1b0n0",
		xnames:         []string{"x3000c0s1b0n0"},
		expectedStatus: http.StatusOK,
	}, {
		parent:         "x2000c0s1b0n0",
		xnames:         []string{},
		expectedStatus: http.StatusOK,
	}, {
		parent:         "foo",
		xnames:         []string{},
		expectedStatus: http.StatusInternalServerError, // TODO CASMHMS-4671 This should probably get a 400 status code
	}}

	for _, test := range tests {
		searchURL := fmt.Sprintf("%s?xname=%s", hwSearchURLBase, test.parent)
		suite.T().Logf("Search URL: %s", searchURL)

		returnedHardware, pd := suite.doSearch(searchURL, test.expectedStatus)
		if test.expectedStatus == http.StatusOK {
			suite.Nil(pd)
		} else {
			suite.NotNil(pd)
		}

		suite.verifyReturnedHardware(returnedHardware, test.xnames)
	}
}

func (suite *HardwareSearchTestSuite) TestSearchExtraProperties() {
	// TODO /v1/search/hardware does not support searching for multiple node NICs due to
	// 	// getting the form value from r.FormValue which only returns 1 value. The extra properties
	// 	// support multiple values as it parses the form values differently.

	tests := []struct {
		propertyName   string
		propertyValues []string
		expectedXnames []string
	}{{
		propertyName:   "Brand",
		propertyValues: []string{"Aruba"},
		expectedXnames: []string{
			"d0w1",
			"d0w2",
			"x3000c0h33s1",
			"x3000c0h34s1",
			"x3000c0h35s1",
			"x3000c0h36s1",
			"x3000c0h37s1",
			"x3000c0h38s1",
			"x3000c0w31",
			"x3000c0w32",
		},
	}, {
		propertyName:   "NID",
		propertyValues: []string{"1011"},
		expectedXnames: []string{
			"x1000c0s2b1n1",
		},
	}, {
		propertyName:   "Role",
		propertyValues: []string{"Management"},
		expectedXnames: []string{
			"x3000c0s1b0n0",
			"x3000c0s2b0n0",
			"x3000c0s3b0n0",
			"x3000c0s4b0n0",
			"x3000c0s5b0n0",
			"x3000c0s6b0n0",
			"x3000c0s7b0n0",
			"x3000c0s8b0n0",
			"x3000c0s9b0n0",
			"x3000c0s10b0n0",
			"x3000c0s11b0n0",
			"x3000c0s12b0n0",
		},
	}, {
		propertyName:   "SubRole",
		propertyValues: []string{"Storage"},
		expectedXnames: []string{
			"x3000c0s10b0n0",
			"x3000c0s11b0n0",
			"x3000c0s12b0n0",
		},
	}, {
		propertyName:   "SubRole",
		propertyValues: []string{"Master", "Storage"},
		expectedXnames: []string{
			// Master nodes
			"x3000c0s1b0n0",
			"x3000c0s2b0n0",
			"x3000c0s3b0n0",

			// Storage nodes
			"x3000c0s10b0n0",
			"x3000c0s11b0n0",
			"x3000c0s12b0n0",
		},
	}, {
		propertyName:   "VendorName",
		propertyValues: []string{"1/1/37"},
		expectedXnames: []string{
			"x3000c0w31j37",
		},
	}, {
		propertyName:   "VendorName",
		propertyValues: []string{"1/1/1"},
		expectedXnames: []string{},
	}}

	for _, test := range tests {
		searchTokens := []string{}
		for _, value := range test.propertyValues {
			token := fmt.Sprintf("extra_properties.%s=%s", test.propertyName, value)
			searchTokens = append(searchTokens, token)
		}

		searchURL := hwSearchURLBase + "?" + strings.Join(searchTokens, "&")
		suite.T().Logf("Search URL: %s", searchURL)

		returnedHardware, pd := suite.doSearch(searchURL, http.StatusOK)
		suite.Nil(pd)

		suite.verifyReturnedHardware(returnedHardware, test.expectedXnames)
	}
}

func TestHardwareSearchTestSuite(t *testing.T) {
	suite.Run(t, new(HardwareSearchTestSuite))
}
