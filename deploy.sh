#!/bin/bash

set -e

BASTION_KEY_PATH="bastion.key"
K8S_KEY_PATH="k8s.key"

function clean() {
    rm -f ${BASTION_KEY_PATH} ${K8S_KEY_PATH} 2>/dev/null
}
function check_vars() {
    [[ -z ${BASTION_HOST} ]] && { echo "BASTION_HOST is not set. Cancelling deploy"; exit 1; }
    [[ -z ${K8S_HOST} ]] && { echo "K8S_HOST is not set. Cancelling deploy"; exit 1; }
    [[ -z ${BASTION_PORT} ]] && { echo "BASTION_PORT is not set. Cancelling deploy"; exit 1; }
    [[ -z ${K8S_PORT} ]] && { echo "K8S_PORT is not set. Cancelling deploy"; exit 1; }
    [[ -z ${BASTION_USER} ]] && { echo "BASTION_USER is not set. Cancelling deploy"; exit 1; }
    [[ -z ${K8S_USER} ]] && { echo "K8S_USER is not set. Cancelling deploy"; exit 1; }
    [[ -z ${BASTION_KEY} ]] && { echo "BASTION_KEY is not set. Cancelling deploy"; exit 1; }
    [[ -z ${K8S_KEY} ]] && { echo "K8S_KEY is not set. Cancelling deploy"; exit 1; }
    echo "All required variables set..."
}

function prepare_keys() {
    echo -e ${BASTION_KEY} | base64 -d > ${BASTION_KEY_PATH}
    echo -e ${K8S_KEY} | base64 -d > ${K8S_KEY_PATH}
    for key in "${BASTION_KEY_PATH} ${K8S_KEY_PATH}"; do
        chmod 400 $key
    done
    echo "Deploy host communication ready..."
}

function redeploy() {
    UPDATE_COMMAND="cd ~/saltbot2.0 && git pull origin master"
    DEPLOY_COMMAND="kubectl -n saltbot rollout restart deployment saltbot"
    STATUS_CHECK="kubectl -n saltbot rollout status deployment saltbot"
    PROXY_COMMAND="ssh -q -o StrictHostKeyChecking=no -i ${BASTION_KEY_PATH} -p ${BASTION_PORT} -W %h:%p ${BASTION_USER}@${BASTION_HOST}"

    echo "Redeploying..."
    ssh -q -v -o StrictHostKeyChecking=no -i ${K8S_KEY_PATH} -p ${K8S_PORT} -o ProxyCommand="${PROXY_COMMAND}" ${K8S_USER}@${K8S_HOST} "${UPDATE_COMMAND}"
    ssh -q -o StrictHostKeyChecking=no -i ${K8S_KEY_PATH} -p ${K8S_PORT} -o ProxyCommand="${PROXY_COMMAND}" ${K8S_USER}@${K8S_HOST} "${DEPLOY_COMMAND}"
    ssh -q -o StrictHostKeyChecking=no -i ${K8S_KEY_PATH} -p ${K8S_PORT} -o ProxyCommand="${PROXY_COMMAND}" ${K8S_USER}@${K8S_HOST} "${STATUS_CHECK}"
    echo "Redeploy Successful!"
}

# Always delete ssh keys before exiting
trap clean EXIT

############################################################
# Main
############################################################

check_vars
prepare_keys
redeploy
