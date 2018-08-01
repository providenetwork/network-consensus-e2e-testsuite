package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"

	provide "github.com/provideservices/provide-go"
)

const abstractStorageGenesisAddress string = "0x0000000000000000000000000000000000000009"
const registryIdxGenesisAddress string = "0x0000000000000000000000000000000000000010"
const providerGenesisAddress string = "0x0000000000000000000000000000000000000011"

const auraGenesisAddress string = "0x0000000000000000000000000000000000000012"
const validatorConsoleGenesisAddress string = "0x0000000000000000000000000000000000000013"
const votingConsoleGenesisAddress string = "0x0000000000000000000000000000000000000014"
const initBridgesGenesisAddress string = "0x0000000000000000000000000000000000000015"
const bridgesConsoleGenesisAddress string = "0x0000000000000000000000000000000000000016"
const oraclesConsoleGenesisAddress string = "0x0000000000000000000000000000000000000017"
const networkConsensusGenesisAddress string = "0x0000000000000000000000000000000000000018"

var networkID = "arbitraryidentifier"
var rpcURL string
var chainSpec map[string]interface{}
var masterOfCeremonyGenesisAddress string
var genesisContractAccounts = map[string]string{
	"AbstractStorage":  abstractStorageGenesisAddress,
	"RegistryIdx":      registryIdxGenesisAddress,
	"Provider":         providerGenesisAddress,
	"Aura":             auraGenesisAddress,
	"ValidatorConsole": validatorConsoleGenesisAddress,
	"VotingConsole":    votingConsoleGenesisAddress,
	"InitBridges":      initBridgesGenesisAddress,
	"BridgesConsole":   bridgesConsoleGenesisAddress,
	"OraclesConsole":   oraclesConsoleGenesisAddress,
	"NetworkConsensus": networkConsensusGenesisAddress,
}

func main() {
	var err error
	var privateKey *ecdsa.PrivateKey

	if !hasCachedMasterOfCeremony() {
		var addr *string
		addr, privateKey, err = provide.GenerateKeyPair()
		if err != nil {
			teardown()
			log.Fatalf("Failed to generate master of ceremony address; %s", err.Error())
		}
		masterOfCeremonyGenesisAddress = common.HexToAddress(*addr).Hex()

		setupTemporaryDirectories()
		_, err = cacheMasterOfCeremonyKeys(masterOfCeremonyGenesisAddress, privateKey)
	} else {
		keystore, err := parseCachedMasterOfCeremonyKeystore()
		if err == nil {
			masterOfCeremonyGenesisAddress = common.HexToAddress(fmt.Sprintf("0x%s", keystore["address"])).Hex()
			privateKey, err = parseCachedMasterOfCeremonyPrivateKey()
		}
	}

	if err == nil {
		rpcURL, chainSpec, err = bootstrap(networkID, masterOfCeremonyGenesisAddress, privateKey, genesisContractAccounts)
	}

	if err != nil {
		teardown()
		log.Fatalf("%s", err.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			teardown()
		}
	}()
}
