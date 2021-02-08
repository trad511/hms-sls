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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	base "stash.us.cray.com/HMS/hms-base"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"

	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
)

/*
 * NOTE: You have to be really careful adding or modifying this test structure below. This config generator has to make
 * a lot of assumptions, so there are a lot of implicit ordering constraints that need to be honored. Also, if you
 * look at this data and then look at the tests you'll probably think well that just doesn't make any sense for a
 * couple of them. Case in point, anything at U20...that will have a slot number of 19. That's just the way the naming
 * convention works and actually the reason for the test.
 */

var HMNConnections = []shcd_parser.HMNRow{
	{
		Source:              "mn01",
		SourceRack:          "x3000",
		SourceLocation:      "u01",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p25",
	},
	{
		Source:         "wn01",
		SourceRack:     "x3000",
		SourceLocation: "u07",
	},
	{
		Source:              "wn02",
		SourceRack:          "x3000",
		SourceLocation:      "u09",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p28",
	},
	{
		Source:              "sn01",
		SourceRack:          "x3000",
		SourceLocation:      "u13",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p30",
	},
	{
		Source:              "nid000001",
		SourceRack:          "x3000",
		SourceLocation:      "u19",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p33",
	},
	{
		Source:              "nid000002",
		SourceRack:          "x3000",
		SourceLocation:      "U19",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p34",
	},
	{
		Source:              "cn-03",
		SourceRack:          "x3000",
		SourceLocation:      "u20",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p35",
	},
	{
		Source:              "cn04",
		SourceRack:          "x3000",
		SourceLocation:      "u20",
		SourceSubLocation:   "L",
		SourceParent:        "SubRack-001-cmc",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p36",
	},
	{
		Source:              "nid000005",
		SourceRack:          "x3000",
		SourceLocation:      "u21",
		SourceSubLocation:   "R",
		SourceParent:        "SubRack-002-cmc",
		DestinationRack:     "x3001",
		DestinationLocation: "u21",
		DestinationPort:     "p21",
	},
	{
		Source:              "SubRack-001-cmc",
		SourceRack:          "x3000",
		SourceLocation:      "u19",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p38",
	},
	{
		Source:              "SubRack-002-cmc",
		SourceRack:          "x3000",
		SourceLocation:      "u21",
		DestinationRack:     "x3001",
		DestinationLocation: "u21",
		DestinationPort:     "p22",
	},
	{
		Source:              "UAN",
		SourceRack:          "x3000",
		SourceLocation:      "u26",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p37",
	},
	{
		Source:              "sw-hsn001",
		SourceRack:          "x3000",
		SourceLocation:      "u22",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p47",
	},
	{
		Source:              "Columbia",
		SourceRack:          "x3000",
		SourceLocation:      "u24",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "p48",
	},
	{
		Source:              "x3000p0",
		SourceRack:          "x3000",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u38",
		DestinationPort:     "j41",
	},
	{
		Source:              "x3000door-Motiv",
		SourceRack:          "x3000",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u36",
		DestinationPort:     "j27",
	},
	{
		Source:              "CAN",
		SourceRack:          "cfcan",
		SourceLocation:      " ",
		DestinationRack:     "x3000",
		DestinationLocation: "u38",
		DestinationPort:     "j49",
	},
	{
		Source:              "x3000p0",
		SourceRack:          "x3000",
		SourceLocation:      "p0",
		DestinationRack:     "x3000",
		DestinationLocation: "u22",
		DestinationPort:     "j48",
	},
}

