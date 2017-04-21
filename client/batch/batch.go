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

package batch

import (
	"fmt"

	"github.com/hyperledger/burrow/client/methods"
	"github.com/hyperledger/burrow/client/packing"
	"github.com/hyperledger/burrow/definitions"
)

// read JSON array into named
type BatchedJson []packing.Inputs

func Batch(do *definitions.ClientDo) error {
	// load ABI from file
	abi, err := packing.LoadAbi(do.AbiPathFlag)
	if err != nil {
		return err
	}

	// get Method from ABI
	method, err := packing.GetMethod(abi, do.MethodFlag)
	if err != nil {
		return err
	}

	// load JSON batched inputs
	batch, err := LoadBatch(do.JsonDataPathFlag)
	if err != nil {
		return err
	}

	// pack bytes
	items := make(chan methods.BatchItem)
	go func() {
		for i := 0; i < len(batch); i++ {
			packedBytes, err := packing.PackInputsForMethod(abi, method, batch[i])
			// NOTE: this err can be removed; error passed on and handled later
			if err != nil {
				fmt.Printf("ERROR: %s\n\n%v", err, packedBytes)
			}
			batchItem := methods.BatchItem{
				Id:          i,
				PackedBytes: packedBytes,
				Transaction: nil,
				Result:      nil,
				Err:         err,
			}
			items <- batchItem
		}
		close(items)
	}()

	// start formulating transactions
	_, done, _ := methods.CallBatchItems(do, items)
	<-done
	return nil
}
