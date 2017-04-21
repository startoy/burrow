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

// NOTE: this package includes GPLv3 licensed code and CANNOT be published
// under the Apache licensed Hyperledger Burrow code.

package commands

import (
	"github.com/hyperledger/burrow/client/batch"
	"github.com/hyperledger/burrow/util"

	"github.com/spf13/cobra"
)

func buildBatchCommand() *cobra.Command {
	batchCmd := &cobra.Command{
		Use:   "batch",
		Short: "burrow-client batch loads JSON defined data and submits entries as transactions to designated contract.",
		// Long:
		// Example:
		Run: func(cmd *cobra.Command, args []string) {
			if err := batch.Batch(clientDo); err != nil {
				// TODO: [ben] elaborate on error
				util.Fatalf("Failed to process batch: %s", err)
			}
		},
	}
	batchCmd.PersistentFlags().StringVarP(&clientDo.SignAddrFlag, "sign-addr", "", defaultKeyDaemonAddress(), "set monax-keys daemon address (default respects $BURROW_CLIENT_SIGN_ADDRESS)")
	batchCmd.PersistentFlags().StringVarP(&clientDo.NodeAddrFlag, "node-addr", "", defaultNodeRpcAddress(), "set the burrow node rpc server address (default respects $BURROW_CLIENT_NODE_ADDRESS)")
	batchCmd.PersistentFlags().StringVarP(&clientDo.PubkeyFlag, "pubkey", "", defaultPublicKey(), "specify the public key to sign with (defaults to $BURROW_CLIENT_PUBLIC_KEY)")
	batchCmd.PersistentFlags().StringVarP(&clientDo.AddrFlag, "addr", "a", defaultAddress(), "specify the account address (for which the public key can be found at monax-keys) (default respects $BURROW_CLIENT_ADDRESS)")
	batchCmd.PersistentFlags().StringVarP(&clientDo.ChainidFlag, "chain", "c", defaultChainId(), "specify the chainID (default respects $CHAIN_ID)")

	batchCmd.PersistentFlags().StringVarP(&clientDo.AmtFlag, "amt", "", "1", "specify an amount")
	batchCmd.PersistentFlags().StringVarP(&clientDo.FeeFlag, "fee", "f", "1", "specify the fee to send")
	batchCmd.PersistentFlags().StringVarP(&clientDo.GasFlag, "gas", "g", "1000", "specify the gas limit for a CallTx")
	batchCmd.PersistentFlags().StringVarP(&clientDo.NonceFlag, "nonce", "", "", "specify the nonce to use for the transaction (should equal the sender account's nonce + 1)")

	batchCmd.PersistentFlags().StringVarP(&clientDo.ToFlag, "contract", "", defaultContractAddress(), "specify ...")
	batchCmd.PersistentFlags().StringVarP(&clientDo.AbiPathFlag, "abi", "", defaultAbiPath(), "specify ...")
	batchCmd.PersistentFlags().StringVarP(&clientDo.JsonDataPathFlag, "json-data", "", defaultJsonDataPath(), "specify the JSON batch file (default respects $BURROW_CLIENT_JSON_DATA_PATH)")
	batchCmd.PersistentFlags().StringVarP(&clientDo.MethodFlag, "method", "", defaultMethod(), "specify ...")
	batchCmd.PersistentFlags().StringVarP(&clientDo.BatchSizeFlag, "batch", "", defaultBatchSize(), "specify ...")

	return batchCmd
}

//------------------------------------------------------------------------------
// Defaults

func defaultContractAddress() string {
	return setDefaultString("BURROW_CLIENT_CONTRACT_ADDRESS", "")
}

func defaultAbiPath() string {
	return setDefaultString("BURROW_CLIENT_ABI_PATH", "")
}

func defaultJsonDataPath() string {
	return setDefaultString("BURROW_CLIENT_JSON_DATA_PATH", "")
}

func defaultMethod() string {
	return setDefaultString("BURROW_CLIENT_METHOD", "")
}

func defaultBatchSize() string {
	return setDefaultString("BURROW_CLIENT_BATCH_SIZE", "100")
}