var TestSLSInputState = sls_common.SLSGeneratorInputState{
	ManagementSwitches: map[string]sls_common.GenericHardware{
		"x3000c0w22": buildMgmtSwitch("x3000", "x3000c0w22", "sw-leaf01", "10.254.0.2"),
		"x3000c0w38": buildMgmtSwitch("x3000", "x3000c0w38", "sw-leaf02", "10.254.0.3"),
		"x3001c0w21": buildMgmtSwitch("x3000", "x3001c0w21", "sw-leaf03", "10.254.0.4"),
	},

	RiverCabinets: map[string]sls_common.GenericHardware{
		"x3000": sls_common.GenericHardware{
			Parent:             "s0",
			Xname:              "x3000",
			Type:               sls_common.Cabinet,
			Class:              sls_common.ClassRiver,
			TypeString:         base.Cabinet,
			ExtraPropertiesRaw: sls_common.ComptypeCabinet{}, // Not required for current unit tests
		},
	},
	HillCabinets: map[string]sls_common.GenericHardware{
		"x5000": sls_common.GenericHardware{
			Parent:             "s0",
			Xname:              "x5000",
			Type:               sls_common.Cabinet,
			Class:              sls_common.ClassHill,
			TypeString:         base.Cabinet,
			ExtraPropertiesRaw: sls_common.ComptypeCabinet{}, // Not required for current unit tests
		},
	},
	MountainCabinets: map[string]sls_common.GenericHardware{
		"x9000": sls_common.GenericHardware{
			Parent:             "s0",
			Xname:              "x9000",
			Type:               sls_common.Cabinet,
			Class:              sls_common.ClassMountain,
			TypeString:         base.Cabinet,
			ExtraPropertiesRaw: sls_common.ComptypeCabinet{}, // Not required for current unit tests
		},
	},
	MountainStartingNid: 1000,
}

func buildMgmtSwitch(parent, xname, name, ipAddress string) sls_common.GenericHardware {
	return sls_common.GenericHardware{
		Parent:     parent,
		Xname:      xname,
		Type:       sls_common.MgmtSwitch,
		Class:      sls_common.ClassRiver,
		TypeString: base.MgmtSwitch,
		ExtraPropertiesRaw: sls_common.ComptypeMgmtSwitch{
			IP4Addr:          ipAddress,
			Model:            "S3048T-ON",
			SNMPAuthPassword: "vault://hms-creds/" + xname,
			SNMPAuthProtocol: "MD5",
			SNMPPrivPassword: "vault://hms-creds/" + xname,
			SNMPPrivProtocol: "DES",
			SNMPUsername:     "testuser",
			Aliases:          []string{name},
		},
	}
}

func stringArrayContains(array []string, needle string) bool {
	for _, item := range array {
		if item == needle {
			return true
		}
	}

	return false
}

type ConfigGeneratorTestSuite struct {
	suite.Suite

	generator   SLSStateGenerator
	allHardware map[string]sls_common.GenericHardware
}

func (suite *ConfigGeneratorTestSuite) SetupSuite() {
	// Setup logger for testing
	encoderCfg := zap.NewProductionEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevelAt(zap.DebugLevel),
	))

	g := NewSLSStateGenerator(logger, TestSLSInputState, HMNConnections)

	suite.allHardware = g.buildHardwareSection()
	suite.generator = g
}

func (suite *ConfigGeneratorTestSuite) TestVerifyNoEmptyHardware() {
	for xname, hardware := range suite.allHardware {
		suite.NotEmpty(xname)
		suite.NotEmpty(hardware.Xname)
		suite.NotEmpty(hardware.Parent)
		suite.NotEmpty(hardware.Type)
		suite.NotEmpty(hardware.TypeString)
		suite.NotEmpty(hardware.Class)

		// Note: The extra properties field maybe empty for some component types
	}
}

