#############      builder                                  #############
FROM golang:1.20.4 AS builder

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-equinix-metal
COPY . .

RUN .ci/build

#############      base                                     #############
FROM gcr.io/distroless/static-debian11:nonroot as base
WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-equinix-metal/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
