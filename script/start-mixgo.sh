#!/bin/bash
#
# Startup script using the current ip address/net to determine matching
# configuration file and lauch the application
#
###
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd ${SCRIPT_DIR}/..
WORKDIR=$(pwd)
echo "Working directory ${WORKDIR}"

CFG='/opt/mixgo/config/config-qu.yml'

if [ "$(ip addr list | grep wlan0 | grep 10.10.10.)" ]
then
    CFG='/opt/mixgo/config/config-xr.yml'
fi

echo "Using config: ${CFG}"

./mixgo ${CFG}
