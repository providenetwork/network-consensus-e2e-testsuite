package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/provideservices/provide-go"
)

const cachedChainspecPath = "./.spec.json"
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
func deployNetwork(networkID string, spec []byte) (string, error) {
	var rpcURL string
	var err error

	err = os.Mkdir(tmpWorkdirPath, 0755)
	if err != nil {
		return "", err
	}

	ioutil.WriteFile(cachedChainspecPath, spec, 0644)
	err = ioutil.WriteFile(tmpChainspecPath, spec, 0644)
	if err == nil {
		rpcURL = fmt.Sprintf("http://localhost:%v", rpcPort)
		go func() {
			runNetworkNode()
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

// attempt to read a cached chainspec from .spec.json; if
// one exists, it will be reused to elimninate the need to
// compile auth_os and network consensus contracts from source.
// running `make clean` or simply removing spec.json will result
// in spec.json being rebuilt
func readCachedChainspec() ([]byte, error) {
	var spec []byte
	var err error
	if _, err = os.Stat(cachedChainspecPath); err == nil {
		spec, err = ioutil.ReadFile(cachedChainspecPath)
	}
	return spec, err
}

// start a parity peer on the given network under test;
func runNetworkNode() {
	cmd := exec.Command("bash", "-c", "./start-node.sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), "LOGGING=debug")
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
func bootstrap(networkID, masterOfCeremony string, genesisContractAccounts map[string]string) (string, map[string]interface{}, error) {
	var rpcURL string
	var spec []byte
	var parsedSpec map[string]interface{}
	var err error
	terminateOrphans()
	spec, err = readCachedChainspec()
	if err != nil {
		osRef := os.Getenv("OS_REF")
		consensusRef := os.Getenv("NETWORK_CONSENSUS_REF")
		log.Printf("Compiling auth_os and network consensus contracts from source (using revisions %s and %s, respectively) for inclusion in custom spec.json", osRef, consensusRef)
		spec, err = provide.BuildChainspec(osRef, consensusRef, masterOfCeremony, genesisContractAccounts)
	} else {
		log.Printf("Using cached spec.json from previous run")
	}
	fmt.Printf("%s", spec)
	if err == nil {
		rpcURL, err = deployNetwork(networkID, spec)
	}
	if err == nil {
		err = json.Unmarshal(spec, &parsedSpec)
	}
	return rpcURL, parsedSpec, err
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
