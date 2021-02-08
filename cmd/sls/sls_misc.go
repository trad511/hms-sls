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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"reflect"
	"stash.us.cray.com/HMS/hms-sls/pkg/sls-common"

	compcredentials "stash.us.cray.com/HMS/hms-compcredentials"
	"stash.us.cray.com/HMS/hms-sls/internal/database"

	base "stash.us.cray.com/HMS/hms-base"
	"stash.us.cray.com/HMS/hms-sls/internal/datastore"
)

// Used for response functions

type Response struct {
	C int    `json:"code"`
	M string `json:"message"`
}

var mapVersion int
var mapTimestamp string

const (
	SLS_VERSION_KEY = "slsVersion"
	SLS_HEALTH_KEY  = "SLS_HEALTH_KEY"
	SLS_HEALTH_VAL  = "SLS_OK"
)

// Fetch the version info from the DB.

func getVersionFromDB() (version sls_common.SLSVersion, err error) {
	currentVersion, err := database.GetCurrentVersion()
	if err != nil {
		log.Println("ERROR: Can't get current version:", err)
		return
	}

	lastModified, err := database.GetLastModified()
	if err != nil {
		log.Println("ERROR: Can't get last modified:", err)
		return
	}

	version = sls_common.SLSVersion{
		Counter:     currentVersion,
		LastUpdated: lastModified,
	}

	return
}

// Check if the database is ready.

func dbReady() bool {
	_, serr := getVersionFromDB()
	if serr != nil {
		log.Println("INFO: Readiness check failed, can't get version info from DB:", serr)
		return false
	}

	return true
}

// /verion API: Get the current version information.

func doVersionGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	// Grab the version info from the DB

	slsVersion, slserr := getVersionFromDB()
	if slserr != nil {
		log.Println("ERROR: Can't get version info from DB:", slserr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Unable to get version info from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	ba, err := json.Marshal(slsVersion)
	if err != nil {
		log.Println("ERROR: JSON marshal of version info failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// HealthResponse - used to report service health stats
type HealthResponse struct {
	Vault        string `json:"Vault"`
	DBConnection string `json:"DBConnection"`
}

func doHealthGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is provided as a debugging aid for administrators to
	//  find out what is going on with the system.  This should return
	//  information in a human-readable format that will help to
	//  determine the state of this service.

	log.Printf("INFO: entering health check")
	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	var stats HealthResponse

	// Check the vault
	// NOTE: may be dangerous to check something in the vault - needed?
	if vaultEnabled {
		if compCredStore.SS == nil {
			log.Printf("INFO: Vault enabled but not initialized")
			stats.Vault = "Enabled but not initialized"
		} else {
			log.Printf("INFO: Vault enabled and initialized")
			stats.Vault = "Enabled and initialized"
		}
	} else {
		log.Printf("INFO: Vault not enabled")
		stats.Vault = "Not enabled"
	}

	//Check that ETCD/DB connection is available
	if database.DB == nil {
		log.Printf("INFO: DB not initialized")
		stats.DBConnection = "Not Initialized"
	} else {
		// NOTE - the Ping command will restore a dropped connection
		dberr := database.DB.Ping()
		if dberr != nil {
			log.Printf("INFO: DB ping error:%s", dberr.Error())
			stats.DBConnection = fmt.Sprintf("Ping error:%s", dberr.Error())
		} else if dbReady() == false {
			// active query from something in database
			log.Printf("INFO: DB not Ready")
			stats.DBConnection = "Not Ready"
		} else {
			log.Printf("INFO: DB Ready")
			stats.DBConnection = "Ready"
		}
	}

	// marshal and send the response
	ba, err := json.Marshal(stats)
	if err != nil {
		log.Println("ERROR: JSON marshal of readiness info failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doLivenessGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is coded in accordance with kubernetes best practices
	//  for liveness/readiness checks.  This function should only be
	//  used to indicate the server is still alive and processing requests.

	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path, r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	w.WriteHeader(http.StatusNoContent)
}

func doReadinessGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is coded in accordance with kubernetes best practices
	//  for liveness/readiness checks.  This function should only be
	//  used to indicate if something is wrong with this service that
	//  prevents usage.  If this fails too many times, the instance
	//  will be killed and re-started.  Only fail this if restarting
	//  this service is likely to fix the problem.

	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	ready := true
	if dbReady() == false {
		log.Printf("INFO: readiness check fails, db not ready")
		ready = false
	}

	if ready {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

//  /dumpstate API

func doDumpState(w http.ResponseWriter, r *http.Request) {
	ret := sls_common.SLSState{
		Hardware: make(map[string]sls_common.GenericHardware),
		Networks: make(map[string]sls_common.Network),
	}
	var shaHash hash.Hash
	var publicKey *rsa.PublicKey

	allHardware, err := datastore.GetAllHardware()
	if err != nil {
		log.Println("ERROR: unable to get hardware: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to get hardware",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	// Only go to the trouble of getting the public key if it was POST'd.
	if r.Method == "POST" {
		// Check to see if we've been given a key to encrypt with.
		publicKeyFile, _, publicKeyErr := r.FormFile("public_key")
		if publicKeyErr == http.ErrMissingFile {
			log.Println("WARNING: Public key not provided, not encrypting or providing any Vault data")
		} else if publicKeyErr != nil {
			log.Println("ERROR: Unable to parse public key form file: ", publicKeyErr)
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				"Unable to parse public key form file",
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			return
		} else {
			var buf bytes.Buffer
			io.Copy(&buf, publicKeyFile)
			publicKeyString := buf.String()
			publicKeyFile.Close()

			if publicKeyString == "" {
				log.Println("ERROR: POST to dumpState with blank public key")
				pdet := base.NewProblemDetails("about: blank",
					"Bad Request",
					"Public key must be included as form data when POSTing to dumpState",
					r.URL.Path, http.StatusBadRequest)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			shaHash = sha256.New()

			block, _ := pem.Decode([]byte(publicKeyString))
			if block == nil {
				log.Println("ERROR: unable to decode public key")
				pdet := base.NewProblemDetails("about: blank",
					"Unsupported Media Type",
					"Failed to decode public key",
					r.URL.Path, http.StatusUnsupportedMediaType)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			publicKeyInterface, parseErr := x509.ParsePKIXPublicKey(block.Bytes)
			if parseErr != nil {
				log.Println("ERROR: unable to parse public key using the x509.ParsePKIXPublicKey method")
				pdet := base.NewProblemDetails("about: blank",
					"Unsupported Media Type",
					"Failed to parse public key",
					r.URL.Path, http.StatusUnsupportedMediaType)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			publicKey = publicKeyInterface.(*rsa.PublicKey)
		}
	}

	for _, hardware := range allHardware {
		if vaultEnabled && publicKey != nil {
			credentials, credErr := compCredStore.GetCompCred(hardware.Xname)
			if credErr != nil {
				log.Println("ERROR: unable to get credentials for hardware:", credErr)
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to get credentials for hardware",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			// Ensure there is actually something in Vault and we didn't just get back an empty structure.
			if reflect.DeepEqual(credentials, compcredentials.CompCredentials{}) {
				hardware.VaultData = nil
			} else {
				// Marshal just the Vault credentials JSON into a byte array for encrypting.
				vaultBytes, marshalErr := json.Marshal(credentials)
				if marshalErr != nil {
					log.Println("ERROR: unable to marshal credentials:", marshalErr)
					pdet := base.NewProblemDetails("about: blank",
						"Internal Server Error",
						"Failed to marshal credentials",
						r.URL.Path, http.StatusInternalServerError)
					base.SendProblemDetails(w, pdet, 0)
					return
				}

				// Encrypt using the public key given to us.
				encryptedBytes, encryptErr := rsa.EncryptOAEP(shaHash, rand.Reader, publicKey, vaultBytes, nil)
				if encryptErr != nil {
					log.Println("ERROR: unable to encrypt credentials:", encryptErr)
					pdet := base.NewProblemDetails("about: blank",
						"Internal Server Error",
						"Failed to encrypt credentials",
						r.URL.Path, http.StatusInternalServerError)
					base.SendProblemDetails(w, pdet, 0)
					return
				}

				// Now create a base64 representation of the encrypted bytes.
				base64EncryptedStringEncoded := base64.StdEncoding.EncodeToString(encryptedBytes)

				hardware.VaultData = base64EncryptedStringEncoded
			}
		}

		ret.Hardware[hardware.Xname] = hardware
	}

	allNetworks, err := datastore.GetAllNetworks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Printf("ERROR: Unable to load existing networks: %s", err)
		return
	}
	for _, network := range allNetworks {
		ret.Networks[network.Name] = network
	}

	jsonBytes, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Printf("ERROR: Unable to create json: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

//  /loadstate API

func doLoadState(w http.ResponseWriter, r *http.Request) {
	var inputData sls_common.SLSState
	var shaHash hash.Hash
	var privateKey *rsa.PrivateKey
	var buf bytes.Buffer

	// Check to see if we've been given a key to encrypt with.
	privateKeyFile, _, privateKeyErr := r.FormFile("private_key")
	if privateKeyErr == http.ErrMissingFile {
		log.Println("WARNING: No private key provided, ignoring any encrypted blocks.")
	} else if privateKeyErr != nil {
		log.Println("ERROR: Unable to parse private key form file: ", privateKeyErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Unable to parse private key form file",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else {
		io.Copy(&buf, privateKeyFile)
		privateKeyString := buf.String()
		privateKeyFile.Close()
		buf.Reset()

		if privateKeyString == "" {
			log.Println("ERROR: POST to dumpState with blank private key")
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				"Private key must be included as form data when POSTing to loadState",
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			return
		}

		// Now just need to convert the private key string into a RSA private key.
		block, _ := pem.Decode([]byte(privateKeyString))
		if block == nil {
			log.Println("ERROR: unable to decode private key")
			pdet := base.NewProblemDetails("about: blank",
				"Unsupported Media Type",
				"Failed to decode private key",
				r.URL.Path, http.StatusUnsupportedMediaType)
			base.SendProblemDetails(w, pdet, 0)
			return
		}

		// This requires an RSA generated private key. Best to use openssl for it:
		// openssl rsa -in private.pem -outform PEM -pubout -out public.pem
		privateKeyInterface, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			log.Println("ERROR: unable to parse private key")
			pdet := base.NewProblemDetails("about: blank",
				"Unsupported Media Type",
				"Failed to parse private key",
				r.URL.Path, http.StatusUnsupportedMediaType)
			base.SendProblemDetails(w, pdet, 0)
			return
		}

		privateKey = privateKeyInterface.(*rsa.PrivateKey)

	}

	// Now get the config file to read back in.
	configFile, _, configFileErr := r.FormFile("sls_dump")
	if configFileErr != nil {
		log.Println("ERROR: Unable to parse SLS dump form file: ", configFileErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Unable to parse SLS dump form file",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	io.Copy(&buf, configFile)
	configFile.Close()

	// Read the body, let's turn it into back into JSON
	marshalErr := json.Unmarshal(buf.Bytes(), &inputData)
	if marshalErr != nil {
		log.Println("ERROR: Unable to unmarshal config file: ", marshalErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Unable to unmarshal config file",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	// Finally we are ready to put the info back into the database.
	shaHash = sha256.New()
	var hardware []sls_common.GenericHardware
	var networks []sls_common.Network

	// Loop through all of the provided hardware looking for those with Vault details that need to go back.
	for _, obj := range inputData.Hardware {
		if vaultEnabled && obj.VaultData != nil && privateKey != nil {
			// Decode the base64 encoded string.
			base64EncryptedStringDecoded, decodeErr := base64.StdEncoding.DecodeString(obj.VaultData.(string))
			if decodeErr != nil {
				log.Println("ERROR: unable to decode credentials:", decodeErr)
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to decode credentials",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			credentialsDecryptedBytes, decryptErr := rsa.DecryptOAEP(shaHash, rand.Reader, privateKey,
				base64EncryptedStringDecoded, nil)
			if decryptErr != nil {
				log.Println("ERROR: unable to decrypt credentials:", decryptErr)
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to decrypt credentials",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			// Get the credentials back into their struct form.
			var credentials compcredentials.CompCredentials
			unmarshalErr := json.Unmarshal(credentialsDecryptedBytes, &credentials)
			if unmarshalErr != nil {
				log.Println("ERROR: unable to unmarshal credentials:", unmarshalErr)
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to unmarshal credentials",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			// Now finally we can put the credentials back into Vault.
			compCredErr := compCredStore.StoreCompCred(credentials)
			if compCredErr != nil {
				log.Println("ERROR: unable to store credentials:", compCredErr)
				pdet := base.NewProblemDetails("about: blank",
					"Internal Server Error",
					"Failed to store credentials",
					r.URL.Path, http.StatusInternalServerError)
				base.SendProblemDetails(w, pdet, 0)
				return
			}
		}

		hardware = append(hardware, obj)
	}

	hardwareErr := datastore.ReplaceGenericHardware(hardware)
	if hardwareErr != nil {
		log.Println("ERROR: unable to replace hardware:", hardwareErr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to replace hardware",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	for _, obj := range inputData.Networks {
		networks = append(networks, obj)
	}

	networksErr := datastore.ReplaceAllNetworks(networks)
	if networksErr != nil {
		log.Println("ERROR: unable to replace networks:", networksErr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to replace networks",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Send a JSON response.  If the ecode indicates an error, send
// a properly formatted RFC7807 problem.
// If it does not, fall back to the original CAPMC format, which will
// now just be used for success cases or odd HTTP status codes that
// don't suggest a RFC7807 problem response.
// We use the 7807 problem format for 4xx and 5xx errors, though
// in practice the latter (server errors) will probably not be used here
// as they do not describe invalid requests but server-specific issues.

func sendJsonRsp(w http.ResponseWriter, ecode int, message string) {
	if ecode < 400 {
		sendJsonResponse(w, ecode, message)
	} else {
		// Use library function in HMS base.  Problem will be
		// a generic one with title matching the HTTP Status code text
		// with message as the details field.  For this type of problem
		// title can just be set to "about:blank" so we need no
		// custom URL.  The optional instance field is omitted as well
		// so no URL/URI is needed there either.
		base.SendProblemDetailsGeneric(w, ecode, message)
	}
}

// Send a simple message for cases where need a non-error response.  If
// a more feature filled message needs to be returned then do it with a
// different function.  Code is the http status response, converted to
// zero for success-related responses.
func sendJsonResponse(w http.ResponseWriter, ecode int, message string) {
	// if the HTTP call was a success then put zero in the returned json
	// error field. This is what capmc does.
	http_code := ecode
	if ecode >= 200 && ecode <= 299 {
		ecode = 0
	}
	data := Response{ecode, message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if http_code != http.StatusNoContent {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Printf("Yikes, I couldn't encode a JSON status response: %s\n", err)
		}
	}
}

func sendJsonCompRsp(w http.ResponseWriter, comp sls_common.GenericHardware) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(comp)
	if err != nil {
		log.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

func sendJsonCompArrayRsp(w http.ResponseWriter, comps []sls_common.GenericHardware) {
	http_code := 200
	if len(comps) == 0 {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if len(comps) == 0 {
		err := json.NewEncoder(w).Encode(comps)
		if err != nil {
			log.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}
