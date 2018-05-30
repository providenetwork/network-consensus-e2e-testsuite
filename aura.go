package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/provideservices/provide-go"
)

const cachedChainspecPath = "./.spec"
const parentPidPath = "./.pid"
const tmpChainspecPath = "./.tmp/spec.json"
const tmpWorkdirPath = "./.tmp"

var (
	nodes   []*exec.Cmd
	rpcPort = 8050
)

// entrypoint to deploy the blockchain network which will be placed under test;
// uses the given chain spec as the genesis block, starts the initial JSON-RPC client
// and returns the JSON-RPC url where it is listening
func deployNetwork(networkID, masterOfCeremony string, masterOfCeremonyPrivateKey *ecdsa.PrivateKey, spec []byte) (string, error) {
	var rpcURL string
	var err error

	err = ioutil.WriteFile(tmpChainspecPath, spec, 0644)
	if err == nil {
		rpcURL = fmt.Sprintf("http://localhost:%v", rpcPort)
		go func() {
			runNetworkNode(masterOfCeremony, masterOfCeremonyPrivateKey)
		}()

		var pid int
		for pid == 0 {
			if len(nodes) > 0 {
				proc := nodes[len(nodes)-1].Process
				if proc != nil {
					pid = proc.Pid
				}
			}
		}
		if len(nodes) == 1 {
			ioutil.WriteFile(parentPidPath, []byte(fmt.Sprintf("%v", pid)), 0644)
		}
		log.Printf("Running provide.network node with JSON-RPC listening on port %v; parent pid: %v", rpcPort, pid)
		rpcPort++

		chainID := provide.GetChainID(networkID, rpcURL)
		for chainID == nil {
			chainID = provide.GetChainID(networkID, rpcURL)
		}
	}

	return rpcURL, err
}

func cachedChainspecJsonPath() string {
	osRef := os.Getenv("OS_REF")
	consensusRef := os.Getenv("NETWORK_CONSENSUS_REF")
	filename := fmt.Sprintf("%s+%s", osRef, consensusRef)
	return fmt.Sprintf("%s/%s.json", cachedChainspecPath, filename)
}

// attempt to read a cached chainspec from .spec.json; if
// one exists, it will be reused to elimninate the need to
// compile auth_os and network consensus contracts from source.
// running `make clean` or simply removing spec.json will result
// in spec.json being rebuilt
func readCachedChainspec() ([]byte, error) {
	var spec []byte
	var err error
	chainspecPath := cachedChainspecJsonPath()
	if _, err = os.Stat(chainspecPath); err == nil {
		spec, err = ioutil.ReadFile(chainspecPath)
	}
	return spec, err
}

// start a parity peer on the given network under test;
func runNetworkNode(masterOfCeremony string, masterOfCeremonyPrivateKey *ecdsa.PrivateKey) {
	cmd := exec.Command("bash", "-c", "./start-node.sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), "LOGGING=debug")
	if masterOfCeremony != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ENGINE_SIGNER=%s", masterOfCeremony))
		if masterOfCeremonyPrivateKey != nil {
			keyPairJSON, err := provide.MarshalEncryptedKey(common.HexToAddress(masterOfCeremony), masterOfCeremonyPrivateKey, hex.EncodeToString(ethcrypto.FromECDSA(masterOfCeremonyPrivateKey)))
			if err == nil {
				fmt.Printf("%s", keyPairJSON)
				cmd.Env = append(cmd.Env, fmt.Sprintf("ENGINE_SIGNER_KEY_JSON=%s", keyPairJSON))
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("ENGINE_SIGNER_PRIVATE_KEY=%s", hex.EncodeToString(ethcrypto.FromECDSA(masterOfCeremonyPrivateKey))))
		}
	}
	nodes = append(nodes, cmd)
	go func() {
		log.Printf("Attempting to run provide.network node with JSON-RPC listening on port %v", rpcPort)
		err := cmd.Run()
		if err != nil {
			log.Printf("Failed to run provide.network node; %s", err.Error())
		} else {
			log.Printf("provide.network node with pid %v exited cleanly", cmd.Process.Pid)
			os.RemoveAll(parentPidPath)
		}
	}()
}

// bootstrap initializes the e2e Aura test harness; returns JSON-RPC url where
// genesis client is listening and a parsed representation of the resolved
// chainspec JSON, which can be used for asserting valid deployment
func bootstrap(networkID, masterOfCeremony string, masterOfCeremonyPrivateKey *ecdsa.PrivateKey, genesisContractAccounts map[string]string) (string, map[string]interface{}, error) {
	var rpcURL string
	var spec []byte
	var parsedSpec map[string]interface{}
	var err error
	terminateOrphans()
	setupTemporaryDirectories()
	spec, err = readCachedChainspec()
	if err != nil {
		osRef := os.Getenv("OS_REF")
		consensusRef := os.Getenv("NETWORK_CONSENSUS_REF")
		log.Printf("Compiling auth_os and network consensus contracts from source (using revisions %s and %s, respectively) for inclusion in custom spec.json", osRef, consensusRef)
		spec, err = provide.BuildChainspec(osRef, consensusRef, masterOfCeremony, genesisContractAccounts)
		err = ioutil.WriteFile(cachedChainspecJsonPath(), spec, 0644)
	} else {
		log.Printf("Using cached spec.json from previous run")
	}
	if err == nil {
		rpcURL, err = deployNetwork(networkID, masterOfCeremony, masterOfCeremonyPrivateKey, spec)
	}
	if err == nil {
		err = json.Unmarshal(spec, &parsedSpec)
	}
	return rpcURL, parsedSpec, err
}

func setupTemporaryDirectories() error {
	err := os.Mkdir(tmpWorkdirPath, 0755)
	if err != nil {
		return err
	}
	err = os.Mkdir(cachedChainspecPath, 0755)
	if err != nil {
		return err
	}
	return nil
}

func teardown() {
	for _, cmd := range nodes {
		log.Printf("Attempting to kill running provide.network node; pid: %v", cmd.Process.Pid)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

func terminateOrphans() {
	if _, err := os.Stat(parentPidPath); err == nil {
		if pidBytes, err := ioutil.ReadFile(parentPidPath); err == nil {
			pid, err := strconv.Atoi(strings.Trim(string(pidBytes), "\n"))
			if err == nil && pid > 0 {
				log.Printf("Attempting to kill orphaned provide.network node; pid: %v", pid)
				syscall.Kill(-pid, syscall.SIGKILL)
			}
		}
		os.RemoveAll(parentPidPath)
	}
}
