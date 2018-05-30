package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/provideservices/provide-go"
	"github.com/stretchr/testify/assert"
)

// This suite provides end-to-end test coverage for the Aura network consensus contracts
// and the integration of auth_os as a means to achieve a very high level of upgradeability
// from the genesis block of newly-deployed networks.

var masterOfCeremonyGenesisAddress string = "0x9077F27fDD606c41822f711231eEDA88317aa67a"

const registryStorageGenesisAddress string = "0x0000000000000000000000000000000000000009"
const initRegistryGenesisAddress string = "0x0000000000000000000000000000000000000010"
const appConsoleGenesisAddress string = "0x0000000000000000000000000000000000000011"
const versionConsoleGenesisAddress string = "0x0000000000000000000000000000000000000012"
const implementationConsoleGenesisAddress string = "0x0000000000000000000000000000000000000013"
const auraGenesisAddress string = "0x0000000000000000000000000000000000000014"
const validatorConsoleGenesisAddress string = "0x0000000000000000000000000000000000000015"
const votingConsoleGenesisAddress string = "0x0000000000000000000000000000000000000016"
const networkConsensusGenesisAddress string = "0x0000000000000000000000000000000000000017"

var expectedGenesisContractAccounts = map[string]string{
	"RegistryStorage":       registryStorageGenesisAddress,
	"InitRegistry":          initRegistryGenesisAddress,
	"AppConsole":            appConsoleGenesisAddress,
	"VersionConsole":        versionConsoleGenesisAddress,
	"ImplementationConsole": implementationConsoleGenesisAddress,
	"Aura":                  auraGenesisAddress,
	"ValidatorConsole":      validatorConsoleGenesisAddress,
	"VotingConsole":         votingConsoleGenesisAddress,
	"NetworkConsensus":      networkConsensusGenesisAddress,
}

var chainSpec map[string]interface{}
var networkID = "arbitraryidentifier"
var rpcURL string

func TestMain(m *testing.M) {
	var retval int
	var err error

	addr, privateKey, err := provide.GenerateKeyPair()
	if err != nil {
		teardown()
		log.Fatalf("Failed to generate master of ceremony address; %s", err.Error())
	}

	rpcURL, chainSpec, err = bootstrap(networkID, *addr, privateKey, expectedGenesisContractAccounts)
	if err != nil {
		teardown()
		log.Fatalf("%s", err.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			teardown()
		}
	}()
	retval = m.Run()
	teardown()
	os.Exit(retval)
}

func TestDeployGenesisBlock(t *testing.T) {
	if accounts, ok := chainSpec["accounts"].(map[string]interface{}); ok {
		masterOfCeremonyAccount, accountOk := accounts[masterOfCeremonyGenesisAddress]
		assert.True(t, accountOk)
		assert.NotNil(t, masterOfCeremonyAccount)

		for genesisContractAccountName, genesisContractAccountAddr := range expectedGenesisContractAccounts {
			genesisContractAccount, accountOk := accounts[genesisContractAccountAddr].(map[string]interface{})
			assert.True(t, accountOk, "it should contain a %s contract account in the chainspec at %s", genesisContractAccountName, genesisContractAccountAddr) // expected address appeared in genesis block
			if accountOk {
				assert.NotNil(t, genesisContractAccount)

				genesisContractBytecode, constructorOk := genesisContractAccount["constructor"].(string)
				assert.True(t, constructorOk, "it should contain a %s contract constructor in the chainspec at %s", genesisContractAccountName, genesisContractAccountAddr)
				assert.NotNil(t, genesisContractBytecode)

				if constructorOk {
					expectedContractBytecode := genesisContractBytecode[142:len(genesisContractBytecode)] // bytecode offset for comparison
					deployedContractBytecode, err := provide.GetCode(networkID, rpcURL, genesisContractAccountAddr, "latest")

					assert.Nil(t, err)
					assert.NotNil(t, deployedContractBytecode)

					if genesisContractAccountAddr != networkConsensusGenesisAddress {
						assert.True(
							t,
							strings.Contains(*deployedContractBytecode, expectedContractBytecode),
							fmt.Sprintf("it should contain deployed contract bytecode for the %s contract as configured in the chainspec at %s", genesisContractAccountName, genesisContractAccountAddr),
						)
					} else {
						assert.True(
							t,
							*deployedContractBytecode != "0x",
							fmt.Sprintf("it should contain deployed contract bytecode for the %s contract as configured in the chainspec at %s", genesisContractAccountName, genesisContractAccountAddr),
						)
					}
				}
			}
		}
	}
}

func TestMasterOfCeremonySigner(t *testing.T) {

	blockNumber := provide.GetBlockNumber(networkID, rpcURL)
	assert.NotNil(t, blockNumber)
	assert.Equal(
		t,
		uint64(0),
		*blockNumber,
		"it should return 0",
	)
}

func TestGetChainID(t *testing.T) {
	chainID := provide.GetChainID(networkID, rpcURL)
	assert.NotNil(t, chainID)
	assert.Equal(
		t,
		chainSpec["params"].(map[string]interface{})["networkID"],
		hexutil.EncodeBig(chainID),
		"it should match the networkID configured in the chainspec",
	)
}

func TestGetBlockNumber(t *testing.T) {
	blockNumber := provide.GetBlockNumber(networkID, rpcURL)
	assert.NotNil(t, blockNumber)
	assert.Equal(
		t,
		uint64(0),
		*blockNumber,
		"it should return 0",
	)
}

func TestGetLatestBlock(t *testing.T) {
	latestBlock, err := provide.GetLatestBlock(networkID, rpcURL)
	assert.Nil(t, err)
	assert.Equal(
		t,
		uint64(0),
		latestBlock,
		"it should return 0",
	)
}

func TestGetNetworkStatus(t *testing.T) {
	status, err := provide.GetNetworkStatus(networkID, rpcURL)
	assert.Nil(t, err)
	assert.NotNil(t, status)

	assert.NotNil(t, status.ChainID)
	assert.Equal(
		t,
		chainSpec["params"].(map[string]interface{})["networkID"],
		hexutil.EncodeBig(status.ChainID),
		"it should include the networkID as configured in the chainspec",
	)

	assert.NotNil(t, status.Block)
	assert.Equal(
		t,
		uint64(0),
		status.Block,
		"it should include the latest block number",
	)

	assert.NotNil(t, status.PeerCount)
	assert.Equal(
		t,
		uint64(0),
		status.PeerCount,
		"it should include the peer count",
	)

	assert.NotNil(t, status.ProtocolVersion)
	assert.Equal(
		t,
		"63", // PV63 (Fast synchronization)
		*status.ProtocolVersion,
		"it should include the protocol version",
	)

	assert.NotNil(t, status.State)
	assert.Equal(
		t,
		"synced",
		*status.State,
		"it should include a human-readable syncing state",
	)

	assert.False(t, status.Syncing, "it should include a flag indicating the chain is up-to-date")
}

func TestGetPeerCount(t *testing.T) {
	peerCount := provide.GetPeerCount(networkID, rpcURL)
	assert.NotNil(t, peerCount)

}

func TestGetProtocolVersion(t *testing.T) {
	protocolVersion := provide.GetProtocolVersion(networkID, rpcURL)
	assert.NotNil(t, protocolVersion)
}

// func TestTraceTx(t *testing.T) {
// 	log.Printf("TestTraceTx")

// }

// func TestGetTxReceipt(t *testing.T) {
// 	log.Printf("TestGetTxReceipt")

// }
