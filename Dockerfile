FROM golang:1.12.7 AS build
WORKDIR /go/src/github.com/storageos/init/
COPY . /go/src/github.com/storageos/init/
RUN make build

FROM registry.access.redhat.com/ubi8/ubi
RUN yum -y update && \
    yum -y install --disableplugin=subscription-manager kmod
COPY scripts/ /scripts
COPY --from=build /go/src/github.com/storageos/init/build/_output/bin/init /init
CMD /init -scripts=/scripts
