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
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-sls/internal/database"
	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
)

type testData struct {
	op        string
	setURL    string
	getURL    string
	setString []byte
	getHWData sls_common.GenericHardware
}

var router *mux.Router
var routes Routes

const (
	hwURLBase = "http://localhost:8376/v1/hardware"
)

// Payloads for POST tests, happy path

var payloads = []testData{
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s0b0n1",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n1","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1236,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n1",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1236, Role: "Compute"},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s0b0",
		json.RawMessage(`{"Parent":"x0c0s0","Xname":"x0c0s0b0","Type":"comptype_ncard","TypeString":"NodeBMC","Class":"Mountain","ExtraProperties":{"IP6addr":"DHCPv6","IP4addr":"10.1.1.1","Username":"root","Password":"vault://root_pw"}}`),
		sls_common.GenericHardware{"x0c0s0",
			nil,
			"x0c0s0b0",
			sls_common.NodeBMC,
			"Mountain",
			"NodeBMC",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNodeBmc{"DHCPv6",
				"10.1.1.1",
				"root",
				"vault://root_pw",
				[]string{}},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s0",
		json.RawMessage(`{"Parent":"x0c0","Xname":"x0c0s0","Type":"comptype_compmod","TypeString":"ComputeModule","Class":"Mountain","ExtraProperties":{"PoweredBy":["x0m0v0","x0m0v1"]}}`),
		sls_common.GenericHardware{"x0c0",
			nil,
			"x0c0s0",
			sls_common.ComputeModule,
			"Mountain",
			"ComputeModule",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeCompmodPowerConnector{[]string{"x0m0v0", "x0m0v1"}},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s1b0n0",
		json.RawMessage(`{"Parent":"x0c0s1b0","Xname":"x0c0s1b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1240,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s1b0",
			nil,
			"x0c0s1b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1240, Role: "Compute"},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s1b0n1",
		json.RawMessage(`{"Parent":"x0c0s1b0","Xname":"x0c0s1b0n1","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1237,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s1b0",
			nil,
			"x0c0s1b0n1",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1237, Role: "Compute"},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s1b0",
		json.RawMessage(`{"Parent":"x0c0s1","Xname":"x0c0s1b0","Type":"comptype_ncard","TypeString":"NodeBMC","Class":"Mountain","ExtraProperties":{"IP6addr":"DHCPv6","IP4addr":"10.1.1.2","Username":"root","Password":"vault://root_pw"}}`),
		sls_common.GenericHardware{"x0c0s1",
			nil,
			"x0c0s1b0",
			sls_common.NodeBMC,
			"Mountain",
			"NodeBMC",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNodeBmc{"DHCPv6",
				"10.1.1.2",
				"root",
				"vault://root_pw",
				[]string{}},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x0c0s1",
		json.RawMessage(`{"Parent":"x0c0","Xname":"x0c0s1","Type":"comptype_compmod","TypeString":"ComputeModule","Class":"Mountain","ExtraProperties":{"PoweredBy":["x0m1v1","x0m1v2"]}}`),
		sls_common.GenericHardware{"x0c0",
			nil,
			"x0c0s1",
			sls_common.ComputeModule,
			"Mountain",
			"ComputeModule",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeCompmodPowerConnector{[]string{"x0m1v1", "x0m1v2"}},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x3m0p0",
		json.RawMessage(`{"Parent":"x3m0","Xname":"x3m0p0","Type":"comptype_cab_pdu","TypeString":"CabinetPDU","Class":"River"}`),
		sls_common.GenericHardware{"x3m0",
			nil,
			"x3m0p0",
			sls_common.CabinetPDU,
			"River",
			"CabinetPDU",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			nil,
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x2c0s0",
		json.RawMessage(`{"Parent":"x2c0","Xname":"x2c0s0","Type":"comptype_compmod","TypeString":"ComputeModule","Class":"Mountain","ExtraProperties":{"PoweredBy":["x2m0v0","x2m0v1"]}}`),
		sls_common.GenericHardware{"x2c0",
			nil,
			"x2c0s0",
			sls_common.ComputeModule,
			"Mountain",
			"ComputeModule",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeCompmodPowerConnector{[]string{"x2m0v0", "x2m0v1"}},
			nil,
		},
	},

	//Hierarchy stuff
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x3c3s0",
		json.RawMessage(`{"Parent":"x3c3","Xname":"x3c3s0","Type":"comptype_compmod","TypeString":"ComputeModule","Class":"River","ExtraProperties":{"PoweredBy":["x3m0v0","x3m0v1"]}}`),
		sls_common.GenericHardware{"x3c3",
			nil,
			"x3c3s0",
			sls_common.ComputeModule,
			"River",
			"ComputeModule",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeCompmodPowerConnector{[]string{"x3m0v0", "x3m0v1"}},
			nil,
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x3c3s0b0",
		json.RawMessage(`{"Parent":"x3c3s0","Xname":"x3c3s0b0","Type":"comptype_ncard","TypeString":"NodeBMC","Class":"River","ExtraProperties":{"IP6addr":"DHCPv6","IP4addr":"10.3.1.1","Username":"root","Password":"vault://root_pw"}}`),
		sls_common.GenericHardware{Parent: "x3c3s0",
			Xname:      "x3c3s0b0",
			Type:       sls_common.NodeBMC,
			Class:      "River",
			TypeString: "NodeBMC",
			ExtraPropertiesRaw: sls_common.ComptypeNodeBmc{"DHCPv6",
				"10.3.1.1",
				"root",
				"vault://root_pw",
				[]string{}},
		},
	},
	testData{"POST",
		hwURLBase,
		hwURLBase + "/x3c3s0b0n3",
		json.RawMessage(`{"Parent":"x3c3s0b0","Xname":"x3c3s0b0n3","Type":"comptype_node","TypeString":"Node","Class":"River","ExtraProperties":{"NID":3303,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x3c3s0b0",
			nil,
			"x3c3s0b0n3",
			sls_common.Node,
			"River",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 3303, Role: "Compute"},
			nil,
		},
	},
}

// Payload for component replacement via PUT, happy path

var putrpl = testData{"PUT",
	hwURLBase + "/x0c0s1b0n0",
	hwURLBase + "/x0c0s1b0n0",
	json.RawMessage(`{"Parent":"x0c0s1b0","Xname":"x0c0s1b0n0","Type":"comptype_node","TypeString":"Node","Class":"River","ExtraProperties":{"NID":5555,"Role":"Management"}}`),
	sls_common.GenericHardware{"x0c0s1b0",
		nil,
		"x0c0s1b0n0",
		sls_common.Node,
		"River",
		"Node",
		0,
		"2014-07-16 20:55:46 +0000 UTC",
		sls_common.ComptypeNode{NID: 5555, Role: "Management"},
		nil,
	},
}

// Payload for component creation via PUT, happy path

var putnewcomp = testData{"PUT",
	hwURLBase + "/x0c0s1b0n7",
	hwURLBase + "/x0c0s1b0n7",
	json.RawMessage(`{"Parent":"x0c0s1b0","Xname":"x0c0s1b0n7","Type":"comptype_node","TypeString":"Node","Class":"River","ExtraProperties":{"NID":7777,"Role":"Management"}}`),
	sls_common.GenericHardware{"x0c0s1b0",
		nil,
		"x0c0s1b0n7",
		sls_common.Node,
		"River",
		"Node",
		0,
		"2014-07-16 20:55:46 +0000 UTC",
		sls_common.ComptypeNode{NID: 7777, Role: "Management"},
		nil,
	},
}

// Payloads for POST error tests

var posterrs = []testData{
	testData{"POST", //Missing Xname
		hwURLBase,
		"",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //Invalid XName
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"z0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //No parent
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //Invalid parent
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"z0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //No class
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //Invalid class
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Zountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //No type
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //Invalid type
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"zomptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", // No typestring
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //Invalid typestring
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Zode","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"POST", //bad JSON
		hwURLBase,
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent","x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
}

// Payloads for GET error tests

var geterrs = []testData{
	testData{"GET",
		hwURLBase + "/z0c0s0b0n0",
		hwURLBase + "/z0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
}

// Payloads for PUT error tests

var puterrs = []testData{
	testData{"PUT", //Invalid xname in URL
		hwURLBase + "/z0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //No xname in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //No parent in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //No type in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //No typestring in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //No class in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //xname in URL != xname in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n1","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //Invalid parent in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"z0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //Invalid type in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"zomptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //Invalid typestring in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Zode","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //Invalid class in payload
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Zountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //Bad JSON
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent","x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"PUT", //type != typestring
		hwURLBase + "/x0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"NodeBMC","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
}

// Payloads for DELETE error tests

var delerrs = []testData{
	testData{"DELETE", //Invalid xname in URL
		hwURLBase + "/z0c0s0b0n0",
		hwURLBase + "/x0c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
	testData{"DELETE", //non-existent component xname in URL
		hwURLBase + "/x9c0s0b0n0",
		hwURLBase + "/x9c0s0b0n0",
		json.RawMessage(`{"Parent":"x0c0s0b0","Xname":"x0c0s0b0n0","Type":"comptype_node","TypeString":"Node","Class":"Mountain","ExtraProperties":{"NID":1234,"Role":"Compute"}}`),
		sls_common.GenericHardware{"x0c0s0b0",
			nil,
			"x0c0s0b0n0",
			sls_common.Node,
			"Mountain",
			"Node",
			0,
			"2014-07-16 20:55:46 +0000 UTC",
			sls_common.ComptypeNode{NID: 1234, Role: "Compute"},
			nil,
		},
	},
}

var dbInitOK bool = false

func hwDBClear() {
	var jdata sls_common.GenericHardwareArray

	//This is kind of cheating, GET /hardware used to test GET /hardware...
	req, rerr := http.NewRequest("GET", hwURLBase, nil)
	if rerr != nil {
		fmt.Println("ERROR creating http GET request:", rerr)
		return
	}
	gw := httptest.NewRecorder()
	router.ServeHTTP(gw, req)
	if gw.Code != http.StatusOK {
		return
	}
	//Got a list back.   Delete it all.

	jerr := json.Unmarshal(gw.Body.Bytes(), &jdata)
	if jerr != nil {
		fmt.Println("ERROR unmarshaling HW array:", jerr)
		return
	}

	for _, hw := range jdata {
		req, rerr := http.NewRequest("DELETE", hwURLBase+"/"+hw.Xname, nil)
		if rerr != nil {
			fmt.Println("ERROR creating http DELETE request:", rerr)
			return
		}
		gw := httptest.NewRecorder()
		router.ServeHTTP(gw, req)
		if gw.Code != http.StatusOK {
			fmt.Printf("ERROR deleting '%s', status %d/%s\n", hw.Xname,
				gw.Code, http.StatusText(gw.Code))
			continue
		}
	}
}

func hwCompare(a, b sls_common.GenericHardware) error {
	if a.Parent != b.Parent {
		return fmt.Errorf("Miscompare of Parent field, '%s'/'%s'",
			a.Parent, b.Parent)
	}
	if a.Xname != b.Xname {
		return fmt.Errorf("Miscompare of Xname field, '%s'/'%s'",
			a.Xname, b.Xname)
	}
	if a.Type != b.Type {
		return fmt.Errorf("Miscompare of Type field, '%s'/'%s'",
			string(a.Type), string(b.Type))
	}
	if a.TypeString != b.TypeString {
		return fmt.Errorf("Miscompare of TypeString field, '%s'/'%s'",
			string(a.TypeString), string(b.TypeString))
	}
	if a.Class != b.Class {
		return fmt.Errorf("Miscompare of Class field, '%s'/'%s'",
			string(a.Class), string(b.Class))
	}

	//Treat ExtraProperties like a map[string]interface{}.

	if (a.ExtraPropertiesRaw == nil) != (b.ExtraPropertiesRaw == nil) {
		return fmt.Errorf("Miscompare of ExtraPropertiesRaw, only one present.")
	}
	if a.ExtraPropertiesRaw != nil {
		switch a.Type {
		case sls_common.ComputeModule:
			ma := a.ExtraPropertiesRaw.(map[string]interface{})
			mb := b.ExtraPropertiesRaw.(map[string]interface{})
			if !reflect.DeepEqual(ma, mb) {
				return fmt.Errorf("Miscompare of ExtraPropertiesRaw: '%v'/'%v'",
					ma, mb)
			}
		case sls_common.NodeBMC:
			ma := a.ExtraPropertiesRaw.(map[string]interface{})
			mb := b.ExtraPropertiesRaw.(map[string]interface{})
			if !reflect.DeepEqual(ma, mb) {
				return fmt.Errorf("Miscompare of ExtraPropertiesRaw: '%v'/'%v'",
					ma, mb)
			}
		case sls_common.Node:
			ma := a.ExtraPropertiesRaw.(map[string]interface{})
			mb := b.ExtraPropertiesRaw.(map[string]interface{})
			if !reflect.DeepEqual(ma, mb) {
				return fmt.Errorf("Miscompare of ExtraPropertiesRaw: '%v'/'%v'",
					ma, mb)
			}
		case sls_common.CabinetPDU:
			ma := a.ExtraPropertiesRaw.(map[string]interface{})
			mb := b.ExtraPropertiesRaw.(map[string]interface{})
			if !reflect.DeepEqual(ma, mb) {
				return fmt.Errorf("Miscompare of ExtraPropertiesRaw: '%v'/'%v'",
					ma, mb)
			}
		default:
			return fmt.Errorf("INTERNAL ERROR: unhandled component type: %s",
				string(a.TypeString))
		}
	}

	return nil
}

func setEnvIfEmpty(env string, value string) {
	if _, ok := os.LookupEnv(env); !ok {
		_ = os.Setenv(env, value)
	}
}

func dbInit() {
	if dbInitOK {
		return
	}

	setEnvIfEmpty("POSTGRES_HOST", "postgres")
	setEnvIfEmpty("POSTGRES_USER", "slsuser")
	setEnvIfEmpty("POSTGRES_PASSWORD", "slsuser")
	setEnvIfEmpty("POSTGRES_DB", "sls")
	setEnvIfEmpty("DBOPTS", "sslmode=disable")
	setEnvIfEmpty("DBUSER", "slsuser")
	setEnvIfEmpty("DBPASS", "slsuser")

	// The NewDatabase method will try forever to connect.
	_ = database.NewDatabase()
	dbInitOK = true
}

//Handles POST/PUT/DELETE

func doSet(pl testData) error {
	fmt.Println(string(pl.setString))
	preq, preqerr := http.NewRequest(pl.op, pl.setURL,
		bytes.NewBuffer(pl.setString))
	if preqerr != nil {
		return fmt.Errorf("ERROR creating http POST request: %v", preqerr)
	}

	pw := httptest.NewRecorder()
	router.ServeHTTP(pw, preq)

	//Check response code

	if pw.Code != http.StatusOK {
		return fmt.Errorf("ERROR response in %s operation: %d/%s",
			pl.op, pw.Code, http.StatusText(pw.Code))
	}

	return nil
}

func doGet(pl testData) (sls_common.GenericHardware, error) {
	var jdata sls_common.GenericHardware

	greq, greqerr := http.NewRequest("GET", pl.getURL, nil)
	if greqerr != nil {
		return jdata, greqerr
	}

	gw := httptest.NewRecorder()
	router.ServeHTTP(gw, greq)

	//Check response code

	if gw.Code != http.StatusOK {
		return jdata, fmt.Errorf("Bad response in GET op: %d/%s",
			gw.Code, http.StatusText(gw.Code))
	}

	jerr := json.Unmarshal(gw.Body.Bytes(), &jdata)
	if jerr != nil {
		return jdata, jerr
	}

	return jdata, nil
}

//Gotta marshal and re-unmarshal to get the ExtraProperties stuff
//to work the same as what came into the HTTP func.  Initializing
//the expected data with a JSON string literal in the ExtraProperties
//field just doesn't seem to work.

func plMassage(hw sls_common.GenericHardware) (sls_common.GenericHardware, error) {
	var jexp sls_common.GenericHardware
	ja, jaerr := json.Marshal(hw)
	if jaerr != nil {
		return jexp, jaerr
	}
	jaerr = json.Unmarshal(ja, &jexp)
	if jaerr != nil {
		return jexp, jaerr
	}

	return jexp, nil
}

func findTest(xname string) *testData {
	for ii, pl := range payloads {
		if pl.getHWData.Xname == xname {
			return &payloads[ii]
		}
	}
	return nil
}

func Test_doHardwarePost(t *testing.T) {
	var jdata, jexp sls_common.GenericHardware
	var tpl *testData
	var targ, child string
	var err error

	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	// Clear the database.
	database.DeleteAllGenericHardware()

	for ii, pl := range payloads {
		t.Logf("POST test %d...\n", ii)

		//Set up and execute the POST/PUT
		psterr := doSet(pl)
		if psterr != nil {
			t.Errorf("ERROR in POST test %d: %v", ii, psterr)
		}

		//Set up and execute the GET

		jdata, gterr := doGet(pl)
		if gterr != nil {
			t.Errorf("ERROR in POST test %d GET op: %v", ii, gterr)
		}

		jexp, jerr := plMassage(pl.getHWData)
		if jerr != nil {
			t.Errorf("ERROR in POST test %d data massage: %v", ii, jerr)
		}

		cmperr := hwCompare(jexp, jdata)
		if cmperr != nil {
			t.Errorf("Data miscompare in POST test %d\n'%v'\n", ii, cmperr)
		}
	}

	//Now do a GET to insure the DB hooked up the children correctly

	t.Log("Checking x3 hierarchy.")

	// x3c3s0

	targ = "x3c3s0"
	child = "x3c3s0b0"
	tpl = findTest(targ)
	if tpl == nil {
		t.Errorf("ERROR: Can't find test data for '%s'", targ)
	}

	//Should have x3c3s0b0 as a child

	jdata, err = doGet(*tpl)
	if err != nil {
		t.Error(err)
	}

	if len(jdata.Children) != 1 {
		t.Errorf("ERROR: '%s' has incorrect number of children, has %d, should be 1.",
			targ, len(jdata.Children))
	}
	if jdata.Children[0] != child {
		t.Errorf("ERROR: '%s' has incorrect child, exp: '%s', got '%s'",
			targ, child, jdata.Children[0])
	}

	// x3c3s0b0

	targ = "x3c3s0b0"
	child = "x3c3s0b0n3"
	tpl = findTest(targ)
	if tpl == nil {
		t.Errorf("ERROR: Can't find test data for '%s'", targ)
	}

	//Should have x3c3s0b0 as a child

	jdata, err = doGet(*tpl)
	if err != nil {
		t.Error(err)
	}

	if len(jdata.Children) != 1 {
		t.Errorf("ERROR: '%s' has incorrect number of children, has %d, should be 1.",
			targ, len(jdata.Children))
	}
	if jdata.Children[0] != child {
		t.Errorf("ERROR: '%s' has incorrect child, exp: '%s', got '%s'",
			targ, child, jdata.Children[0])
	}

	t.Log("X3 hiearchy is intact.")

	//Do a DELETE of the top of the hierarchical stuff in x3

	t.Log("Deleting X3 hierarchy via topmost component")
	targ = "x3c3s0"
	tpl = findTest(targ)
	if tpl == nil {
		t.Errorf("ERROR: Can't find test data for '%s'", targ)
	}
	//Make copy and use DELETE for the operation, GET URL for DELETE
	tplcp := *tpl
	tplcp.op = "DELETE"
	tplcp.setURL = tplcp.getURL
	err = doSet(tplcp)
	if err != nil {
		t.Errorf("ERROR in DELETE test: %v", err)
	}

	//Do a GET to insure the right stuff got deleted

	var htargs = []string{"x3c3s0", "x3c3s0b0", "x3c3s0b0n3"}

	for _, st := range htargs {
		tpl = findTest(st)
		if tpl == nil {
			t.Errorf("ERROR: Can't find test data for '%s'", targ)
		}

		//GET should fail

		greq, greqerr := http.NewRequest("GET", tpl.getURL, nil)
		if greqerr != nil {
			t.Error("ERROR creating http GET request:", greqerr)
		}

		gw := httptest.NewRecorder()
		router.ServeHTTP(gw, greq)

		if gw.Code != http.StatusNotFound {
			t.Errorf("ERROR, component '%s' was found, shouldn't have been.",
				st)
		}
	}

	//Test PUT on existing component

	t.Log("Testing PUT on existing component.")

	err = doSet(putrpl)
	if err != nil {
		t.Error("ERROR creating http PUT request:", err)
	}

	jdata, err = doGet(putrpl)
	if err != nil {
		t.Error(err)
	}

	jexp, err = plMassage(putrpl.getHWData)
	if err != nil {
		t.Error(err)
	}

	cmperr := hwCompare(jexp, jdata)
	if cmperr != nil {
		t.Errorf("Data miscompare in PUT Replace test '%v'\n", cmperr)
	}

	//Test PUT on non-existing component

	t.Log("Testing PUT on non-existing component.")

	err = doSet(putnewcomp)
	if err != nil {
		t.Error("ERROR creating http PUT request:", err)
	}

	jdata, err = doGet(putnewcomp)
	if err != nil {
		t.Error(err)
	}

	jexp, err = plMassage(putnewcomp.getHWData)
	if err != nil {
		t.Error(err)
	}

	cmperr = hwCompare(jexp, jdata)
	if cmperr != nil {
		t.Errorf("Data miscompare in PUT Replace test '%v'\n", cmperr)
	}

	//Test POST on existing component

	t.Log("Testing POST on existing component.")

	preq, preqerr := http.NewRequest(payloads[0].op, payloads[0].setURL,
		bytes.NewBuffer(payloads[0].setString))
	if preqerr != nil {
		t.Error("ERROR creating http POST request:", preqerr)
	}

	pw := httptest.NewRecorder()
	router.ServeHTTP(pw, preq)
	if pw.Code == http.StatusOK {
		t.Errorf("ERROR POST to existing component worked, should not have.")
	}

	//Test GET from non-existent component

	preq, preqerr = http.NewRequest("GET", hwURLBase+"/"+"x77c7s7b0n7", nil)
	if preqerr != nil {
		t.Error("ERROR creating http GET request:", preqerr)
	}

	pw = httptest.NewRecorder()
	router.ServeHTTP(pw, preq)
	if pw.Code == http.StatusOK {
		t.Error("ERROR, GET from non-existent component worked, should not have.")
	}
}

func Test_HardwarePostErrs(t *testing.T) {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	for ii, pl := range posterrs {
		t.Logf("POST error test %d...\n", ii)
		psterr := doSet(pl)
		if psterr == nil {
			t.Errorf("ERROR in POST error test %d: didn't fail!", ii)
		}
	}
}

func Test_HardwareGetErrs(t *testing.T) {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	t.Logf("GET error test...")
	_, psterr := doGet(geterrs[0])
	if psterr == nil {
		t.Errorf("ERROR in GET error test, didn't fail!")
	}
}

func Test_HardwarePutErrs(t *testing.T) {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	for ii, pl := range puterrs {
		t.Logf("PUT error test %d...\n", ii)
		psterr := doSet(pl)
		if psterr == nil {
			t.Errorf("ERROR in PUT error test %d: didn't fail!", ii)
		}
	}
}

func Test_HardwareDelErrs(t *testing.T) {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	for ii, pl := range delerrs {
		t.Logf("DELETE error test %d...\n", ii)
		psterr := doSet(pl)
		if psterr == nil {
			t.Errorf("ERROR in DELETE error test %d: didn't fail!", ii)
		}
	}
}

func Test_doHardwareGetAll(t *testing.T) {
	var jdata sls_common.GenericHardwareArray

	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()
	hwDBClear()

	//First POST all the stuff in from the happy path POST list

	for ii, pl := range payloads {
		t.Logf("Loading payload %d (%s)...\n", ii, pl.getHWData.Xname)

		//Set up and execute the POST/PUT
		psterr := doSet(pl)
		if psterr != nil {
			t.Errorf("ERROR in POST of payload %d: %v", ii, psterr)
		}
	}

	//Now do a GET from /hardware

	t.Logf("Fetching all components via GET /hardware...")
	req, rerr := http.NewRequest("GET", hwURLBase, nil)
	if rerr != nil {
		t.Error("ERROR creatting http GET for /hardware:", rerr)
	}
	gw := httptest.NewRecorder()
	router.ServeHTTP(gw, req)
	if gw.Code != http.StatusOK {
		t.Errorf("ERROR bad response from /hardware GET: %d/%s\n",
			gw.Code, http.StatusText(gw.Code))
	}

	//Iterate the list we got back and verify that everything is there and
	//correct

	jerr := json.Unmarshal(gw.Body.Bytes(), &jdata)
	if jerr != nil {
		t.Error("ERROR unmarshaling /hardware GET data:", jerr)
	}

	for _, hw := range jdata {
		tdp := findTest(hw.Xname)
		if tdp == nil {
			t.Errorf("ERROR, can't find happy path test data for '%s'\n",
				hw.Xname)
		}

		cmpa, _ := plMassage(tdp.getHWData)
		cmpb, _ := plMassage(hw)
		cmperr := hwCompare(cmpa, cmpb)
		if cmperr != nil {
			t.Error("ERROR, miscompare of /hardware GET data:", cmperr)
		}
	}
}

type HardwareTestSuite struct {
	suite.Suite
}

func (suite *HardwareTestSuite) SetupSuite() {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}

	dbInit()
	hwDBClear()
}

func (suite *HardwareTestSuite) TestVerifyPOSTAllTypes() {
	// Verify the hardware search endpoint accepts the following SLS types via the type query param
	tests := []string{
		// TODO Due to CASMHMS-4270 the following can only only be posted into SLS whith an empty parent:
		// "d0", // "comptype_cdu",                       // dD
		// "x1", // "comptype_cabinet",                   // xX

		// TODO due to CASMHMS-4669 the follow can't be posted in due to invalid parent field
		// "x1c1r1t1f1", // "comptype_rtr_tor_fpga",            // xXcCrRtTfF
		// "x1c1h1s1",   // "comptype_hl_switch",              // xXcChHsS

		"d0w1",     // "comptype_cdu_mgmt_switch",           // dDwW
		"x1d1",     // "comptype_cab_cdu",                   // xXdD
		"x1m1",     // "comptype_cab_pdu_controller",        // xXmM
		"x1m1p0",   // "comptype_cab_pdu",                   // xXmMpP
		"x1m1i1",   // "comptype_cab_pdu_nic",               // xXmMiI
		"x1m1p0j1", // "comptype_cab_pdu_outlet",            // xXmMpPjJ DEPRECATED
		"x1m1p0v1", // "comptype_cab_pdu_pwr_connector",     // xXmMpPvV

		"x1c1",         // "comptype_chassis",                 // xXcC
		"x1c1b0",       // "comptype_chassis_bmc",             // xXcCbB
		"x1c1t0",       // "comptype_cmm_rectifier",           // xXcCtT
		"x1c1f0",       // "comptype_cmm_fpga",                // xXcCfF
		"x1e1",         // "comptype_cec",                     // xXeE
		"x1c1s1",       // "comptype_compmod",                 // xXcCsS
		"x1c1r1",       // "comptype_rtrmod",                  // xXcCrR
		"x1c1s1b1",     // "comptype_ncard",                   // xXcCsSbB
		"x1c1s1b1i1",   // "comptype_bmc_nic",                 // xXcCsSbBiI
		"x1c1s1e1",     // "comptype_node_enclosure",          // xXcCsSeE
		"x1c1s1v1",     // "comptype_compmod_power_connector", // xXcCsSvV
		"x1c1s1b1n1",   // "comptype_node",                    // xXcCsSbBnN
		"x1c1s1b1n1p1", // "comptype_node_processor",          // xXcCsSbBnNpP
		"x1c1s1b1n1i1", // "comptype_node_nic",                // xXcCsSbBnNiI
		"x1c1s1b1n1h1", // "comptype_node_hsn_nic",            // xXcCsSbBnNhH
		"x1c1s1b1n1d1", // "comptype_dimm",                    // xXcCsSbBnNdD
		"x1c1s1b1n1a1", // "comptype_node_accel",              // xXcCsSbBnNaA
		"x1c1s1b1f0",   // "comptype_node_fpga",               // xXcCsSbBfF
		"x1c1r1a1",     // "comptype_hsn_asic",                // xXcCrRaA
		"x1c1r1f1",     // "comptype_rtr_fpga",                // xXcCrRfF
		"x1c1r1b1",     // "comptype_rtr_bmc",                 // xXcCrRbB
		"x1c1r1b1i1",   // "comptype_rtr_bmc_nic",             // xXcCrRbBiI

		"x1c1r1e1",   // "comptype_hsn_board",             // xXcCrReE
		"x1c1r1a1l1", // "comptype_hsn_link",              // xXcCrRaAlL
		"x1c1r1j1",   // "comptype_hsn_connector",         // xXcCrRjJ
		"x1c1r1j1p1", // "comptype_hsn_connector_port",    // xXcCrRjJpP
		"x1c1w1",     // "comptype_mgmt_switch",           // xXcCwW
		"x1c1w1j1",   // "comptype_mgmt_switch_connector", // xXcCwWjJ
	}

	for _, xname := range tests {
		suite.True(base.IsHMSCompIDValid(xname), xname)
		hmsType := base.GetHMSType(xname)
		slsType := sls_common.HMSTypeToHMSStringType(hmsType)

		suite.NotEqual("INVALID", slsType)

		h := sls_common.GenericHardware{
			Parent:     base.GetHMSCompParent(xname),
			Xname:      xname,
			Class:      sls_common.ClassRiver,
			Type:       slsType,
			TypeString: hmsType,
		}

		payload, err := json.Marshal(h)
		suite.NoError(err)

		suite.T().Log(string(payload))

		req, preqerr := http.NewRequest("POST", hwURLBase, bytes.NewBuffer(payload))
		suite.NoError(preqerr, "creating http POST request")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		//Check response code
		suite.Equal(http.StatusOK, response.Code, "Response: %s", response.Body.String())
	}
}

func (suite *HardwareTestSuite) TestVerifyPUTAllTypes() {
	// TODO CASMHMS-4670 PUT will create an object if it does not already exist
	// Right now this test expliots that fact, but when CASMHMS-4670 is addressed this test needs to
	// be updated.

	// Verify the hardware search endpoint accepts the following SLS types via the type query param
	tests := []string{
		// TODO Due to CASMHMS-4270 the following components can not be used with PUT
		// "d0",       // "comptype_cdu",                // dD
		// "x1",       // "comptype_cabinet",            // xX

		// TODO Due to CASMHMS-4667 the following components do not have a valid parent type defined
		// "x1c1r1t1f1", //"x1c1r1T1f1",   // "comptype_rtr_tor_fpga",
		// "x1c1h1s1",   // "comptype_hl_switch",             // xXcChHsS

		"d0w1",     // "comptype_cdu_mgmt_switch",    // dDwW
		"x1d1",     // "comptype_cab_cdu",            // xXdD
		"x1m1p0",   // "comptype_cab_pdu",            // xXmMpP
		"x1m1i1",   // "comptype_cab_pdu_nic",        // xXmMiI
		"x1m1p0j1", // "comptype_cab_pdu_outlet",     // xXmMpPjJ DEPRECATED
		"x1m1p0v1", // "comptype_cab_pdu_pwr_connector",     // xXmMpPvV

		"x1c1",         // "comptype_chassis",                 // xXcC
		"x1c1b0",       // "comptype_chassis_bmc",             // xXcCbB
		"x1c1t0",       // "comptype_cmm_rectifier",           // xXcCtT
		"x1c1f0",       // "comptype_cmm_fpga",                // xXcCfF
		"x1e1",         // "comptype_cec",                     // xXeE
		"x1c1s1",       // "comptype_compmod",                 // xXcCsS
		"x1c1r1",       // "comptype_rtrmod",                  // xXcCrR
		"x1c1s1b1",     // "comptype_ncard",                   // xXcCsSbB
		"x1c1s1b1i1",   // "comptype_bmc_nic",                 // xXcCsSbBiI
		"x1c1s1e1",     // "comptype_node_enclosure",          // xXcCsSeE
		"x1c1s1v1",     // "comptype_compmod_power_connector", // xXcCsSvV
		"x1c1s1b1n1",   // "comptype_node",                    // xXcCsSbBnN
		"x1c1s1b1n1p1", // "comptype_node_processor",          // xXcCsSbBnNpP
		"x1c1s1b1n1i1", // "comptype_node_nic",                // xXcCsSbBnNiI
		"x1c1s1b1n1h1", // "comptype_node_hsn_nic",            // xXcCsSbBnNhH
		"x1c1s1b1n1d1", // "comptype_dimm",                    // xXcCsSbBnNdD
		"x1c1s1b1n1a1", // "comptype_node_accel",              // xXcCsSbBnNaA
		"x1c1s1b1f0",   // "comptype_node_fpga",               // xXcCsSbBfF
		"x1c1r1a1",     // "comptype_hsn_asic",                // xXcCrRaA
		"x1c1r1f1",     // "comptype_rtr_fpga",                // xXcCrRfF

		"x1c1r1b1",   // "comptype_rtr_bmc",                 // xXcCrRbB
		"x1c1r1b1i1", // "comptype_rtr_bmc_nic",             // xXcCrRbBiI

		"x1c1r1e1",   // "comptype_hsn_board",             // xXcCrReE
		"x1c1r1a1l1", // "comptype_hsn_link",              // xXcCrRaAlL
		"x1c1r1j1",   // "comptype_hsn_connector",         // xXcCrRjJ
		"x1c1r1j1p1", // "comptype_hsn_connector_port",    // xXcCrRjJpP
		"x1c1w1",     // "comptype_mgmt_switch",           // xXcCwW
		"x1c1w1j1",   // "comptype_mgmt_switch_connector", // xXcCwWjJ
	}

	for _, xname := range tests {
		suite.True(base.IsHMSCompIDValid(xname), xname)
		hmsType := base.GetHMSType(xname)
		slsType := sls_common.HMSTypeToHMSStringType(hmsType)

		suite.NotEqual("INVALID", slsType)

		h := sls_common.GenericHardware{
			Parent:     base.GetHMSCompParent(xname),
			Xname:      xname,
			Class:      sls_common.ClassRiver,
			Type:       slsType,
			TypeString: hmsType,
		}

		payload, err := json.Marshal(h)
		suite.NoError(err)

		suite.T().Log(string(payload))

		putURL := hwURLBase + "/" + xname
		req, preqerr := http.NewRequest("PUT", putURL, bytes.NewBuffer(payload))
		suite.NoError(preqerr, "creating http POST request")

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)

		//Check response code
		suite.Equal(http.StatusOK, response.Code, "Response: %s", response.Body.String())
	}
}

func TestHardwareTestSuite(t *testing.T) {
	suite.Run(t, new(HardwareTestSuite))
}
