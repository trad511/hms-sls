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
	"fmt"
	"path"
	"strings"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-sls/internal/database"
)

const xnameKeyPrefix = "/sls/xnames/"
const nwKeyPrefix = "/sls/networks/"

func makeKeyXname(key string) string {
	if !strings.HasPrefix(key, xnameKeyPrefix) {
		key = path.Join(xnameKeyPrefix, base.NormalizeHMSCompID(key))
	}

	return key
}

func makeKeyNetwork(key string) string {
	if !strings.HasPrefix(key, nwKeyPrefix) {
		key = path.Join(nwKeyPrefix, key)
	}

	return key
}

/*
GetXname retrieves an object by its xname.  Objects are returned as
GenericHardware objects.  Callers should use reflect.GetType if they
wish to have this object as a more specific type, then cast as:
reflect.ValueOf(GenericHardwareObject).Interface().(Type)
Parameters:
 * xname (string): The xname to look up
Returns:
 * *GenericHardware: The object corresponding to that xname or nil if the
   xname was not found.
 * *error: Any error that occurred during the lookup
*/
func GetXname(xname string) (*sls_common.GenericHardware, error) {
	// check if xname exists
	xname = base.NormalizeHMSCompID(xname)
	res, err := database.GetGenericHardwareFromXname(xname)
	if err == database.NoSuch {
		return nil, nil
	}
	return &res, err
}

/*
Reduce all xnames to sane, normalized values
*/
func normalizeCommonFields(obj *sls_common.GenericHardware) error {
	obj.Xname = base.NormalizeHMSCompID(obj.Xname)

	if obj.Parent != "" {
		obj.Parent = base.NormalizeHMSCompID(obj.Parent)
	} else {
		obj.Parent = base.GetHMSCompParent(obj.Xname)
	}

	for i := range obj.Children {
		obj.Children[i] = base.NormalizeHMSCompID(obj.Children[i])
	}

	return nil
}

func normalizeFields(obj sls_common.GenericHardware) (sls_common.GenericHardware, error) {
	// Find old object and merge in, if possible
	// xname must be present
	obj.SetXname(base.NormalizeHMSCompID(obj.GetXname()))

	obj.SetParent(base.NormalizeHMSCompID(obj.GetParent()))

	// Now do the weird ones (where possible)

	return obj, nil
}

/*
Validate that a Hardware object is valid.  Input should previously
have been passed through normalizeFileds()
*/
func validateCommonFields(obj sls_common.GenericHardware) error {
	// xname
	// type
	// class
	// Verify the xname is valid (maps back to a sane type)
	xnameErr := validateXname(obj.GetXname())
	if xnameErr != nil {
		return xnameErr
	}
	xnameType := base.GetHMSType(obj.GetXname())
	if xnameType == base.HMSTypeInvalid {
		return fmt.Errorf("%s: xname %s is invalid", obj.GetXname(), obj.GetXname())
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
		return fmt.Errorf(
			"xname %s represents a type (%s) that cannot be stored",
			obj.GetXname(), xnameType)
	}

	// Check the type field matches the xname
	if sls_common.HMSStringTypeToHMSType(obj.GetType()) != base.GetHMSType(obj.GetXname()) {
		return fmt.Errorf(
			"%s: Type field (%s) does not match type of Xname field (%s)",
			obj.GetXname(), sls_common.HMSStringTypeToHMSType(obj.GetType()), base.GetHMSType(obj.GetXname()))
	}

	// Check the class hint is empty or a valid value
	if obj.GetClass() != "" && obj.GetClass() != sls_common.ClassRiver &&
		obj.GetClass() != sls_common.ClassMountain &&
		obj.GetClass() != sls_common.ClassHill {
		return fmt.Errorf("%s: Class value %s is invalid", obj.GetXname(), obj.GetClass())
	}

	// Check parent
	if obj.GetParent() != base.GetHMSCompParent(obj.GetXname()) {
		return fmt.Errorf(
			"%s: Parent field (%s) does not match derived parent (%s)",
			obj.GetXname(), obj.GetParent(), base.GetHMSCompParent(obj.GetXname()))
	}

	// We don't validate children, because we don't store them/build the list
	// dynamically

	// Verify both type strings are the same
	if obj.GetTypeString() != sls_common.HMSStringTypeToHMSType(obj.GetType()) {
		return fmt.Errorf("%s: Type (%s) and TypeString (%s) do not match",
			obj.GetXname(), obj.GetType(), obj.GetTypeString())
	}

	return nil
}

func validateFields(obj sls_common.GenericHardware) error {
	// First validate common fields:
	err := validateCommonFields(obj)
	if err != nil {
		return err
	}

	// This loooooooong series of switches lets us check each type to verify
	// the input format is correct
	switch obj.GetType() {
	case sls_common.HMSTypeAll, sls_common.HMSTypeAllComp, sls_common.HMSTypeAllSvc, sls_common.HMSTypeInvalid, sls_common.Partition:
		err := fmt.Errorf("%s: An %s object cannot be stored in SLS", obj.GetXname(), obj.GetType())
		return err

	/* Items in this section have specific properties that require validation */
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
	case sls_common.CabinetPDUPowerConnector:
	case sls_common.CDUMgmtSwitch:

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
		err := fmt.Errorf("cannot determine type of object with type field %s", obj.GetType())
		return err
	}

	return nil
}

/*
SetXname updates a specified xname with new or updated properties
*/
func SetXname(xname string, obj sls_common.GenericHardware) error {
	// Setup: make sure all data is clean
	obj, err := normalizeFields(obj)
	if err != nil {
		return err
	}

	err = validateFields(obj)
	if err != nil {
		return err
	}

	// check if xname exists
	_, err = database.GetGenericHardwareFromXname(obj.Xname)
	if err != nil && err != database.NoSuch {
		return err
	} else if err == database.NoSuch {
		err = database.InsertGenericHardware(obj)
	} else {
		err = database.UpdateGenericHardware(obj)
	}

	// TODO If this is a connector object, make sure to update the peer (old and new) as well.

	return err
}

/*
DeleteXname removes hardware witht he appropriate name from the datastore.
It handles updating the parent and any peers.
*/
func DeleteXname(xname string) error {
	// check if xname exists
	_, err := database.GetGenericHardwareFromXname(base.NormalizeHMSCompID(xname))
	if err != nil {
		return err
	}
	gh := sls_common.GenericHardware{}
	gh.Xname = base.NormalizeHMSCompID(xname)
	return database.DeleteGenericHardware(gh)
}

/*
GetAllXnames  gets a list of names of all stored xnames
*/
func GetAllXnames() ([]string, error) {
	ret := make([]string, 0)
	hw, err := database.GetAllGenericHardware()
	if err != nil {
		return nil, err
	}
	for _, gh := range hw {
		ret = append(ret, gh.Xname)
	}
	return ret, nil
}

func GetAllHardware() ([]sls_common.GenericHardware, error) {
	return database.GetAllGenericHardware()
}

/*
GetAllXnameObjects get a list of all stored GenericHardware objects
*/
func GetAllXnameObjects() ([]sls_common.GenericHardware, error) {
	ret, err := database.GetAllGenericHardware()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

/*
ConfigureStorage configures the interface for interacting with the storage module.
Args:
* dstype (string) - the type of database to use.  Always "etcd"
* connInfo (string) - the databse-specific connection information
* args ([]string) - a list of arguments to pass to teh database engine
*/
func ConfigureStorage(dstype string, connInfo string, args []string) error {
	err := database.NewDatabase()
	return err
}
