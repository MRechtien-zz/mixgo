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

while ! ip -force -4 addr show wlan0 | grep -q inet; do
    echo "Waiting for wlan0 to be up"
    sleep 3
done

CFG='config/config-qu.yml'
if ip addr list | grep wlan0 | grep -q 10.10.10.; then
    CFG='config/config-xr.yml'
fi

echo "Using config: ${CFG}"
./mixgo ${CFG}
