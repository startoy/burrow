package marmot

import (
	"fmt"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/db"
)

// Flip this to remove logging (makes test cases easier to see)
var debugLogging = true

// This function first executes the deployContractTxBytes using the privateAccount to send and sign the transaction,
// and then tries to call the isAuthorised function on the deployed contract. Relies upon being passed transaction bytes
// that do as suggested (deploy a contract with such a function).
func accountIsAuthorised(t *testing.T, privateAccount acm.PrivateAccount, deployContractTxBytes []byte) bool {
	//--------------------------------------------------
	// General setup

	// Our execution interface abciApp allows us to execute transactions, emitter allows us to listen for results,
	// and genesisDoc defines the initial participants
	exec := newTestExecutionEnvironment(t, defaultAccountFromPrivateAccount(privateAccount))

	//--------------------------------------------------
	// Execute transaction 1 - deploy the contract

	// We deploy the contract using the transaction we have been passed as an argument
	exec.deliverAndCommitTx(t, deployContractTxBytes)

	//--------------------------------------------------
	// Execute transaction 2 - call the isAuthorised function

	// Deterministically derived address where we know the contract will be deployed
	contractAddress := acm.NewContractAddress(privateAccount.Address(), 1)
	// We formulate the appropriate transaction to call the isAuthorised function
	callIsAuthorisedTxBytes := callIsAuthorisedFunctionTx(t, privateAccount, exec.GenesisDoc)
	// And execute a block containing the transaction
	returnValue := exec.deliverAndCommitTxReturn(t, callIsAuthorisedTxBytes, &contractAddress)

	//--------------------------------------------------
	// Check the return value
	// Return value should be 1 indicating true if we are authorised!
	return binary.LeftPadWord256([]byte{1}) == binary.LeftPadWord256(returnValue)
}

func callIsAuthorisedFunctionTx(t *testing.T, privAcc acm.PrivateAccount, genesisDoc *genesis.GenesisDoc) []byte {
	// This is the identifier for the isAuthorised function, defined by the EVM ABI to be the first 4 bytes of the SHA3
	// keccak-256 hash of the function signature (excluded return types)
	functionIdentifierCode := sha3.Sha3([]byte("isAuthorised()"))[:4]
	acc := defaultAccountFromPrivateAccount(privAcc)
	// We are able to determine the address of the contract by assuming that privAcc has only deployed one contract
	// at sequence number 1, and so its address is known deterministically
	contractAddress := acm.NewContractAddress(acc.Address(), 1)
	// Formulate a CallTx transaction which will set up the EVM to call a contract, the function identifier will be
	// provided as an input to the EVM stack machine when it runs the contract code found at contractAddress
	// We are able to predict the sequence number (of the calling account) assuming all it has done so far is to create
	// a contract (incrementing its sequence number twice)
	checkAuthorisedTx := txs.NewCallTxWithSequence(acc.PublicKey(), &contractAddress, functionIdentifierCode, 10, 200, 10, 3)
	// We sign the transaction by setting its signature (note the bytes we sign obviously do not include the signature
	// to avoid a recursive problem - yes this is a bit of a weird way of doing things)
	checkAuthorisedTx.Input.Signature = acm.ChainSign(privAcc, genesisDoc.ChainID(), checkAuthorisedTx)
	// Before we broadcast to the consensus engine we serialise transactions, and the consensus engine orders them and
	// provides them unaltered to the abciApp, therefore we must provide bytes to DeliverTx, which we encode here using
	// the 'go wire' serialisation that Burrow uses
	callIsAuthorisedTxBytes, err := txs.NewGoWireCodec().EncodeTx(checkAuthorisedTx)
	require.NoError(t, err)
	return callIsAuthorisedTxBytes

}

type testExecutionEnvironment struct {
	height uint64
	*execution.State
	abci_types.Application
	event.Emitter
	*genesis.GenesisDoc
}

