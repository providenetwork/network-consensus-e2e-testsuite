package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/provideservices/provide-go"
	"github.com/stretchr/testify/assert"
)

// This suite provides end-to-end test coverage for the Aura network consensus contracts
// and the integration of auth_os as a means to achieve a very high level of upgradeability
// from the genesis block of newly-deployed networks.

const defaultStepsUnderTest int = 1 // default number of sealed blocks to test against (certain cases wait for defaultStepsUnderTest * stepDuration)

func TestMain(m *testing.M) {
	var retval int
	main()
	retval = m.Run()
	teardown()
	os.Exit(retval)
}

func TestDeployGenesisBlock(t *testing.T) {
	if accounts, ok := chainSpec["accounts"].(map[string]interface{}); ok {
		masterOfCeremonyAccount, accountOk := accounts[masterOfCeremonyGenesisAddress]
		assert.True(t, accountOk)
		assert.NotNil(t, masterOfCeremonyAccount)

		for genesisContractAccountName, genesisContractAccountAddr := range genesisContractAccounts {
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

func testWhileSealing(nBlocks int, test func(int, int)) {
	offset := int(*provide.GetBlockNumber(networkID, rpcURL)) // block from which test case started; used to offset the current block for assertion purposes
	i := 0
	for i < nBlocks {
		test(i, offset+i)
		i++
		time.Sleep(time.Second * 5) // FIXME-- use stepDuration dynamically as configured in chainspec
	}
}

func getValidators(t *testing.T) ([]common.Address, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getValidators", &_abi, _params)
	assert.Nil(t, err)

	validators, validatorsOk := (*resp).([]common.Address)
	assert.True(t, validatorsOk)

	return validators, nil
}

func getMasterOfCeremonyPrivateKey() *string {
	privateKey, _ := parseCachedMasterOfCeremonyPrivateKey()
	return stringOrNil(hex.EncodeToString(ethcrypto.FromECDSA(privateKey)))
}

func getValidatorCount(t *testing.T) (uint64, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getValidatorCount", &_abi, _params)
	assert.Nil(t, err)

	validatorCount, countOk := (*resp).(*big.Int)
	assert.True(t, countOk)

	return validatorCount.Uint64(), nil
}

func getValidatorSupportCount(t *testing.T, addr string) (uint64, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	_params = append(_params, addr)
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getValidatorSupportCount", &_abi, _params)
	assert.Nil(t, err)

	supportCount, supportCountOk := (*resp).(*big.Int)
	assert.True(t, supportCountOk)

	return supportCount.Uint64(), nil
}

func getValidatorSupportDivisor(t *testing.T) (uint64, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getValidatorSupportDivisor", &_abi, _params)
	assert.Nil(t, err)

	divisor, divisorOk := (*resp).(*big.Int)
	assert.True(t, divisorOk)

	return divisor.Uint64(), nil
}

func getPendingValidators(t *testing.T) ([]common.Address, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getPendingValidators", &_abi, _params)
	assert.Nil(t, err)

	validators, validatorsOk := (*resp).([]common.Address)
	assert.True(t, validatorsOk)

	return validators, nil
}

func getPendingValidatorCount(t *testing.T) (uint64, error) {
	_abi, err := parseNetworkConsensusABI()
	assert.Nil(t, err)
	assert.NotNil(t, _abi)

	var _params []interface{}
	resp, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "getPendingValidatorCount", &_abi, _params)
	assert.Nil(t, err)

	validatorCount, countOk := (*resp).(*big.Int)
	assert.True(t, countOk)

	return validatorCount.Uint64(), nil
}

func TestMasterOfCeremonySigner(t *testing.T) {
	// testWhileSealing(defaultStepsUnderTest, func(i, block int) {
	// 	blockNumber := provide.GetBlockNumber(networkID, rpcURL)
	// 	assert.NotNil(t, blockNumber)
	// 	assert.Equal(
	// 		t,
	// 		uint64(block),
	// 		*blockNumber,
	// 		fmt.Sprintf("it should return %v", block),
	// 	)
	// })

	_abi, err := parseNetworkConsensusABI()
	assert.NotNil(t, _abi)

	masterOfCeremonySupportCount, err := getValidatorSupportCount(t, masterOfCeremonyGenesisAddress)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), masterOfCeremonySupportCount)

	// ADD A VALIDATOR...
	validator, _, err := provide.GenerateKeyPair()
	assert.Nil(t, err)
	var params []interface{}
	params = append(params, *validator)
	_, err = provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "addValidator", &_abi, params)
	assert.Nil(t, err)

	testWhileSealing(2, func(i, block int) {
		if i == 1 {
			// validatorSupportCount, err := getValidatorSupportCount(t, *validator)
			// assert.Nil(t, err)
			// assert.Equal(t, uint64(1), validatorSupportCount)

			validators, _ := getValidators(t)
			validatorCount, _ := getValidatorCount(t)
			assert.Equal(t, 1, len(validators))
			assert.Equal(t, uint64(len(validators)), validatorCount)

			pendingValidators, _ := getPendingValidators(t)
			pendingValidatorCount, _ := getPendingValidatorCount(t)
			assert.Equal(t, 2, len(pendingValidators))
			assert.Equal(t, uint64(len(pendingValidators)), pendingValidatorCount)
		}
	})
}

