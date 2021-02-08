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

package datastore

import (
	"log"
	"reflect"
	"testing"

	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"

	"github.com/stretchr/testify/suite"

	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-sls/internal/database"
)

type DatastoreTestSuite struct {
	suite.Suite
}

func TestGenericHardwareSuite(t *testing.T) {
	suite.Run(t, new(DatastoreTestSuite))
}

func (suite *DatastoreTestSuite) SetupSuite() {
	err := database.NewDatabase()
	if err != nil {
		suite.FailNowf("Unable create database", "err: %s", err)
	}
}

func (suite *DatastoreTestSuite) TestMakeKeyXname_noPrefix() {
	got := makeKeyXname("x0c0")
	exp := xnameKeyPrefix + "x0c0"
	if got != exp {
		suite.FailNowf("Got did not equal expected", "Got: %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestMakeKeyXname_Prefix() {
	got := makeKeyXname(xnameKeyPrefix + "x0c0")
	exp := xnameKeyPrefix + "x0c0"
	if got != exp {
		suite.FailNowf("Got did not equal expected", "Got: %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetParent() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetParent()
	exp := "x0"
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetChildren() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetChildren()
	exp := []string{"x0c0s1"}
	if !reflect.DeepEqual(got, exp) {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetXname() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetXname()
	exp := "x0c0"
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetType() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetType()
	exp := sls_common.Chassis
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetClass() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetClass()
	exp := sls_common.ClassRiver
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_GetTypeString() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	got := o.GetTypeString()
	exp := base.Chassis
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGenericHardware_JSONRoundTrip() {
	o := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{"x0c0s1"},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassRiver,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	json, err := o.ToJson()
	if err != nil {
		suite.FailNowf("Error making json", "err: %s", err)
	}
	o2 := sls_common.GenericHardware{}
	o2.FromJson(*json)

	if !reflect.DeepEqual(o, o2) {
		suite.FailNowf("JSON roundtrip failed", "Unequal objects. \n\tInitial: %s\n\tFinal: %s", o, o2)
	}
}

func (suite *DatastoreTestSuite) TestGComptypeCompmodPowerConnector_GetParent() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetParent()
	exp := "x0c0s1"
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestComptypeCompmodPowerConnector_GetChildren() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetChildren()
	exp := []string{}
	if !reflect.DeepEqual(got, exp) {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestComptypeCompmodPowerConnector_GetXname() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetXname()
	exp := "x0c0s1v0"
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestComptypeCompmodPowerConnector_GetType() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetType()
	exp := sls_common.NodePowerConnector
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestComptypeCompmodPowerConnector_GetClass() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetClass()
	exp := sls_common.ClassRiver
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestComptypeCompmodPowerConnector_GetTypeString() {
	o := sls_common.GenericHardware{
		Parent:             "x0c0s1",
		Children:           []string{},
		Xname:              "x0c0s1v0",
		Type:               sls_common.NodePowerConnector,
		Class:              sls_common.ClassRiver,
		TypeString:         base.NodePowerConnector,
		ExtraPropertiesRaw: map[string]interface{}{"PoweredBy": "x0m0p0j13"},
	}

	got := o.GetTypeString()
	exp := base.NodePowerConnector
	if got != exp {
		suite.FailNowf("Unexpected return", "Got %s, expected: %s", got, exp)
	}
}

func (suite *DatastoreTestSuite) TestGetXname_okay() {
	robj := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{},
		Xname:              "x0c0",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassMountain,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	err := SetXname(robj.Xname, robj)
	if err != nil {
		suite.FailNowf("Unable to set xname", "err: %s", err)
	}

	res, err := GetXname("x0c0")
	if err != nil {
		suite.FailNowf("Unexpected error fetching data", "err: %s", err)
	}

	if res == nil {
		suite.FailNowf("Was not able to retrieve stored data", "")
	}

	log.Printf("Result: %v", res)

	if res.GetXname() != "x0c0" {
		suite.FailNowf("Retrieved xname not as expected", "Expected x0c0, got %s",
			res.GetXname())
	}
}

func (suite *DatastoreTestSuite) TestSetXname_okay() {
	robj := sls_common.GenericHardware{
		Parent:             "x0",
		Children:           []string{},
		Xname:              "x0c01",
		Type:               sls_common.Chassis,
		Class:              sls_common.ClassMountain,
		TypeString:         base.Chassis,
		ExtraPropertiesRaw: nil,
	}

	err := ConfigureStorage("etcd", "mem:", []string{})
	if err != nil {
		suite.FailNowf("Unexpected error configuring storage", "err: %s", err)
	}

	err = SetXname("x000c0001", robj)
	if err != nil {
		suite.FailNowf("Unexpected error setting object", "err: %s", err)
	}

	res, err := GetXname("x0c01")
	if err != nil {
		suite.FailNowf("Unexpected error fetching data", "err: %s", err)
	}

	if res == nil {
		suite.FailNowf("Was not able to retrieve stored data", "")
	}

	log.Printf("Result: %v", res)

	if reflect.DeepEqual(robj, *res) {
		suite.FailNowf("Retrieved xname not as expected", "Expected \n%v, got \n%v",
			robj, res)
	}
}

func (suite *DatastoreTestSuite) Test_DeleteXname() {
	robj := sls_common.GenericHardware{
		Parent:     "x0c01",
		Children:   []string{},
		Xname:      "x0c1w02",
		Type:       sls_common.MgmtSwitch,
		Class:      sls_common.ClassRiver,
		TypeString: base.MgmtSwitch,
		ExtraPropertiesRaw: map[string]interface{}{
			"IPV4addr": "",
			"IPV6addr": "",
			"Username": "",
			"Password": "",
		},
	}

	err := ConfigureStorage("etcd", "mem:", []string{})
	if err != nil {
		suite.FailNowf("Unexpected error configuring storage", "err: %s", err)
	}

	err = SetXname("x000c1w002", robj)
	if err != nil {
		suite.FailNowf("Unexpected error setting object", "err: %s", err)
	}

	_, err = GetXname("x0c1w2")
	if err != nil {
		suite.FailNowf("Unexpected error checking data entry went OK", "err: %s", err)
	}

	err = DeleteXname("x0c01w002")
	if err != nil {
		suite.FailNowf("Unexpected error deleting data entry", "err: %s", err)
	}

	res, err := GetXname("x0c1w2")
	if err != nil {
		suite.FailNowf("Unexpected error while validating deletion", "err: %s", err)
	}
	if res != nil {
		suite.FailNowf("Object is unexpectedly still set?!", "value: %v", res)
	}

}

func (suite *DatastoreTestSuite) Test_GetNetwork() {
	nw := sls_common.Network{
		Name:     "HSN",
		FullName: "High Speed Network",
		Type:     sls_common.NetworkTypeSS10,
		IPRanges: []string{},
	}

	err := ConfigureStorage("etcd", "mem:", []string{})
	suite.NoError(err, "Unexpected error configuring storage")

	err = SetNetwork(nw)
	suite.NoError(err, "Failed to store Network")

	// ok, now get it back...
	res, err := GetNetwork("HSN")
	suite.NoError(err, "Error retrieving object")
	suite.NotEmpty(res)

	// Zero out the Last modifed fields, as those always change
	res.LastUpdated = 0
	res.LastUpdatedTime = ""
	suite.Equal(nw, res, "Returned object is not equal to input")
}

func (suite *DatastoreTestSuite) Test_GetNetwork_NotFound() {
	err := ConfigureStorage("etcd", "mem:", []string{})
	suite.NoError(err, "Unexpected error configuring storage")

	_, err = GetNetwork("foo")
	suite.Equal(database.NoSuch, err, "Unexpected error")
}

func (suite *DatastoreTestSuite) Test_SetNetwork() {
	nw := sls_common.Network{
		Name:     "HSN",
		FullName: "High Speed Network",
		Type:     sls_common.NetworkTypeSS10,
		IPRanges: []string{},
	}

	err := ConfigureStorage("etcd", "mem:", []string{})
	suite.NoError(err, "Unexpected error configuring storage")

	err = SetNetwork(nw)
	suite.NoError(err, "Failed to store Network")

	res, err := GetNetwork(nw.Name)
	suite.NoError(err, "Error retrieving object")
	suite.NotEmpty(res)

	// Zero out the Last modifed fields, as those always change
	res.LastUpdated = 0
	res.LastUpdatedTime = ""
	suite.Equal(nw, res, "Returned object is not equal to input")
}

func (suite *DatastoreTestSuite) Test_DeleteNetwork() {
	nw := sls_common.Network{
		Name:     "HSN",
		FullName: "High Speed Network",
		Type:     sls_common.NetworkTypeSS10,
		IPRanges: []string{},
	}

	err := ConfigureStorage("etcd", "mem:", []string{})
	suite.NoError(err, "Unexpected error configuring storage")

	err = SetNetwork(nw)
	suite.NoError(err, "Failed to store Network object")

	res, err := GetNetwork(nw.Name)
	suite.NoError(err, "Unable to fetch network (to verify present)")
	suite.NotEmpty(res)

	err = DeleteNetwork(nw.Name)
	suite.NoError(err, "Unable to delete network")

	_, err = GetNetwork(nw.Name)
	suite.Equal(database.NoSuch, err, "Unexpected error")
}
