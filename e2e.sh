#!/usr/bin/env bash

set -Eeuxo pipefail

cluster="init"

prepare_host() {
    sudo apt -y update
    sudo apt -y install linux-modules-extra-$(uname -r)
    sudo mount --make-shared /sys
    sudo mount --make-shared /
    sudo mount --make-shared /dev
}

run_kind() {
    echo "Download kind binary..."
    wget -O kind 'https://kind.sigs.k8s.io/dl/v0.8.1/kind-linux-amd64' --no-check-certificate && chmod +x kind && sudo mv kind /usr/local/bin/

    echo "Download kubectl..."
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/"${K8S_VERSION}"/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
    echo

    if [ $(kind get clusters | grep -c ^$cluster$) -eq 0 ]; then
        echo "Create Kubernetes cluster with kind..."
        # kind create cluster --image=kindest/node:"$K8S_VERSION"
        kind create cluster --image storageos/kind-node:"$K8S_VERSION" --name "$cluster"
    fi

    echo "Export kubeconfig..."
    kind get kubeconfig --name="$cluster" > kubeconfig.yaml
    export KUBECONFIG="kubeconfig.yaml"
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

    echo
    echo "Prepare for k8s test"
    echo

    # Copy the init container image into KinD.
    x=$(docker ps -f name=${cluster}-control-plane -q)
    docker save storageos/init:test > init.tar
    docker cp init.tar $x:/init.tar

    # containerd load image from tar archive (KinD with containerd).
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/storageos/init:test /init.tar"

    # Use busybox in place of the node container.
    docker pull busybox:1.32
    docker save busybox:1.32 > busybox.tar
    docker cp busybox.tar $x:/busybox.tar
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/busybox:1.32 /busybox.tar"

    # Create storageos daemonset with node v1.5.3
    kubectl apply -f daemonset.yaml
    sleep 5

    # Get pod name.
    stospod=$(kubectl get pods --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep storageos)

    echo "Waiting for the storageos pod to be ready"
    until kubectl get pod $stospod --template='{{ (index .status.containerStatuses 0).ready }}' | grep -q true; do sleep 5; done
    # until kubectl get pod $stospod --no-headers -o go-template='{{.status.phase}}' | grep -q Running; do sleep 5; done
    echo "storageos pod found ready"

    echo
    echo "init container logs:"
    kubectl logs $stospod -c storageos-init
    echo

    echo "Checking init container exit code"
    exitCode=$(kubectl get pod $stospod --no-headers -o go-template='{{(index .status.initContainerStatuses 0).state.terminated.exitCode}}')
    kubectl delete -f daemonset.yaml
    if [ "$exitCode" == "0" ]; then
        echo "init successful!"
        exit 0
    fi
    echo "init failed"
    exit 1
}

main
