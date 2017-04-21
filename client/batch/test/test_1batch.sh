#!/usr/bin/env bash
set -e

name=batch
chains_dir=$HOME/.monax/chains
name_full="$name"_full_000
chain_dir="$chains_dir"/"$name"
CHAIN_ID=$name
echo "$chain_dir"

export BURROW_CLIENT_ADDRESS=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
echo "Marmots will act from full account"
echo "BURROW_CLIENT_ADDRESS" $BURROW_CLIENT_ADDRESS

# jobs_output gets overwritten by querying; transferred through variable STORE
STORE=$(cat result_query.json | jq .STORE | sed -e 's/^"//' -e 's/"$//')
export BURROW_CLIENT_CONTRACT_ADDRESS=$STORE

echo
echo "Marmots send data to contract"
echo "BURROW_CLIENT_CONTRACT_ADDRESS: " $BURROW_CLIENT_CONTRACT_ADDRESS

COMMIT_SHA=$(echo `git rev-parse --short --verify HEAD`)

export BURROW_CLIENT_ABI_PATH="$(pwd)/abi/Store"
export BURROW_CLIENT_METHOD="saveItem"
export BURROW_CLIENT_JSON_DATA_PATH="$(pwd)/names.json"

# get the client executable
client=../../../target/burrow-client-$COMMIT_SHA
echo "$client"

# run the batch command (note that the env vars provide the input)
$client batch -c batch -a $BURROW_CLIENT_ADDRESS --contract $BURROW_CLIENT_CONTRACT_ADDRESS

sleep 2

monax pkgs do -c batch -a $BURROW_CLIENT_ADDRESS -f query.yaml --set STORE=$STORE
