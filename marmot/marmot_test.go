package marmot

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Toggle logging
	debugLogging = false
}

// Steppe blessings young Marmot! The Grand Marmota (may his hawks always be blind) says you have great potential...
// But first you must be able to prove our identity.

// The first test is the one you have to fix, and the second one is included to help understand the process of sending
// a transaction.

// This is the test you need to fix. We'd also like to know what the 'isAuthorised' function is doing in the smart
// contract.
// You may find the online Solidity compiler/IDE at https://remix.ethereum.org useful but it is not strictly necessary
// to solve the problem.
func TestMarmotAuthorisation(t *testing.T) {
	// This is your private identity (containing our private key matter for signing). This should not be changed
	// Its address is B179AB338CA29A4C4F3F87BFE2B8DF6A604CFBA9
	marmotAccount := getMarmotAccount()

	// The transaction below, deployTxBytes, deploys a simple EVM contract that is designed to check that we are who
	// we say we are.
	// The contract was compiled from the following (partially redacted) Solidity source code that compiles down to
	// EVM bytecode that is embedded in the transaction:
	/*
		pragma solidity ^0.4.0;

		contract AccessContract {
			// Returns true when the caller is authorised...
			function isAuthorised() public constant returns (bool) {
				// Some secret code ...
			}
		}
	*/

	// Unfortunately before we signed and encoded the transaction a cosmic ray flipped one bit in the EVM bytecode
	// (but nowhere else - and only one bit) we want to deploy and now when we call isAuthorised it does not return
	// true for our marmotAccount as it should.

	// Here is the hex-encoding of the properly signed and wire encoded transaction (i.e. we signed the faulty EVM code)
	badDeployTxHex := "0201b179ab338ca29a4c4f3f87bfe2b8df6a604cfba900000000" +
		"0000000a00000000000000010144a4502f8923914e2972333a9b0b231b26479fb597ff4183841b8d" +
		"3abd20b8b37dd77890b3e8f8bb20b08379e75886d113bd85c902335468c41a9f2f992aa806014e4e" +
		"4cbf37c17c2c95473c7a23607a72339d01a95f31a1cb544fa2552a7ddb5f0000000000000000c800" +
		"0000000000000a01c4606060405260b48060106000396000f360606040526000357c010000000000" +
		"000000000000000000000000000000000000000000000090048063d4e2dc85146039576035565b60" +
		"02565b34600257604860048050506060565b60405180821515815260200191505060405180910390" +
		"f35b6000600073b179ab338ca29a6c4f3f87bfe2b8df6a604cfba990508073ffffffffffffffffff" +
		"ffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614915060b0" +
		"565b509056"

	// Here are the transaction bytes that should deploy our gat-keeper contract, if you can fix the transaction so
	// the deployed contract works then you will need to re-sign, re-wire-encode, and re-hex-encode the transaction
	// to get the correct bytes to replace the ones above
	deployTxBytes, err := hex.DecodeString(badDeployTxHex)
	require.NoError(t, err)

	// accountIsAuthorised handles setting up an EVM execution environment, against which it runs two transactions:
	//
	// 1. The transaction above that will deploy a contract
	// 2. A transaction that calls the deployed contract's 'isAuthorised' function
	//
	// It will return true if the deployed contract is functioning correctly and classifies marmotAccount (that's you)
	// as 'authorised' as it should do.
	//
	// Please see the comments in accountIsAuthorised to understand how it works in more detail - and the code it
	// depends on.
	assert.True(t, accountIsAuthorised(t, marmotAccount, deployTxBytes),
		"We need our gate-keeper contract to recognise us return accountIsAuthorised!")
}

// This test is included as a demonstration of how we formulate a transaction
func TestFormulateTx(t *testing.T) {
	// We'll use our marmot account as the transactions sender (equivalently 'the input account' and 'the caller')
	marmotAccount := getMarmotAccount()
	// set up an execution environemnt to use later
	exec := newTestExecutionEnvironment(t, defaultAccountFromPrivateAccount(marmotAccount))

	// This is a simple piece of EVM bytecode that stores the 6 bytes of 'marmot' in memory, then returns it
	initialisationBytecode, err := acm.NewBytecode(asm.PUSH6, "marmot", asm.PUSH1, 0, asm.MSTORE,
		asm.PUSH1, 32, asm.PUSH1, 0, asm.RETURN)
	require.NoError(t, err)

	// We can print the bytecode as raw bytes or we can tokenise to get back a similar representation we used to build it
	// Note if you have some bytecode as type []byte you need to convert it to bytecode with asm.NewBytecode(bytecode)
	// in order to have access to the Tokens() method
	tokens, err := initialisationBytecode.Tokens()
	require.NoError(t, err)
	fmt.Printf("Raw bytecode: %s\nTokenised bytecode: %s\n", initialisationBytecode, strings.Join(tokens, " "))

	// This creates a CallTx (the workhorse of Burrow) - the same type as the deploy tx in TestMarmot
	// When we call a non-existent contract (i.e. to == nil) Burrow will create a contract whose code is whatever our
	// initialisation bytecode returns (when run in EVM) - in this case 'marmot'
	tx := txs.NewCallTxWithSequence(marmotAccount.PublicKey(), nil, initialisationBytecode, 200, 200, 10, 1)

	// Now we have to sign the transaction by inserting a signature into the transaction itself and using the ChainID
	// metadata. We sign using the account we will send with
	tx.Input.Signature = acm.ChainSign(marmotAccount, exec.GenesisDoc.ChainID(), tx)

	// The final stage before we can submit the transaction to the execution environment is to wire serialise it.
	// We do this using a GoWireCodec (note/HINT: it can DecodeTx too)
	txBytes, err := txs.NewGoWireCodec().EncodeTx(tx)

	// We can determinsitically predict the address where the new contract will be deployed
	contractAddress := acm.NewContractAddress(marmotAccount.Address(), 1)

	// Now we can deliver the Tx
	returnValue := exec.deliverAndCommitTxReturn(t, txBytes, &contractAddress)
	// Return value should be 'marmot'
	assert.Equal(t, binary.LeftPadWord256([]byte("marmot")), binary.LeftPadWord256(returnValue))

	// We'll check now the code was deployed to the contract
	// Retrieve contract account from state
	contractAccount, err := exec.State.GetAccount(contractAddress)
	require.NoError(t, err)

	// When you have a theme, stick to it...
	// The contract's code should be the same
	assert.Equal(t, binary.LeftPadWord256([]byte("marmot")), binary.LeftPadWord256(contractAccount.Code()))
}