func (suite *ConfigGeneratorTestSuite) TestMasterNode() {
	/*
	  "x3000c0s1b0n0": {
	    "Parent": "x3000c0s1b0",
	    "Xname": "x3000c0s1b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100001,
	      "Role": "Management",
	      "SubRole": "Master",
	      "Aliases": [
	        "ncn-m001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s1b0n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s1b0")
	suite.Equal(hardware.Xname, "x3000c0s1b0n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Role, "Management")
	suite.Equal(hardwareExtraProperties.SubRole, "Master")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "ncn-m001"))
}

func (suite *ConfigGeneratorTestSuite) TestCANWorkerNode() {
	/*
	  "x3000c0s7b0n0": {
	    "Parent": "x3000c0s7b0",
	    "Xname": "x3000c0s7b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100002,
	      "Role": "Management",
	      "SubRole": "Worker",
	      "Aliases": [
	        "ncn-w001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s7b0n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s7b0")
	suite.Equal(hardware.Xname, "x3000c0s7b0n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Role, "Management")
	suite.Equal(hardwareExtraProperties.SubRole, "Worker")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "ncn-w001"))
}

func (suite *ConfigGeneratorTestSuite) TestWorkerNode() {
	/*
	  "x3000c0s9b0n0": {
	    "Parent": "x3000c0s9b0",
	    "Xname": "x3000c0s9b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100003,
	      "Role": "Management",
	      "SubRole": "Worker",
	      "Aliases": [
	        "ncn-w002"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s9b0n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s9b0")
	suite.Equal(hardware.Xname, "x3000c0s9b0n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Role, "Management")
	suite.Equal(hardwareExtraProperties.SubRole, "Worker")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "ncn-w002"))
}

func (suite *ConfigGeneratorTestSuite) TestStorageNode() {
	/*
	  "x3000c0s13b0n0": {
	    "Parent": "x3000c0s13b0",
	    "Xname": "x3000c0s13b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 100004,
	      "Role": "Management",
	      "SubRole": "Storage",
	      "Aliases": [
	        "ncn-s001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s13b0n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s13b0")
	suite.Equal(hardware.Xname, "x3000c0s13b0n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Role, "Management")
	suite.Equal(hardwareExtraProperties.SubRole, "Storage")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "ncn-s001"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_NID() {
	/*
	  "x3000c0s19b1n0": {
	    "Parent": "x3000c0s19b1",
	    "Xname": "x3000c0s19b1n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 1,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000001"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b1n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s19b1")
	suite.Equal(hardware.Xname, "x3000c0s19b1n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NID, 1)
	suite.Equal(hardwareExtraProperties.Role, "Compute")
	suite.Equal(hardwareExtraProperties.SubRole, "")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "nid000001"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_NID_CapitolSourceU() {
	/*
	  "x3000c0s19b2n0": {
	    "Parent": "x3000c0s19b2",
	    "Xname": "x3000c0s19b2n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 2,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000002"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b2n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s19b2")
	suite.Equal(hardware.Xname, "x3000c0s19b2n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NID, 2)
	suite.Equal(hardwareExtraProperties.Role, "Compute")
	suite.Equal(hardwareExtraProperties.SubRole, "")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "nid000002"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CN_WithHyphen() {
	/*
	  "x3000c0s19b3n0": {
	    "Parent": "x3000c0s19b3",
	    "Xname": "x3000c0s19b3n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 3,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000003"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b3n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s19b3")
	suite.Equal(hardware.Xname, "x3000c0s19b3n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NID, 3)
	suite.Equal(hardwareExtraProperties.Role, "Compute")
	suite.Equal(hardwareExtraProperties.SubRole, "")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "nid000003"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CN_WithoutHyphen() {
	/*
	  "x3000c0s19b4n0": {
	    "Parent": "x3000c0s19b4",
	    "Xname": "x3000c0s19b4n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "NID": 4,
	      "Role": "Compute",
	      "Aliases": [
	        "nid000004"
	      ]
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b4n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s19b4")
	suite.Equal(hardware.Xname, "x3000c0s19b4n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NID, 4)
	suite.Equal(hardwareExtraProperties.Role, "Compute")
	suite.Equal(hardwareExtraProperties.SubRole, "")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "nid000004"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_SwitchDifferentCabinet() {
	hardware, ok := suite.allHardware["x3000c0s21b1n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s21b1")
	suite.Equal(hardware.Xname, "x3000c0s21b1n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NID, 5)
	suite.Equal(hardwareExtraProperties.Role, "Compute")
	suite.Equal(hardwareExtraProperties.SubRole, "")
	suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "nid000005"))
}

func (suite *ConfigGeneratorTestSuite) TestCompute_CMC() {
	/*
	  "x3000c0s19b999": {
	    "Parent": "x3000",
	    "Xname": "x3000c0s19b999",
	    "Type": "comptype_chassis_bmc",
	    "Class": "River",
	    "TypeString": "ChassisBMC"
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s19b999"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000")
	suite.Equal(hardware.Xname, "x3000c0s19b999")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_chassis_bmc"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("ChassisBMC"))
}

func (suite *ConfigGeneratorTestSuite) TestUAN() {
	/*
	  "x3000c0s26b0n0": {
	    "Parent": "x3000c0s26b0",
	    "Xname": "x3000c0s26b0n0",
	    "Type": "comptype_node",
	    "Class": "River",
	    "TypeString": "Node",
	    "ExtraProperties": {
	      "Role": "Application",
	      "SubRole": "UAN"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0s26b0n0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0s26b0")
	suite.Equal(hardware.Xname, "x3000c0s26b0n0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Node"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
	suite.True(ok, "ExtraProperties type is not expected type.")

	// TODO: CASMHMS-3598
	//suite.Equal(hardwareExtraProperties.NID, 4) // No NIDs on UANs yet.
	suite.Equal(hardwareExtraProperties.Role, "Application")
	suite.Equal(hardwareExtraProperties.SubRole, "UAN")
	// suite.True(stringArrayContains(hardwareExtraProperties.Aliases, "uan-01")) // No alias on UANs yet.
}

func (suite *ConfigGeneratorTestSuite) TestManagementSwitch() {
	/*
	  "x3000c0w22": {
	    "Parent": "x3000",
	    "Xname": "x3000c0w22",
	    "Type": "comptype_mgmt_switch",
	    "Class": "River",
	    "TypeString": "MgmtSwitch",
	    "ExtraProperties": {
	      "IP4addr": "10.254.0.2",
	      "Model": "S3048T-ON",
	      "SNMPAuthPassword": "vault://hms-creds/x3000c0w22",
	      "SNMPAuthProtocol": "MD5",
	      "SNMPPrivPassword": "vault://hms-creds/x3000c0w22",
	      "SNMPPrivProtocol": "DES",
	      "SNMPUsername": "testuser"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0w22"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000")
	suite.Equal(hardware.Xname, "x3000c0w22")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_mgmt_switch"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("MgmtSwitch"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeMgmtSwitch)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.IP4Addr, "10.254.0.2")
	suite.Equal(hardwareExtraProperties.Model, "S3048T-ON")
	suite.Equal(hardwareExtraProperties.SNMPAuthPassword, "vault://hms-creds/x3000c0w22")
	suite.Equal(hardwareExtraProperties.SNMPAuthProtocol, "MD5")
	suite.Equal(hardwareExtraProperties.SNMPPrivPassword, "vault://hms-creds/x3000c0w22")
	suite.Equal(hardwareExtraProperties.SNMPPrivProtocol, "DES")
	suite.Equal(hardwareExtraProperties.SNMPUsername, "testuser")
}

func (suite *ConfigGeneratorTestSuite) TestHSNSwitch_HSN() {
	/*
	  "x3000c0r22b0": {
	    "Parent": "x3000",
	    "Xname": "x3000c0r22b0",
	    "Type": "comptype_rtr_bmc",
	    "Class": "River",
	    "TypeString": "RouterBMC",
	    "ExtraProperties": {
	      "Username": "vault://hms-creds/x3000c0r22b0",
	      "Password": "vault://hms-creds/x3000c0r22b0"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0r22b0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000")
	suite.Equal(hardware.Xname, "x3000c0r22b0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_rtr_bmc"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("RouterBMC"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeRtrBmc)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Username, "vault://hms-creds/x3000c0r22b0")
	suite.Equal(hardwareExtraProperties.Password, "vault://hms-creds/x3000c0r22b0")
}

func (suite *ConfigGeneratorTestSuite) TestHSNSwitch_Columbia() {
	/*
	  "x3000c0r24b0": {
	    "Parent": "x3000",
	    "Xname": "x3000c0r24b0",
	    "Type": "comptype_rtr_bmc",
	    "Class": "River",
	    "TypeString": "RouterBMC",
	    "ExtraProperties": {
	      "Username": "vault://hms-creds/x3000c0r24b0",
	      "Password": "vault://hms-creds/x3000c0r24b0"
	    }
	  },
	*/
	hardware, ok := suite.allHardware["x3000c0r24b0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000")
	suite.Equal(hardware.Xname, "x3000c0r24b0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_rtr_bmc"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("RouterBMC"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeRtrBmc)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.Username, "vault://hms-creds/x3000c0r24b0")
	suite.Equal(hardwareExtraProperties.Password, "vault://hms-creds/x3000c0r24b0")
}

