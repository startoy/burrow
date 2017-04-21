// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package methods

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/client/rpc"
	"github.com/hyperledger/burrow/definitions"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/txs"
)

type BatchItem struct {
	// inputs      packing.Inputs
	Id          int
	PackedBytes []byte
	Transaction *txs.CallTx
	Result      *rpc.TxResult
	Err         error
}

func CallBatchItems(do *definitions.ClientDo, batchItem <-chan BatchItem) (<-chan BatchItem, <-chan struct{}, error) {
	out := make(chan BatchItem)
	done := make(chan struct{})
	logger, err := loggerFromClientDo(do, "Batch")
	if err != nil {
		return nil, nil, fmt.Errorf("Could not generate logging config from ClientDo: %s", err)
	}

	burrowKeyClient := keys.NewBurrowKeyClient(do.SignAddrFlag, logger)
	burrowNodeClient := client.NewBurrowNodeClient(do.NodeAddrFlag, logger)

	// do checkCommon upfront to anticipate the nonce subsequently.
	publicKey, _, nonce, err := rpc.CheckCommon(burrowNodeClient, burrowKeyClient,
		do.PubkeyFlag, do.AddrFlag, do.AmtFlag, do.NonceFlag)
	if err != nil {
		fmt.Printf("Failed to check node: %s\n", err)
		close(out)
		done <- struct{}{}
		return out, done, err
	}

	publicKeyS := publicKey.KeyString()
	errCount := 0

	go func() {
		for {
			if item, more := <-batchItem; more {
				if item.Err == nil {
					// form the call transaction
					callTransaction, err := rpc.Call(burrowNodeClient, burrowKeyClient,
						publicKeyS, "", do.ToFlag, do.AmtFlag, strconv.FormatInt(nonce, 10),
						do.GasFlag, do.FeeFlag, fmt.Sprintf("%X", item.PackedBytes))
					if err != nil {
						item.Err = fmt.Errorf("Failed on forming Call Transaction: %s", err)
						out <- item
						fmt.Printf("ERROR: %v\n", item.Err)
						errCount += 1
						continue
					}
					txResult, err := rpc.SignAndBroadcast(do.ChainidFlag, burrowNodeClient, burrowKeyClient,
						callTransaction, true, true, false)
					if err != nil {
						item.Transaction = callTransaction
						item.Err = fmt.Errorf("Failed on signing (and broadcasting) transaction: %s", err)
						fmt.Printf("ERROR: %v\n", item.Err)
						errCount += 1
						continue
					}
					item.Transaction = callTransaction
					item.Result = txResult
					fmt.Printf("SAMPLE (%v), CALL NONCE(%v), tx hash: \t\t\t %X\n", item.Id, nonce, item.Result.Hash)
					nonce = nonce + 1
				} else {
					fmt.Printf("ERROR: %v\n", item.Err)
					errCount += 1
				}
			} else {
				fmt.Printf("CLOSED BATCH\n")
				done <- struct{}{}
				return
			}
		}
	}()

	return out, done, nil
}

// // anticipateNonce prevents checkCommon in rpc.Call to get a bad
// // behaviour on the nonce; as we expect to submit multiple transactions
// // per block; we should query the nonce only once and then assume success
// // and update the nonce accordingly.  This strategy should be refined on worker.
// func anticipateNonce(nodeClient client.NodeClient, nonceS string) (int64, error) {
// 	if nonceS == "" {
// 		if nodeClient == nil {
// 			return 0, fmt.Errorf("input must specify a nonce with the --nonce flag or use --node-addr (or BURROW_CLIENT_NODE_ADDR) to fetch the nonce from a node")
// 		}
// 		// fetch nonce from node
// 		account, err2 := nodeClient.GetAccount(addrBytes)
// 		if err2 != nil {
// 			return pub, amt, nonce, err2
// 		}
// 		nonce = int64(account.Sequence) + 1
// 		logging.TraceMsg(nodeClient.Logger(), "Fetch nonce from node",
// 			"nonce", nonce,
// 			"account address", addrBytes,
// 		)
// 	} else {
// 		nonce, err = strconv.ParseInt(nonceS, 10, 64)
// 		if err != nil {
// 			err = fmt.Errorf("nonce is misformatted: %v", err)
// 			return
// 		}
// 	}
// }
