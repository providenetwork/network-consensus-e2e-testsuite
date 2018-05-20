#!/usr/bin/env bash

PWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ -z "${BASE_PATH}" ]]; then
  BASE_PATH=$PWD
fi

if [[ -z "${CHAIN_SPEC}" ]]; then
  CHAIN_SPEC=spec.json
fi

if [[ -z "${BOOTNODES}" ]]; then
  BOOTNODES=
fi

if [[ -z "${LOGGING}" ]]; then
  LOGGING=warning
fi

if [[ -z "${LOG_PATH}" ]]; then
  LOG_PATH="${BASE_PATH}/parity.log"
fi

if [[ -z "${ALLOW_IPS}" ]]; then
  ALLOW_IPS=all
fi

if [[ -z "${AUTO_UPDATE}" ]]; then
  AUTO_UPDATE=all
fi

if [[ -z "${PRUNING}" ]]; then
  PRUNING=auto
fi

if [[ -z "${RESEAL_ON_TXS}" ]]; then
  RESEAL_ON_TXS=all
fi

if [[ -z "${FAT_DB}" ]]; then
  FAT_DB=auto
fi

if [[ -z "${PRUNING}" ]]; then
  PRUNING=auto
fi

if [[ -z "${TRACING}" ]]; then
  TRACING=auto
fi

if [[ -z "${PORT}" ]]; then
  PORT=30300
fi

if [[ -z "${PORTS_SHIFT}" ]]; then
  PORTS_SHIFT=0
fi

if [[ -z "${JSON_RPC_INTERFACE}" ]]; then
  JSON_RPC_INTERFACE=all
fi

if [[ -z "${JSON_RPC_PORT}" ]]; then
  JSON_RPC_PORT=8050
fi

if [[ -z "${JSON_RPC_HOSTS}" ]]; then
  JSON_RPC_HOSTS=all
fi

if [[ -z "${JSON_RPC_CORS}" ]]; then
  JSON_RPC_CORS=all
fi

if [[ -z "${WS_APIS}" ]]; then
  WS_APIS=web3,eth,pubsub,net,parity,parity_pubsub,traces,rpc,shh,shh_pubsub
fi

if [[ -z "${WS_PORT}" ]]; then
  WS_PORT=8051
fi

if [[ -z "${WS_INTERFACE}" ]]; then
  WS_INTERFACE=all
fi

if [[ -z "${WS_HOSTS}" ]]; then
  WS_HOSTS=all
fi

if [[ -z "${WS_ORIGINS}" ]]; then
  WS_ORIGINS=all
fi

if [[ -z "${UI_INTERFACE}" ]]; then
  UI_INTERFACE=all
fi

if [[ -z "${UI_PORT}" ]]; then
  UI_PORT=8052
fi

if [[ -z "${UI_HOSTS}" ]]; then
  UI_HOSTS=all
fi

if [[ -z "${JSON_RPC_APIS}" ]]; then
  JSON_RPC_APIS=web3,eth,net,personal,parity,parity_set,traces,rpc,parity_accounts
fi

if [[ -z "${ENGINE_SIGNER}" ]]; then
  ENGINE_SIGNER=0x0000000000000000000000000000000000000000
fi

if [[ -z "${ENGINE_SIGNER_KEY_PATH}" ]]; then
  ENGINE_SIGNER_KEY_PATH="${BASE_PATH}/.${ENGINE_SIGNER}.key"
fi

if [ ! -f "${ENGINE_SIGNER_KEY_PATH}" ]; then
  touch "${ENGINE_SIGNER_KEY_PATH}"
fi
chmod 0600 "${ENGINE_SIGNER_KEY_PATH}"

PARITY_BIN=$(which parity)
if [ $? -eq 0 ]
then
  echo "provide.network node starting in ${BASE_PATH}; parity bin: ${PARITY_BIN}"
fi

$PARITY_BIN --chain $CHAIN_SPEC \
            --base-path "${BASE_PATH}" \
            --bootnodes "${BOOTNODES}" \
            --logging $LOGGING \
            --log-file "${LOG_PATH}" \
            --allow-ips $ALLOW_IPS \
            --auto-update $AUTO_UPDATE \
            --force-sealing \
            --reseal-on-txs $RESEAL_ON_TXS \
            --fat-db $FAT_DB \
            --pruning $PRUNING \
            --tracing $TRACING \
            --port $PORT \
            --ports-shift $PORTS_SHIFT \
            --jsonrpc-apis $JSON_RPC_APIS \
            --jsonrpc-interface $JSON_RPC_INTERFACE \
            --jsonrpc-port $JSON_RPC_PORT \
            --jsonrpc-hosts $JSON_RPC_HOSTS \
            --jsonrpc-cors $JSON_RPC_CORS \
            --ws-apis $WS_APIS \
            --ws-port $WS_PORT \
            --ws-interface $WS_INTERFACE \
            --ws-hosts $WS_HOSTS \
            --ws-origins $WS_ORIGINS \
            --ui-interface $UI_INTERFACE \
            --ui-port $UI_PORT \
            --ui-hosts $UI_HOSTS \
            --engine-signer $ENGINE_SIGNER \
            --password "${ENGINE_SIGNER_KEY_PATH}"