// This helper function sets up a BatchCommitter which is the object to which a Burrow blockchain feeds transactions.
// Itcreates a state with a single participant account (that can send transactions). You shouldn't need to change this.
func newTestExecutionEnvironment(t *testing.T, accounts ...acm.Account) *testExecutionEnvironment {
	// We are mostly here just wiring the the dependencies for our abciApp
	// Genesis building
	validator := acm.AsValidator(acm.NewConcreteAccountFromSecret("validator").Account())
	genesisAccounts := make(map[string]acm.Account, len(accounts))
	for i, acc := range accounts {
		genesisAccounts[fmt.Sprintf("genesis_account_%v", i)] = acc
	}
	genesisTime, err := time.Parse("Jan _2 2006 15:04", "Aug 6 2006 16:35")
	require.NoError(t, err)
	genesisDoc := genesis.MakeGenesisDocFromAccounts("MarmotChain", nil, genesisTime,
		genesisAccounts, map[string]acm.Validator{"validator": validator})
	// From the GenesisDoc we can make the genesis or initial state
	state := execution.MakeGenesisState(db.NewMemDB(), genesisDoc)

	// uncomment the below and comment the above to see log output on stderr (can be a bit noisy)

	logger := loggers.NewNoopInfoTraceLogger()
	if debugLogging {
		logger, _ = lifecycle.NewStdErrLogger()
	}

	// A pub-sub event hub
	emitter := event.NewEmitter(logger)
	// block state
	chain := blockchain.NewBlockchain(genesisDoc)
	// Unique code determined from GenesisDoc
	chainID := genesisDoc.ChainID()

	// Mempool checker - used by CheckTx, which we do not use in these tests
	checker := execution.NewBatchChecker(state, chainID, chain, logger)
	// Block committer - used by DeliverTx to update the EVM state
	committer := execution.NewBatchCommitter(state, chainID, chain, emitter, logger)
	// The abciApp is the interface that our Tendermint consensus mechanism talks to
	return &testExecutionEnvironment{
		State:       state,
		Application: abci.NewApp(chain, checker, committer, logger),
		Emitter:     emitter,
		GenesisDoc:  genesisDoc,
	}
}

func (exec *testExecutionEnvironment) deliverAndCommitTx(t *testing.T, txBytes []byte) {
	exec.deliverAndCommitTxReturn(t, txBytes, nil)
}

// This helper runs the sequence that occurs when we commit a block (ish) with just one transaction
func (exec *testExecutionEnvironment) deliverAndCommitTxReturn(t *testing.T, txBytes []byte,
	contractAddress *acm.Address) []byte {

	ch := make(chan []byte, 1)
	if contractAddress != nil {
		exec.Emitter.Subscribe("test_sub", events.EventStringAccCall(*contractAddress),
			func(data event.AnyEventData) {
				ch <- data.EventDataCall().Return
			})
	} else {
		go func() {
			ch <- nil
		}()
	}
	exec.height++
	exec.Application.BeginBlock(abci_types.RequestBeginBlock{
		Hash: []byte("foo"),
		Header: &abci_types.Header{
			NumTxs: 1,
			Height: exec.height,
		},
	})
	result := exec.Application.DeliverTx(txBytes)
	if !assert.Equal(t, abci_types.CodeType_OK, result.Code) {
		fmt.Println(result.Error())
	}
	exec.Application.EndBlock(exec.height)
	result = exec.Application.Commit()
	if !assert.Equal(t, abci_types.CodeType_OK, result.Code) {
		fmt.Println(result.Error())
	}
	//--------------------------------------------------
	// Now we wait for the return from isAuthorised. May time out if something else goes wrong.
	select {
	case <-time.After(time.Second * 2):
		t.Fatalf("timed out waiting for return value")
		return nil
	case returnValue := <-ch:
		return returnValue
	}
}

// This gets you the marmot account you should use in the test (you shouldn't change this function)
func getMarmotAccount() acm.PrivateAccount {
	// Generates a pseudorandom private key based on the seed secret
	return acm.GeneratePrivateAccountFromSecret("marmot")
}

// This just creates a full account with the default permissions using the public address and key of the
// private account and with a balance so it can act. Account is unit of state on an Ethereum blockchain
func defaultAccountFromPrivateAccount(privAcc acm.PrivateAccount) acm.Account {
	return acm.ConcreteAccount{
		Address:     privAcc.Address(),
		PublicKey:   privAcc.PublicKey(),
		Balance:     99999,
		Permissions: permission.DefaultAccountPermissions,
	}.Account()
}
