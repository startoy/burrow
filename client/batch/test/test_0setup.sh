#!/usr/bin/env bash
set -e

name=batch
chains_dir=$HOME/.monax/chains
name_full="$name"_full_000
chain_dir="$chains_dir"/"$name"
CHAIN_ID=$name
echo "$chain_dir"

# monax chains make batch
# monax chains start batch --init-dir batch/batch_full_000

BURROW_CLIENT_ADDRESS=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
echo "Marmots will act from full account"
echo "BURROW_CLIENT_ADDRESS" $BURROW_CLIENT_ADDRESS

monax pkgs do -c $CHAIN_ID -a $BURROW_CLIENT_ADDRESS -f epm.yaml
cp jobs_output.json result_deploy.json

STORE=$(cat result_deploy.json | jq .deployStore | sed -e 's/^"//' -e 's/"$//')

echo
echo "Marmots query the number of calls to the store"
echo "STORE: " $STORE

monax pkgs do -c $CHAIN_ID -a $BURROW_CLIENT_ADDRESS -f query.yaml --set STORE=$STORE
cp jobs_output.json	result_query.json

BURROW_CLIENT_ABI_PATH="$(pwd)/abi/Store"
BURROW_CLIENT_METHOD="saveItem"