func TestGetValidatorSupportDivisor(t *testing.T) {
	validatorSupportDivisor, err := getValidatorSupportDivisor(t)
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), validatorSupportDivisor)
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
	testWhileSealing(2, func(i, block int) {
		blockNumber := provide.GetBlockNumber(networkID, rpcURL)
		assert.NotNil(t, blockNumber)
		assert.Equal(
			t,
			uint64(block),
			*blockNumber,
			fmt.Sprintf("it should return %v", block),
		)
	})
}

func TestGetLatestBlock(t *testing.T) {
	testWhileSealing(2, func(i, block int) {
		latestBlock, err := provide.GetLatestBlock(networkID, rpcURL)
		assert.Nil(t, err)
		actual, err := hexutil.DecodeUint64(latestBlock.Result.(map[string]interface{})["number"].(string))
		assert.Nil(t, err)
		assert.Equal(
			t,
			uint64(block),
			actual,
			fmt.Sprintf("it should return %v", block),
		)
	})
}

func TestGetNetworkStatus(t *testing.T) {
	status, err := provide.GetNetworkStatus(networkID, rpcURL)
	assert.Nil(t, err)
	assert.NotNil(t, status)

	assert.NotNil(t, status.ChainID)
	chainID, _ := hexutil.DecodeUint64(*status.ChainID)
	actual, err := hexutil.DecodeUint64(chainSpec["params"].(map[string]interface{})["networkID"].(string))
	assert.Nil(t, err)
	assert.Equal(
		t,
		chainID,
		actual,
		"it should include the networkID as configured in the chainspec",
	)

	assert.NotNil(t, status.Block)
	assert.Equal(
		t,
		*provide.GetBlockNumber(networkID, rpcURL),
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

func TestAddValidator(t *testing.T) {
	testWhileSealing(1, func(i, block int) {
		networkConsensusABI, _ := parseNetworkConsensusABI()
		var _params []interface{}
		_params = append(_params, "0x87b7af6915fa56a837fa85e31ad6a450c41e8fab")
		_, err := provide.ExecuteContract(networkID, rpcURL, masterOfCeremonyGenesisAddress, stringOrNil(networkConsensusGenesisAddress), getMasterOfCeremonyPrivateKey(), nil, nil, "addValidator", networkConsensusABI, _params)
		assert.Nil(t, err)

		// validators, _ := getValidators(t)
		// validatorCount, _ := getValidatorCount(t)
		// assert.Equal(t, 1, len(validators))
		// assert.Equal(t, uint64(len(validators)), validatorCount)

		// pendingValidators, _ := getPendingValidators(t)
		// pendingValidatorCount, _ := getPendingValidatorCount(t)
		// assert.Equal(t, 2, len(pendingValidators))
		// assert.Equal(t, uint64(len(pendingValidators)), pendingValidatorCount)
		// fmt.Printf("resp: %s", response)
		// assert.Equal(
		// 	t,
		// 	uint64(block),
		// 	latestBlock,
		// 	fmt.Sprintf("it should add the validator %v", block),
		// )
	})
}

// func TestTraceTx(t *testing.T) {
// 	log.Printf("TestTraceTx")

// }

// func TestGetTxReceipt(t *testing.T) {
// 	log.Printf("TestGetTxReceipt")

// }
