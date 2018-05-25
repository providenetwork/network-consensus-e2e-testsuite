.PHONY: clean test test_unicorn

clean:
	rm -rf .tmp*

test: clean
	BASE_PATH=.tmp CHAIN_SPEC=.tmp/spec.json go test

test_unicorn: clean
	BASE_PATH=.tmp CHAIN_SPEC=.tmp/spec.json CHAINSPEC_REF=unicorn go test
