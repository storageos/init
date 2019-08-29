#!/usr/bin/env bash

set -Eeuxo pipefail

prepare_host() {
    sudo apt -y update
    sudo apt -y install linux-modules-extra-$(uname -r)
    sudo mount --make-shared /sys
    sudo mount --make-shared /
    sudo mount --make-shared /dev
}

run_kind() {
    echo "Download kind binary..."
    wget -O kind 'https://docs.google.com/uc?export=download&id=1-oy-ui0ZE_T3Fglz1c8ZgnW8U-A4yS8u' --no-check-certificate && chmod +x kind && sudo mv kind /usr/local/bin/

    echo "Download kubectl..."
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/"${K8S_VERSION}"/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
    echo

    echo "Create Kubernetes cluster with kind..."
    # kind create cluster --image=kindest/node:"$K8S_VERSION"
    kind create cluster --image storageos/kind-node:"$K8S_VERSION" --name kind-1

    echo "Export kubeconfig..."
    export KUBECONFIG="$(kind get kubeconfig-path --name="kind-1")"
    echo

    echo "Get cluster info..."
    kubectl cluster-info
    echo

    echo "Wait for kubernetes to be ready"
    JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done
    echo
}

main() {
    make unittest
    make image

    prepare_host
    run_kind

    echo "Ready for e2e testing"

    # Copy the build container image into KinD.
    x=$(docker ps -f name=kind-1-control-plane -q)
    docker save storageos/init:test > init.tar
    docker cp init.tar $x:/init.tar

    # containerd load image from tar archive (KinD with containerd).
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/storageos/init:test /init.tar"

    kubectl apply -f pod.yaml

    echo "Waiting for the test-pod to run"
    until kubectl get pod test-pod --no-headers -o go-template='{{.status.phase}}' | grep -q Running; do sleep 5; done
    echo "test-pod found running"

    echo "init container logs"
    kubectl logs test-pod -c storageos-init
    echo

    echo "Checking init container exit code"
    exitCode=$(kubectl get pod test-pod --no-headers -o go-template='{{(index .status.initContainerStatuses 0).state.terminated.exitCode}}')
    if [ "$exitCode" == "0" ]; then
        echo "init successful!"
        exit 0
    fi
    echo "init failed"
    exit 1
}

main
