#! /usr/bin/env python3

#  MIT License
#
#  (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
#
#  Permission is hereby granted, free of charge, to any person obtaining a
#  copy of this software and associated documentation files (the "Software"),
#  to deal in the Software without restriction, including without limitation
#  the rights to use, copy, modify, merge, publish, distribute, sublicense,
#  and/or sell copies of the Software, and to permit persons to whom the
#  Software is furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included
#  in all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
#  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
#  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
#  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
#  OTHER DEALINGS IN THE SOFTWARE.
#
#  Except as permitted by contract or express written permission of Cray Inc.,
#  no part of this work or its content may be modified, used, reproduced or
#  disclosed in any form. Modifications made without express permission of
#  Cray Inc. may damage the system the software is installed within, may
#  disqualify the user from receiving support from Cray Inc. under support or
#  maintenance contracts, or require additional support services outside the
#  scope of those contracts to repair the software or system.
#

import json
import argparse
import sys
import re
import requests

def getParent(name):
    if re.match(r"^x[0-9]+$", name) != None:
        # Cabinet's parent is the system
        return ""
    m = re.match(r"^(x[0-9].+)[a-z][0-9]+", name)
    return m.group(1)

def getMEDSInfo(file):
    # Open the JSON file and read its contents.
    json_data: dict = {}
    with open(file, "r") as json_file:
        json_data = json.load(json_file)

    if not json_data:
        print("Unable to load JSON data!")
        sys.exit(1)

    cabinet_data = json_data['cabinets']

    retData = {}
    for rack in cabinet_data:
        retData["x" + rack] = {
            "Xname": "x" + rack,
            "Parent": getParent("x" + rack),
            "Class": "Mountain",
            "Type": "comptype_cabinet",
            "TypeString": "Cabinet",
            "ExtraProperties": {
                "Network": "HMN",
                "IP4Base": cabinet_data[rack]['ip'],
                "MACprefix": "02"
            }
        }
    
    return retData

def getMAASInfo(file):
    json_data: dict = {}
    with open(file, "r") as json_file:
        json_data = json.load(json_file)
    
    if not json_data:
        print("Unable to load JSON data!")
        sys.exit(1)

    retData = {}
    for item in json_data:
        retData[item["ID"]] = {
            "Xname": item["ID"],
            "Parent": getParent(item["ID"]),
            "Class": "River",
            "Type": "comptype_ncard",
            "TypeString": "NodeBMC",
            "ExtraProperties": {
                "Network": "HSN",
                "IP6addr": "DHCPv6",
                "IP4addr": item["IPAddress"],
                "Username": item["User"],
                "Password": item["Password"],
            }
        }
    return retData

def getREDSInfo(address, token):
    r = requests.get(address + "/admin/port_xname_map", headers={"Authorization": "Bearer " + token}, verify=False)
    data = json.loads(r.text)

    if not data:
        print("Unable to load JSON data from REDS!")
        sys.exit(1)

    retData = {}
    for switch in data["switches"]:
        # First add the switch
        retData[switch["id"]] = {
            "Xname": switch["id"],
            "Parent": getParent(switch["id"]),
            "Type": "comptype_mgmt_switch",
            "TypeString": "MgmtSwitch",
            "Class": "River",
            "ExtraProperties": {
                "IP6addr": "DHCPv6",
                "IP4addr": switch["address"],
                # Omitted username and password
                "SNMPUsername": switch["snmpUser"],
                "SNMPAuthPassword": switch["snmpAuthPassword"],
                "SNMPAuthProtocol": switch["snmpAuthProtocol"],
                "SNMPPrivPassword": switch["snmpPrivPassword"],
                "SNMPPrivProtocol": switch["snmpPrivProtocol"],
                "Model": ""
            }
        }
        # Then add the ports
        for port in switch["ports"]:
            m = re.match(r"^.*/([0-9]+)", port["ifName"])
            portID = "%sj%s" % (switch["id"], m.group(1))
            retData[portID] = {
                "Xname": portID,
                "Parent": switch["id"],
                "Type": "comptype_mgmt_switch_connector",
                "TypeString": "MgmtSwitchConnector",
                "Class":  "River",
                "ExtraProperties": {
                    "NodeNics": [
                        port["peerID"]
                    ],
                    "VendorName": port["ifName"],
                }
            }

    return retData

