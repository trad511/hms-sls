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
	"fmt"

	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-sls/internal/database"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

var InvalidExtraProperties = errors.New("extra properties does not match expected format")
var InvalidXname = errors.New("xname is invalid")
var InvalidXnameType = errors.New("xname type is invalid")
var InvalidClass = errors.New("class is invalid")
var UnsupportedType = errors.New("type can not be stored in SLS")
var UnknownType = errors.New("type is unknown")

func validateXname(xname string) error {
	xnameType := base.GetHMSType(xname)
	if xnameType == base.HMSTypeInvalid {
		return InvalidXname
	}

	invalidTypes := map[base.HMSType]struct{}{
		base.Partition:      {},
		base.HMSTypeAll:     {},
		base.HMSTypeAllComp: {},
		base.HMSTypeAllSvc:  {},
		base.HMSTypeInvalid: {},
	}
	_, isInvalid := invalidTypes[xnameType]
	if isInvalid {
		return InvalidXnameType
	}

	return nil
}

func validateClass(class sls_common.CabinetType) error {
	if class != "" &&
		class != sls_common.ClassRiver &&
		class != sls_common.ClassMountain &&
		class != sls_common.ClassHill {
		return InvalidClass
	}

	return nil
}

func validateType(typeObj sls_common.HMSStringType) error {
	switch typeObj {
	case sls_common.HMSTypeAll, sls_common.HMSTypeAllComp, sls_common.HMSTypeAllSvc, sls_common.HMSTypeInvalid, sls_common.Partition:
		return UnsupportedType

	/* Items in this section have specific properties that require validation */
	case sls_common.CabinetPDUPowerConnector:
	case sls_common.NodePowerConnector:
		//PoweredBy
	case sls_common.HSNConnector:
		// NodeNics
	case sls_common.MgmtSwitch:
		//IP6addr
		//IP4addr
		//Username
		//Password
	case sls_common.MgmtSwitchConnector:
		// NodeNics
		// VendorName
	case sls_common.MgmtHLSwitch:
	case sls_common.CDUMgmtSwitch:
	case sls_common.RouterBMC:
		//IP6addr
		//IP4addr
		//Username
		//Password
	case sls_common.RouterBMCNic:
	case sls_common.CabinetPDUNic:
	case sls_common.NodeBMCNic:
	case sls_common.NodeHsnNIC:
	case sls_common.NodeNIC:
	case sls_common.RouterModule:
	case sls_common.ComputeModule:
	case sls_common.Node:
	case sls_common.NodeBMC:

	/* These all have no specific properties that need validation */
	/* for these, do nothing */
	case sls_common.CDU:
	case sls_common.CEC:
	case sls_common.CMMFpga:
	case sls_common.CMMRectifier:
	case sls_common.Cabinet:
	case sls_common.CabinetCDU:
	case sls_common.CabinetPDU:
	case sls_common.CabinetPDUController:
	case sls_common.CabinetPDUOutlet:
	case sls_common.Chassis:
	case sls_common.ChassisBMC:
	case sls_common.HSNAsic:
	case sls_common.HSNBoard:
	case sls_common.HSNConnectorPort:
	case sls_common.HSNLink:
	case sls_common.Memory:
	case sls_common.NodeAccel:
	case sls_common.NodeEnclosure:
	case sls_common.NodeFpga:
	case sls_common.Processor:
	case sls_common.RouterFpga:
	case sls_common.RouterTORFpga:
	case sls_common.SMSBox:
	case sls_common.System:

	/* Finally, default to "no good" */
	default:
		return UnknownType
	}

	return nil
}

// ReplaceGenericHardware will in a single transaction remove all hardware from the database and subsequently insert
// all of the provided hardware in its place. This make this a safe function to use for any bulk load operations.
func ReplaceGenericHardware(hardware []sls_common.GenericHardware) error {
	return database.ReplaceAllGenericHardware(hardware)
}

func SearchGenericHardware(searchHardware sls_common.GenericHardware) (returnHardware []sls_common.GenericHardware, err error) {
	conditions := make(map[string]string)

	// Build conditions map.
	if searchHardware.Xname != "" {
		err = validateXname(searchHardware.Xname)
		if err != nil {
			return
		}

		conditions["xname"] = searchHardware.Xname
	}
	if searchHardware.Parent != "" {
		parentErr := validateXname(searchHardware.Parent)
		if parentErr != nil {
			err = fmt.Errorf("invalid parent: %s", parentErr)
			return
		}

		conditions["parent"] = searchHardware.Parent
	}
	if searchHardware.Type != "" {
		err = validateType(searchHardware.Type)
		if err != nil {
			return
		}

		conditions["comp_type"] = string(searchHardware.Type)
	}
	if searchHardware.Class != "" {
		err = validateClass(searchHardware.Class)
		if err != nil {
			return
		}

		conditions["comp_class"] = string(searchHardware.Class)
	}

	propertiesMap, ok := searchHardware.ExtraPropertiesRaw.(map[string]interface{})
	if !ok {
		err = InvalidExtraProperties
		return
	}

	returnHardware, err = database.SearchGenericHardware(conditions, propertiesMap)

	return
}