func (suite *ConfigGeneratorTestSuite) TestCabinetPDUController() {
	/*
	   "x3000m0": {
	     "Parent": "x3000",
	     "Xname": "x3000m0",
	     "Type": "comptype_cab_pdu_controller",
	     "Class": "River",
	     "TypeString": "CabinetPDUController",
	   }
	*/

	hardware, ok := suite.allHardware["x3000m0"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000")
	suite.Equal(hardware.Xname, "x3000m0")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_cab_pdu_controller"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("CabinetPDUController"))

	suite.Nil(hardware.ExtraPropertiesRaw, "ExtraProperties type is not nil")
}

func (suite *ConfigGeneratorTestSuite) TestMgmtSwitchConnector_CabinetPDUController() {
	/*
		"x3000c0w22j48": {
		"Parent": "x3000c0w22",
		"Xname": "x3000c0w22j48",
		"Type": "comptype_mgmt_switch_connector",
		"Class": "River",
		"TypeString": "MgmtSwitchConnector",
		"ExtraProperties": {
			"NodeNics": [
				"x3000m0"
			],
			"VendorName": "ethernet1/1/48"
		}
	*/
	hardware, ok := suite.allHardware["x3000c0w22j48"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3000c0w22")
	suite.Equal(hardware.Xname, "x3000c0w22j48")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_mgmt_switch_connector"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("MgmtSwitchConnector"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeMgmtSwitchConnector)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NodeNics, []string{"x3000m0"})
	suite.Equal(hardwareExtraProperties.VendorName, "ethernet1/1/48")
}

func (suite *ConfigGeneratorTestSuite) TestMgmtSwitchConnector_ComputeSwitchDifferentCabinet() {
	hardware, ok := suite.allHardware["x3001c0w21j21"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "x3001c0w21")
	suite.Equal(hardware.Xname, "x3001c0w21j21")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_mgmt_switch_connector"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("MgmtSwitchConnector"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeMgmtSwitchConnector)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(hardwareExtraProperties.NodeNics, []string{"x3000c0s21b1"})
	suite.Equal(hardwareExtraProperties.VendorName, "ethernet1/1/21")
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_River() {
	/*
		{
			"Parent": "s0",
			"Xname": "x3000",
			"Type": "comptype_cabinet",
			"Class": "River",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x3000"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "s0")
	suite.Equal(hardware.Xname, "x3000")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_cabinet"))
	suite.Equal(hardware.Class, sls_common.CabinetType("River"))
	suite.Equal(hardware.TypeString, base.HMSType("Cabinet"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeCabinet)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(sls_common.ComptypeCabinet{}, hardwareExtraProperties)
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_Hill() {
	/*
		{
			"Parent": "s0",
			"Xname": "x5000",
			"Type": "comptype_cabinet",
			"Class": "Hill",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x5000"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "s0")
	suite.Equal(hardware.Xname, "x5000")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_cabinet"))
	suite.Equal(hardware.Class, sls_common.CabinetType("Hill"))
	suite.Equal(hardware.TypeString, base.HMSType("Cabinet"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeCabinet)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(sls_common.ComptypeCabinet{}, hardwareExtraProperties)
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassisBMC_Hill() {
	chassisBMCs := []string{
		"x5000c1",
		"x5000c3",
	}

	// Verify Hill Chassis BMCs
	for _, chassisBMC := range chassisBMCs {
		/*
			{
				"Parent": "x5000",
				"Xname": "x5000c0",
				"Type": "comptype_chassis_bmc",
				"Class": "Hill",
				"TypeString": "ChassisBMC"
			}
		*/
		hardware, ok := suite.allHardware[chassisBMC]
		suite.True(ok, "Unable to find xname.")

		suite.Equal(hardware.Parent, "x5000")
		suite.Equal(hardware.Xname, chassisBMC)
		suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_chassis_bmc"))
		suite.Equal(hardware.Class, sls_common.CabinetType("Hill"))
		suite.Equal(hardware.TypeString, base.HMSType("ChassisBMC"))

		suite.Nil(hardware.ExtraPropertiesRaw, "ExtraProperties type is not nil")
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyComputeNodes_Hill() {
	chassisBMCs := []string{
		"x5000c1",
		"x5000c3",
	}

	nodeBMCs := []string{
		"s0b0", "s0b1", // Slot 0
		"s1b0", "s1b1", // Slot 1
		"s2b0", "s2b1", // Slot 2
		"s3b0", "s3b1", // Slot 3
		"s4b0", "s4b1", // Slot 4
		"s5b0", "s5b1", // Slot 5
		"s6b0", "s6b1", // Slot 6
		"s7b0", "s7b1", // Slot 7
	}

	nodes := []string{
		"n0",
		"n1",
	}

	expectedNid := TestSLSInputState.MountainStartingNid

	// Verify Hill Compute Nodes
	for _, chassisBMC := range chassisBMCs {
		for _, nodeBMCSuffix := range nodeBMCs {
			for _, node := range nodes {
				/*
				   "x5000c1s0b0n0": {
				     "Parent": "x5000c1s0b0",
				     "Xname": "x5000c1s0b0n0",
				     "Type": "comptype_node",
				     "Class": "Hill",
				     "TypeString": "Node",
				     "ExtraProperties": {
				       "NID": 1000,
				       "Role": "Compute",
				       "Aliases": [
				         "nid001000"
				       ]
				     }
				   }
				*/

				// Calculate xnames for BMC and node
				nodeBMC := chassisBMC + nodeBMCSuffix
				nodeXname := nodeBMC + node

				hardware, ok := suite.allHardware[nodeXname]
				suite.True(ok, "Unable to find xname.")

				suite.Equal(hardware.Parent, nodeBMC)
				suite.Equal(hardware.Xname, nodeXname)
				suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
				suite.Equal(hardware.Class, sls_common.CabinetType("Hill"))
				suite.Equal(hardware.TypeString, base.HMSType("Node"))

				hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
				suite.True(ok, "ExtraProperties type is not expected type.")

				suite.Equal(hardwareExtraProperties.Role, "Compute")
				suite.Equal(hardwareExtraProperties.NID, expectedNid)

				expectedAlias := fmt.Sprintf("nid%06d", expectedNid)
				suite.Equal(hardwareExtraProperties.Aliases, []string{expectedAlias})

				expectedNid++
			}
		}
	}
}

func (suite *ConfigGeneratorTestSuite) TestCabinet_Mountain() {
	/*
		{
			"Parent": "s0",
			"Xname": "x9000",
			"Type": "comptype_cabinet",
			"Class": "Mountain",
			"TypeString": "Cabinet",
			"ExtraProperties": {}
		}
	*/

	hardware, ok := suite.allHardware["x9000"]
	suite.True(ok, "Unable to find xname.")

	suite.Equal(hardware.Parent, "s0")
	suite.Equal(hardware.Xname, "x9000")
	suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_cabinet"))
	suite.Equal(hardware.Class, sls_common.CabinetType("Mountain"))
	suite.Equal(hardware.TypeString, base.HMSType("Cabinet"))

	hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeCabinet)
	suite.True(ok, "ExtraProperties type is not expected type.")

	suite.Equal(sls_common.ComptypeCabinet{}, hardwareExtraProperties)
}

func (suite *ConfigGeneratorTestSuite) TestVerifyChassisBMC_Mountain() {
	chassisBMCs := []string{
		"x9000c0",
		"x9000c1",
		"x9000c2",
		"x9000c3",
		"x9000c4",
		"x9000c5",
		"x9000c6",
		"x9000c7",
	}

	// Verify Mountain Chassis BMCs
	for _, chassisBMC := range chassisBMCs {
		/*
			{
				"Parent": "x9000",
				"Xname": "x9000c0",
				"Type": "comptype_chassis_bmc",
				"Class": "Mountain",
				"TypeString": "ChassisBMC"
			}
		*/
		hardware, ok := suite.allHardware[chassisBMC]
		suite.True(ok, "Unable to find xname.")

		suite.Equal(hardware.Parent, "x9000")
		suite.Equal(hardware.Xname, chassisBMC)
		suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_chassis_bmc"))
		suite.Equal(hardware.Class, sls_common.CabinetType("Mountain"))
		suite.Equal(hardware.TypeString, base.HMSType("ChassisBMC"))

		suite.Nil(hardware.ExtraPropertiesRaw, "ExtraProperties type is not nil")
	}
}

func (suite *ConfigGeneratorTestSuite) TestVerifyComputeNodes_Mountain() {
	chassisBMCs := []string{
		"x9000c0",
		"x9000c1",
		"x9000c2",
		"x9000c3",
		"x9000c4",
		"x9000c5",
		"x9000c6",
		"x9000c7",
	}

	nodeBMCs := []string{
		"s0b0", "s0b1", // Slot 0
		"s1b0", "s1b1", // Slot 1
		"s2b0", "s2b1", // Slot 2
		"s3b0", "s3b1", // Slot 3
		"s4b0", "s4b1", // Slot 4
		"s5b0", "s5b1", // Slot 5
		"s6b0", "s6b1", // Slot 6
		"s7b0", "s7b1", // Slot 7
	}

	nodes := []string{
		"n0",
		"n1",
	}

	hillCabinetOffset := 64 // Nids for Mountain Hardware are generated after Hill
	expectedNid := TestSLSInputState.MountainStartingNid + hillCabinetOffset

	// Verify Mountain Compute Nodes
	for _, chassisBMC := range chassisBMCs {
		for _, nodeBMCSuffix := range nodeBMCs {
			for _, node := range nodes {
				/*
				   "x9000c1s0b0n0": {
				     "Parent": "x5000c1s0b0",
				     "Xname": "x5000c1s0b0n0",
				     "Type": "comptype_node",
				     "Class": "Mountain",
				     "TypeString": "Node",
				     "ExtraProperties": {
				       "NID": 1000,
				       "Role": "Compute",
				       "Aliases": [
				         "nid001000"
				       ]
				     }
				   }
				*/

				// Calculate xnames for BMC and node
				nodeBMC := chassisBMC + nodeBMCSuffix
				nodeXname := nodeBMC + node

				hardware, ok := suite.allHardware[nodeXname]
				suite.True(ok, "Unable to find xname.")

				suite.Equal(hardware.Parent, nodeBMC)
				suite.Equal(hardware.Xname, nodeXname)
				suite.Equal(hardware.Type, sls_common.HMSStringType("comptype_node"))
				suite.Equal(hardware.Class, sls_common.CabinetType("Mountain"))
				suite.Equal(hardware.TypeString, base.HMSType("Node"))

				hardwareExtraProperties, ok := hardware.ExtraPropertiesRaw.(sls_common.ComptypeNode)
				suite.True(ok, "ExtraProperties type is not expected type.")

				suite.Equal(hardwareExtraProperties.Role, "Compute")
				suite.Equal(hardwareExtraProperties.NID, expectedNid)

				expectedAlias := fmt.Sprintf("nid%06d", expectedNid)
				suite.Equal(hardwareExtraProperties.Aliases, []string{expectedAlias})

				expectedNid++
			}
		}
	}
}

func (suite *ConfigGeneratorTestSuite) Test_getSortedCabinetXNames() {
	cabinetXnames := []string{
		"x3000",
		"x9000",
		"x5001",
		"x0",
		"x100",
		"x110",
		"x111",
		"x10",
	}

	// Build up the list of cabinets from the list of xnames. We only care about the xname of the cabinet
	// when sorting.
	cabinets := map[string]sls_common.GenericHardware{}
	for _, xname := range cabinetXnames {
		cab := sls_common.GenericHardware{
			Xname: xname,
		}

		cabinets[xname] = cab
	}

	sortedCabinets := suite.generator.getSortedCabinetXNames(cabinets)

	expected := []string{
		"x0",
		"x10",
		"x100",
		"x110",
		"x111",
		"x3000",
		"x5001",
		"x9000",
	}

	suite.Equal(expected, sortedCabinets)
}

func TestConfigGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigGeneratorTestSuite))
}
