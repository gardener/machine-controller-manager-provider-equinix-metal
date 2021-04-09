# machine-controller-manager-provider-equinix-metal
Out of tree (controller based) implementation for `equinix-metal` as a new provider.

## About

- The [Equinix Metal](https://metal.equinix.com), formerly [Packet](https://packet.com), out-of-tree provider implements the interface defined at [MCM OOT Driver](https://github.com/gardener/machine-controller-manager/blob/master/pkg/util/provider/driver/driver.go)

## Fundamental Design Principles:
Following are the basic principles kept in mind while developing the external plugin.
* Communication between this Machine Controller (MC) and Machine Controller Manager (MCM) is achieved using the Kubernetes native declarative approach.
* Machine Controller (MC) behaves as the controller used to interact with the cloud provider and manage the VMs corresponding to the machine objects.
* Machine Controller Manager (MCM) deals with higher level objects such as machine-set and machine-deployment objects.

## Support for a new provider
- Steps to be followed while implementing a new provider are mentioned [here](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md)

## Testing the Equinix Metal Out of Tree Provider

Prerequisites:

* [git](https://git-scm.com)
* Three open terminal windows
* [go](http://golang.org), v1.15 or newer
* A deployed Kubernetes cluster to control machines. Any cluster will do, including one deployed on [Equinix Metal](https://metal.equinix.com), locally, or even [kind](https://kind.sigs.k8s.io) or [k3d](https://k3d.io)

Theoretically, you need a target cluster as well, a cluster that the newly deployed machines will join. This cluster's control plane
should have the following characteristics:

* reachable from any new machines you create
* userData that you supply to the newly-created machines is sufficient for them to join

However, since this local testing is not testing the entire flow, just the creation/deletion of machines, we will ignore cluster joining,
and just check that the machine was created with the right userData.

For your deployed Kubernetes cluster, ensure that you know the following:

* path to the kubeconfig for the target cluster; for testing purposes, this will be your control cluster
* path to the kubeconfig for the control cluster which holds the machine custom resource objects
* namespace in the control cluster where you will deploy `Machine` and `MachineDeployment` custom resource objects, defaults to `default`

Process:

1. In the first terminal window:
   1. Get the latest copy of the [Machine Controller Manager (MCM)](https://github.com/gardener/machine-controller-manager), if you do not have it:
      ```bash
      mkdir -p $GOPATH/src/github.com/gardener
      cd $_
      git clone https://github.com/gardener/machine-controller-manager
      cd machine-controller-manager
      ```
   1. Deploy the required CRDs to the _control_ cluster:
        ```bash
        kubectl --kubeconfig=${CONTROL_KUBECONFIG} apply -f kubernetes/crds/
        ```
   1. Run the machine-controller-manager in the `cmi-client` branch:
        ```bash
        make start TARGET_KUBECONFIG=path/to/target/kubeconfig CONTROL_KUBECONFIG=path/to/control/kubeconfig CONTROL_NAMESPACE=control_namespace
        ```
1. In the second terminal window:
   1. Get this repository, if you do not have it already
      ```bash
      mkdir -p $GOPATH/src/github.com/gardener
      cd $_
      git clone https://github.com/gardener/machine-controller-manager-provider-equinix-metal
      cd machine-controller-manager-provider-equinix-metal
      ```
   1. Start the driver in this repository with `make start`, setting the following make variables:
      ```
      make start TARGET_KUBECONFIG=path/to/target/kubeconfig CONTROL_KUBECONFIG=path/to/control/kubeconfig CONTROL_NAMESPACE=control_namespace`
      ```
1. In the third terminal:
   1. Change directory to this repository:
      ```bash
      cd $GOPATH/src/github.com/gardener/machine-controller-manager-provider-equinix-metal
      ```
   1. All activities will be against the _control_ cluster. With each command, you can run `kubectl --kubeconfig=${CONTROL_KUBECONFIG}`. However, to simplify the commands, set the environment variable:
      ```bash
      export KUBECONFIG=${CONTROL_KUBECONFIG}
      ```
   1. Fill in the object files given below and deploy them:
        ```bash
        kubectl apply -f kubernetes/machine-class.yaml
        ```
   1. Fill in the Kubernetes `Secret` with the Equinix Metal API key and deploy:
        ```bash
        kubectl apply -f kubernetes/secret.yaml
        ```
   1. Deploy a `Machine` object and make sure it is created and has the userData you supplied.
        ```bash
        kubectl apply -f kubernetes/machine.yaml
        ```
   1. Once the `Machine` has passed testing, deploy a `MachineDeployment` and make sure it is created and has the userData you supplied. If you provided it with an actual joinable target cluster and userData to join it, wait until all of the machines join that cluster successfully:
        ```bash
        kubectl apply -f kubernetes/machine-deployment.yaml
        ```
   1. Clean up by deleting both the `machine` and `machine-deployment` object after use.
        ```bash
        kubectl delete -f kubernetes/machine.yaml
        kubectl delete -f kubernetes/machine-deployment.yaml
        ```
1. Stop the processes in the first and second terminal windows
