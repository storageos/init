FROM golang:1.15.0 AS build

WORKDIR /go/src/github.com/storageos/init/
COPY . /go/src/github.com/storageos/init/
RUN make build

FROM registry.access.redhat.com/ubi8/ubi
LABEL name="StorageOS Init" \
    maintainer="support@storageos.com" \
    vendor="StorageOS" \
    version="v2.0.0" \
    release="1" \
    distribution-scope="public" \
    architecture="x86_64" \
    url="https://docs.storageos.com" \
    io.k8s.description="The StorageOS Init container prepares a node for running StorageOS." \
    io.k8s.display-name="StorageOS Init" \
    io.openshift.tags="storageos,storage,operator,pv,pvc,storageclass,persistent,csi" \
    summary="Highly-available persistent block storage for containerized applications." \
    description="StorageOS transforms commodity server or cloud based disk capacity into enterprise-class storage to run persistent workloads such as databases in containers. Provides high availability, low latency persistent block storage. No other hardware or software is required."

RUN yum -y update && \
    yum -y install --disableplugin=subscription-manager kmod

COPY scripts/ /scripts
COPY --from=build /go/src/github.com/storageos/init/LICENSE /licenses/
COPY --from=build /go/src/github.com/storageos/init/build/_output/bin/init /init
CMD /init -scripts=/scripts
