# Batch

pre-requisits
- monax v0.16
- install jq, glide (go get github.com/Masterminds/glide)

FROM github.com/hyperledger/burrow RUN     
	`git remote add monax git@github.com:monax/burrow`     
	`git fetch monax`     
	`git checkout -b worker-client monax/worker-client`    

(if you don't have hyperledger/burrow, run `go get github.com/hyperledger/burrow`, navigate to that folder, then run the above)


1. `glide install`
2. `make build_client`
3. `cd ./client/batch/test`
4. `monax clean -y` (wipes chains dir !!)
5. `monax chains make batch`
6. `monax chains start batch --init-dir batch/batch_full_000`
7. ./test_0setup.sh -> deploys contract and does initial query: 0 saved
8. ./test_1batch.sh -> reads names.json with 500 names, and uses abi to formulate
        transactions, signs them, sends them to the chain
        afterwards, runs query on contract to validate number of saves
9. repeat 8. should add 500 hits to contract every time
10. use ./test_2query.sh to only query the contract number of hits (without sending new txs)
