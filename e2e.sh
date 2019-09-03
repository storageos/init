#!/usr/bin/env bash

set -Eeuxo pipefail

prepare_host() {
    sudo apt -y update
    sudo apt -y install linux-modules-extra-$(uname -r)
    sudo mount --make-shared /sys
    sudo mount --make-shared /
    sudo mount --make-shared /dev
    docker run --name enable_lio --privileged --rm --cap-add=SYS_ADMIN -v /lib/modules:/lib/modules -v /sys:/sys:rshared storageos/init:0.2
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

# This tests the dataplane upgrade on the host, independent of any k8s cluster.
nok8s_dpupgrade() {
    # Run StorageOS and create old dp database.
    docker pull storageos/node:1.3.0
    NODE_IMAGE=storageos/node:1.3.0 bash run-stos.sh
    sleep 10

    # Wait until storageos node is healthy.
    health=$(docker inspect storageos --type container | jq '.[0] | .State.Health.Status')
    until [ $health == '"healthy"' ]
    do
        health=$(docker inspect storageos --type container | jq '.[0] | .State.Health.Status')
        echo "storageos status $health"
        echo "waiting for storageos node to be healthy..."
        sleep 3
    done

    echo
    echo "storageos node is healthy"

    # Stop storageos.
    docker rm storageos -f

    # Run dp upgrade on old database.
    echo
    echo "Attempting dp upgrade"
    UPGRADELOGS=/tmp/dpupgrade.log
    make run > $UPGRADELOGS
    if ! grep "successfully upgraded database" $UPGRADELOGS; then
        echo "dpupgrade failed!"
        echo
        echo "init logs:"
        cat $UPGRADELOGS
        exit 1
    fi
    echo "upgrade successful"
}

# This tests the dataplane upgrade in a k8s cluster by updating node 1.3.0
# storageos node to 1.4.0.
k8s_dpupgrade() {
    # Get KinD container id.
    x=$(docker ps -f name=kind-1-control-plane -q)

    # Re-tag node 1.3.0 as storageos/node:1.4.0 and copy into KinD.
    # This is temporary until storageos/node:1.4.0 is released.
    docker tag storageos/node:1.3.0 storageos/node:1.4.0
    docker save storageos/node:1.4.0 > stos140.tar
    docker cp stos140.tar $x:/stos140.tar
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/storageos/node:1.4.0 /stos140.tar"

    # Get storageos pod.
    stospod=$(kubectl get pods --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep storageos)

    echo "Waiting for the storageos container to be ready"
    # First and only container in this pod is the storageos container.
    until kubectl get pod $stospod --template='{{ (index .status.containerStatuses 0).ready }}' | grep -q true; do sleep 5; done
    echo "storageos pod found ready"

    sleep 5

    # Patch DaemonSet with new storageos version.
    kubectl set image ds/storageos-daemonset storageos=storageos/node:1.4.0

    # kill the pod for the update to apply.
    kubectl delete pod $stospod

    sleep 5

    # Get new pod name.
    stospod=$(kubectl get pods --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep storageos)

    echo "Waiting for the storageos pod to be ready"
    until kubectl get pod $stospod --template='{{ (index .status.containerStatuses 0).ready }}' | grep -q true; do sleep 5; done
    echo "storageos pod found ready"

    echo "checking init container logs"
    UPGRADELOGS=/tmp/dpupgrade-k8s.log
    kubectl logs $stospod -c storageos-init > $UPGRADELOGS
    echo

    if ! grep "successfully upgraded database" $UPGRADELOGS; then
        echo "dpupgrade failed!"
        echo
        echo "init logs:"
        cat $UPGRADELOGS
        exit 1
    fi
    echo "upgrade successful"
    echo
}

main() {
    make unittest
    make image

    prepare_host
    run_kind

    echo "Ready for e2e testing"

    # Run out of k8s dpupgrade test.
    echo
    echo "No k8s dp upgrade test"
    echo
    nok8s_dpupgrade

    echo
    echo "Prepare for k8s test"
    echo

    # Copy the init container image into KinD.
    x=$(docker ps -f name=kind-1-control-plane -q)
    docker save storageos/init:test > init.tar
    docker cp init.tar $x:/init.tar

    # containerd load image from tar archive (KinD with containerd).
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/storageos/init:test /init.tar"

    # Load storageos node 1.3.0 container image into KinD.
    docker pull storageos/node:1.3.0
    docker save storageos/node:1.3.0 > stos130.tar
    docker cp stos130.tar $x:/stos130.tar
    docker exec $x bash -c "ctr -n k8s.io images import --base-name docker.io/storageos/node:1.3.0 /stos130.tar"

    # Create storageos daemonset with node 1.3.0
    kubectl apply -f daemonset.yaml
    sleep 5

    # Run dp upgrade test.
    k8s_dpupgrade

    # Get pod name.
    stospod=$(kubectl get pods --template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' | grep storageos)

    # NOTE: Once dataplane upgrade test is no longer needed, remove the
    # k8s_dpupgrade call above and uncomment the following wait code.

    # echo "Waiting for the storageos pod to be ready"
    # until kubectl get pod $stospod --template='{{ (index .status.containerStatuses 0).ready }}' | grep -q true; do sleep 5; done
    # # until kubectl get pod $stospod --no-headers -o go-template='{{.status.phase}}' | grep -q Running; do sleep 5; done
    # echo "storageos pod found ready"

    echo
    echo "init container logs:"
    kubectl logs $stospod -c storageos-init
    echo

    echo "Checking init container exit code"
    exitCode=$(kubectl get pod $stospod --no-headers -o go-template='{{(index .status.initContainerStatuses 0).state.terminated.exitCode}}')
    if [ "$exitCode" == "0" ]; then
        echo "init successful!"
        exit 0
    fi
    echo "init failed"
    exit 1
}

main
