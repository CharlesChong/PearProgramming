#!/bin/bash


if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

if [ -n "$(go version | grep 'darwin/amd64')" ]; then    
    GOOS="darwin_amd64"
elif [ -n "$(go version | grep 'linux/amd64')" ]; then
    GOOS="linux_amd64"
else
    echo "FAIL: only 64-bit Mac OS X and Linux operating systems are supported"
    exit 1
fi

# Build the test binary to use to test the student's libstore implementation.
# Exit immediately if there was a compile-time error.
go install pear/tests/servertest
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

go install pear/runners/crunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Pick random ports between [10000, 20000).
# STORAGE_PORT=$(((RANDOM % 10000) + 10000))
# LIB_PORT=$(((RANDOM % 10000) + 10000))

PEAR_SERVER=$GOPATH/bin/servertest
PEAR_CENTRAL=$GOPATH/bin/crunner
#LIB_TEST=$GOPATH/bin/libtest

${PEAR_CENTRAL} &
PEAR_CENTRAL_PID=$!
sleep 3

# Start an instance of the staff's official storage server implementation.
${PEAR_SERVER} & #-port=${STORAGE_PORT} 2 #> /dev/null &
PEAR_SERVER_PID=$!
sleep 3

# Start the test.
# ${LIB_TEST} -port=${LIB_PORT} "localhost:${STORAGE_PORT}"

# Kill the storage server.
kill -9 ${PEAR_SERVER_PID}
wait ${PEAR_SERVER_PID} 2> /dev/null
kill -9 ${PEAR_CENTRAL_PID}
wait ${PEAR_CENTRAL_PID} 2> /dev/null

