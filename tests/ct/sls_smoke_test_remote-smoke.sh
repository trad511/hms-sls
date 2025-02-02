#!/bin/bash -l
#
# MIT License
#
# (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
###############################################################
#
#     CASM Test - Cray Inc.
#
#     TEST IDENTIFIER   : sls_smoke_test
#
#     DESCRIPTION       : Automated test for verifying basic SLS API
#                         infrastructure and installation on Cray
#                         Shasta systems.
#                         
#     AUTHOR            : Mitch Schooler
#
#     DATE STARTED      : 09/23/2020
#
#     LAST MODIFIED     : 03/30/2021
#
#     SYNOPSIS
#       This is a smoke test for the HMS SLS API that makes basic HTTP
#       requests using curl to verify that the service's API endpoints
#       respond and function as expected after an installation.
#
#     INPUT SPECIFICATIONS
#       Usage: sls_smoke_test
#       
#       Arguments: None
#
#     OUTPUT SPECIFICATIONS
#       Plaintext is printed to stdout and/or stderr. The script exits
#       with a status of '0' on success and '1' on failure.
#
#     DESIGN DESCRIPTION
#       This smoke test is based on the Shasta health check srv_check.sh
#       script in the CrayTest repository that verifies the basic health of
#       various microservices but instead focuses exclusively on the SLS
#       API. It was implemented to run from the ct-portal container off
#       of the NCN of the system under test within the DST group's Continuous
#       Testing (CT) framework as part of the remote-smoke test suite.
#
#     SPECIAL REQUIREMENTS
#       Must be executed from the ct-portal container on a remote host
#       (off of the NCNs of the test system) with the Continuous Test
#       infrastructure installed.
#
#     UPDATE HISTORY
#       user       date         description
#       -------------------------------------------------------
#       schooler   09/23/2020   initial implementation
#       schooler   03/30/2021   add check_job_status test
#
#     DEPENDENCIES
#       - hms_smoke_test_lib_ncn-resources_remote-resources.sh which is
#         expected to be packaged in
#         /opt/cray/tests/remote-resources/hms/hms-test in the ct-portal
#         container.
#
#     BUGS/LIMITATIONS
#       None
#
###############################################################

# HMS test metrics test cases: 9
# 1. Check cray-sls pod statuses
# 2. Check cray-sls job statuses
# 3. GET /version API response code
# 4. GET /health API response code
# 5. GET /liveness API response code
# 6. GET /readiness API response code
# 7. GET /hardware API response code
# 8. GET /networks API response code
# 9. GET /dumpstate API response code

# initialize test variables
TEST_RUN_TIMESTAMP=$(date +"%Y%m%dT%H%M%S")
TEST_RUN_SEED=${RANDOM}
OUTPUT_FILES_PATH="/tmp/sls_smoke_test_out-${TEST_RUN_TIMESTAMP}.${TEST_RUN_SEED}"
SMOKE_TEST_LIB="/opt/cray/tests/remote-resources/hms/hms-test/hms_smoke_test_lib_ncn-resources_remote-resources.sh"
CURL_ARGS="-k -i -s -S"
MAIN_ERRORS=0
CURL_COUNT=0

# cleanup
function cleanup()
{
    echo "cleaning up..."
    rm -f ${OUTPUT_FILES_PATH}*
}

# main
function main()
{
    AUTH_ARG="-H \"Authorization: Bearer $TOKEN\""

    # GET tests
    for URL_ARGS in \
        "apis/sls/v1/version" \
        "apis/sls/v1/health" \
        "apis/sls/v1/liveness" \
        "apis/sls/v1/readiness" \
        "apis/sls/v1/hardware" \
        "apis/sls/v1/networks" \
        "apis/sls/v1/dumpstate"
    do
        URL=$(url "${URL_ARGS}")
        URL_RET=$?
        if [[ ${URL_RET} -ne 0 ]] ; then
            cleanup
            exit 1
        fi
        run_curl "GET ${AUTH_ARG} ${URL}"
        if [[ $? -ne 0 ]] ; then
            ((MAIN_ERRORS++))
        fi
    done

    echo "MAIN_ERRORS=${MAIN_ERRORS}"
    return ${MAIN_ERRORS}
}

# check_pod_status
function check_pod_status()
{
    run_check_pod_status "cray-sls"
    return $?
}

# check_job_status
function check_job_status()
{
    run_check_job_status "cray-sls"
    return $?
}

# TARGET_SYSTEM is expected to be set in the ct-portal container
if [[ -z ${TARGET_SYSTEM} ]] ; then
    >&2 echo "ERROR: TARGET_SYSTEM environment variable is not set"
    cleanup
    exit 1
else
    echo "TARGET_SYSTEM=${TARGET_SYSTEM}"
    TARGET="auth.${TARGET_SYSTEM}"
    echo "TARGET=${TARGET}"
fi

# TOKEN is expected to be set in the ct-portal container
if [[ -z ${TOKEN} ]] ; then
    >&2 echo "ERROR: TOKEN environment variable is not set"
    cleanup
    exit 1
else
    echo "TOKEN=${TOKEN}"
fi

trap ">&2 echo \"recieved kill signal, exiting with status of '1'...\" ; \
    cleanup ; \
    exit 1" SIGHUP SIGINT SIGTERM

# source HMS smoke test library file
if [[ -r ${SMOKE_TEST_LIB} ]] ; then
    . ${SMOKE_TEST_LIB}
else
    >&2 echo "ERROR: failed to source HMS smoke test library: ${SMOKE_TEST_LIB}"
    exit 1
fi

# make sure filesystem is writable for output files
touch ${OUTPUT_FILES_PATH}
if [[ $? -ne 0 ]] ; then
    >&2 echo "ERROR: output file location not writable: ${OUTPUT_FILES_PATH}"
    cleanup
    exit 1
fi

echo "Running sls_smoke_test..."

# run initial pod status test
check_pod_status
if [[ $? -ne 0 ]] ; then
    echo "FAIL: sls_smoke_test ran with failures"
    cleanup
    exit 1
fi

# run initial job status test
check_job_status
if [[ $? -ne 0 ]] ; then
    echo "FAIL: sls_smoke_test ran with failures"
    cleanup
    exit 1
fi

# run main API tests
main
if [[ $? -ne 0 ]] ; then
    echo "FAIL: sls_smoke_test ran with failures"
    cleanup
    exit 1
else
    echo "PASS: sls_smoke_test passed!"
    cleanup
    exit 0
fi
