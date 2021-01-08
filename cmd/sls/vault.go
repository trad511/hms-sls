/*
 * Copyright 2019 Cray Inc. All Rights Reserved.
 *
 * Except as permitted by contract or express written permission of Cray Inc.,
 * no part of this work or its content may be modified, used, reproduced or
 * disclosed in any form. Modifications made without express permission of
 * Cray Inc. may damage the system the software is installed within, may
 * disqualify the user from receiving support from Cray Inc. under support or
 * maintenance contracts, or require additional support services outside the
 * scope of those contracts to repair the software or system.
 *
 */

package main

import (
	"log"
	compcredentials "stash.us.cray.com/HMS/hms-compcredentials"
	securestorage "stash.us.cray.com/HMS/hms-securestorage"
	"time"
)

func setupVault() {
	log.Printf("DEBUG: Connecting to Vault...\n")

	for Running {
		// Start a connection to Vault
		if secureStorage, err := securestorage.NewVaultAdapter(""); err != nil {
			log.Printf("Unable to connect to Vault, err: %s! Trying again in 1 second...\n", err)
			time.Sleep(1 * time.Second)
		} else {
			log.Printf("INFO: Connected to Vault.\n")

			compCredStore = *compcredentials.NewCompCredStore(vaultKeypath, secureStorage)
			break
		}
	}
}
