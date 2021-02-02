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
