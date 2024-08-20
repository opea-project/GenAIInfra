#!/usr/bin/env bash
set -e

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
cd $SCRIPT_DIR && cd ../../
GenAIInfra_DIR=$(pwd)
cd $GenAIInfra_DIR/microservices-connector

# kubectl delete -k config/samples/
helm delete -n system gmc
kubectl delete crd gmconnectors.gmc.opea.io