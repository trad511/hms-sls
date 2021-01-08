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

package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type VersionHistoryTestSuite struct {
	suite.Suite
}

func (suite *VersionHistoryTestSuite) SetupSuite() {
	err := NewDatabase()
	if err != nil {
		suite.FailNowf("Unable create database", "err: %s", err)
	}
}

func (suite *VersionHistoryTestSuite) TestVersion_HappyPath() {
	trans, beginErr := DB.Begin()
	if beginErr != nil {
		fmt.Printf("TestVersion_HappyPath: unable to begin transaction: %s", beginErr)
		return
	}

	_, err := IncrementVersion(trans, "foo")
	suite.NoError(err)

	// Now finally we can commit the entire transaction. Assuming this works, we're done here.
	commitErr := trans.Commit()
	if commitErr != nil {
		fmt.Printf("TestVersion_HappyPath: unable to commit transaction: %s", commitErr)
		return
	}

	version, err := GetCurrentVersion()
	suite.NoError(err)

	suite.Greater(version, 0)

	fmt.Printf("\tGot version %d.\n", version)

	lastModified, err := GetLastModified()
	suite.NoError(err)

	fmt.Printf("\tGot last modified %s.\n", lastModified)
}

func TestVersionHistorySuite(t *testing.T) {
	suite.Run(t, new(VersionHistoryTestSuite))
}
