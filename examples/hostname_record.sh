#!/bin/bash
# Script tested on ubuntu and os x
# Requires 'jq' (http://stedolan.github.com/jq/)


# Explicitly set PATH to make sure we find all needed binaries
PATH="${PATH}:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Default settings
RRR_URL="https://example.com"
TOKEN="my_secret_token"
SUBDOMAIN=$(hostname)
IFACE="-a"

# Override any defaults with flags
while getopts "u:s:i:t:" FLAG; do
    case $FLAG in
        u)
            RRR_URL=$OPTARG
            ;;
        s)
            SUBDOMAIN=$OPTARG
            ;;
        i)
            IFACE=$OPTARG
            ;;
        t)
            TOKEN=$OPTARG
            ;;
    esac
done

# Get first ip address that does not start with 127
IP=$(ifconfig $IFACE | awk '/inet / {print $2}' | sed -e 's/addr://' | grep -v '^127' | head -n 1)

# Exit if we are missing any information
if [[ -z "$RRR_URL" || -z "$SUBDOMAIN" || -z "$IP" ]]; then
    echo "Missing information"
    exit 1
fi

# Check if there exists a record with the same name and ip
VALUE=$(curl --silent --insecure -H "Token:${TOKEN}" ${RRR_URL}/${SUBDOMAIN} | jq -r .value)

# Exit if it already exist
if [[ "$VALUE" == $IP ]]; then
    echo "Record '$SUBDOMAIN => $IP' already present"
    exit 0
fi

# Add new record
RESULT=$(curl --silent --insecure -H "Token:${TOKEN}" -X POST -d "{\"name\": \"$SUBDOMAIN\", \"value\": \"$IP\"}" ${RRR_URL})
SUCCESS=$(echo "$RESULT" | jq -r ".success")

if [[ $SUCCESS == "true" ]]; then
    echo "Record '$SUBDOMAIN => $IP' added"
else
    echo $(echo "$RESULT" | jq -r ".error")
fi