def getHSMInfo(address, token, existingData, maasdata):
    r = requests.get(address + "/Defaults/NodeMaps", headers={"Authorization": "Bearer " + token}, verify=False)
    data = json.loads(r.text)

    if not data:
        print("Unable to load JSON data from REDS!")
        sys.exit(1)

    retData = {}
    for key in maasdata:
        item = maasdata[key]
        print(item)
        xname = item["Xname"] + "n0"
        cabID = re.match("^(x[0-9]+)", item["Xname"]).group(1)
        cabClass = "River"
        if cabID in existingData:
            cabClass = existingData[cabID]["Class"]
        retData[xname] = {
            "Xname": xname,
            "Parent": item["Xname"],
            "Type": "comptype_node",
            "TypeString": "Node",
            "Class": cabClass,
            "ExtraProperties": {
                "Role": "Management",
            },
        }

    for item in data["NodeMaps"]:
        role = "Compute"
        if item["ID"] in maasdata or getParent(item["ID"]) in maasdata:
            role="Management"
        if "Role" in item.keys():
            role = item["Role"]
        cabID = re.match("^(x[0-9]+)", item["ID"]).group(1)
        cabClass = "River"
        if cabID in existingData:
            cabClass = existingData[cabID]["Class"]
        retData[item["ID"]] = {
            "Xname": item["ID"],
            "Parent": getParent(item["ID"]),
            "Type": "comptype_node",
            "TypeString": "Node",
            "Class": cabClass,
            "ExtraProperties": {
                "NID": item["NID"],
                "Role": role,
            },
        }

    return retData

def sendToSLS(address, token, data):
    errs = []

    for item in data:
        jdata = json.dumps(data[item])
        r = requests.post(address + "/hardware", data=jdata, headers={"Authorization": "Bearer " + token}, verify=False)
        if r.status_code != requests.codes.okay:
            if r.status_code == requests.codes.conflict:
                print("%s already in SLS" % data[item]["Xname"])
                continue
            errs += [[r.status_code, r.text, data[item]]]
            # but keep going
    return errs

def dumpSLS(address, token, key, outfile):
    with open(key, "r") as key_file:
        data = key_file.read(10240)
        r = requests.post(address + "/dumpstate", files={"public_key": data}, headers={"Authorization": "Bearer " + token}, verify=False)
        res = r.text
    with open(outfile, "w") as o_file:
        o_file.write(res)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--out-file", help="The location to write the output file")
    parser.add_argument("--public-key", help="The public key to use when dumping SLS state.")
    parser.add_argument("--meds-json-file", help="The location of the MEDS description file", default=None)
    parser.add_argument("--maas-data-file", help="The location of the MaaS data file", default=None)
    parser.add_argument("--reds-address", help="The address to use to reach REDS", default="https://api-gw-service-nmn.local/apis/reds/v1")
    parser.add_argument("--hsm-address", help="The address to use to reach HSM", default="https://api-gw-service-nmn.local/apis/smd/hsm/v1")
    parser.add_argument("--sls-address", help="The address to use to reach SLS", default="https://api-gw-service-nmn.local/apis/sls/v1")
    parser.add_argument("--auth-token", help="An authorization token for interacting with the k8s mesh")

    args = parser.parse_args()

    if args.out_file is None:
        print("Output file must be specified!")
        sys.exit(2)

    if args.auth_token is None:
        print("Auth token must be provided to communicate with services")
        sys.exit(2)
    
    if args.public_key is None:
        print("Public key must be provided.  Generate using `openssl genrsa -out private.pem 2048; openssl rsa -in private.pem -outform PEM -pubout -out public.pem`")
        sys.exit(2)
    
    if args.out_file is None:
        print("Please specify an output file for the dump with --out-file")
        sys.exit(2)

    hardware = {}

    medsdata = {}
    if args.meds_json_file is not None:
        medsdata = getMEDSInfo(args.meds_json_file)
    redsdata = getREDSInfo(args.reds_address, args.auth_token)
    hardware = { **hardware, **medsdata, **redsdata }
    maasdata = {}
    if args.maas_data_file is not None:
        maasdata = getMAASInfo(args.maas_data_file)
    hsmdata = getHSMInfo(args.hsm_address, args.auth_token, hardware, maasdata)
    hardware = { **hardware, **hsmdata, **maasdata }

    print(json.dumps(hardware, indent=4, sort_keys=True))

    errs = sendToSLS(args.sls_address, args.auth_token, hardware)

    print("The following errors occurred uploading hardware:")
    for err in errs:
        print("%s: HTTP %s: %s" % (err[2]["Xname"], err[0], err[1]))

    dumpSLS(args.sls_address, args.auth_token, args.public_key, args.out_file)