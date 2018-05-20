.PHONY: clean test test_unicorn

clean:
	rm -rf .tmp/

test: clean
	BASE_PATH=.tmp CHAIN_SPEC=.tmp/spec.json go test

test_unicorn: clean
	BASE_PATH=.tmp CHAIN_SPEC=.tmp/spec.json OS_REF=storage-unmarshall NETWORK_CONSENSUS_REF=registryexec-subclass-reverted CHAINSPEC_REF=unicorn go test
