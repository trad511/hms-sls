#! /bin/sh

pushd .

cd $(dirname $0)/../..

BLDID=$(docker build -q .)

RUNID=$(docker run -d -p 8080:8376 $BLDID)

STATUS=$(curl -s -o /dev/null -w "%{http_code}" localhost:8080)

if [ "$STATUS" -ne "404" ]
then
    echo "FAIL: sls base HTTP response failed"
else
    echo "PASS: sls base response is as expected"
fi

docker stop $RUNID

popd