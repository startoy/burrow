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

echo
echo "Marmots send data to contract"
echo "BURROW_CLIENT_CONTRACT_ADDRESS: " $BURROW_CLIENT_CONTRACT_ADDRESS


monax pkgs do -c batch -a $BURROW_CLIENT_ADDRESS -f query.yaml --set STORE=$STORE
