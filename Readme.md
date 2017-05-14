# Demo Kubernetes Controller

The `Demo Kubernetes Controller` is a Kubernetes controller that provides an example of a Kubernetes client, writtin
in `go`, that can be run from the command line, or from within a cluster to respond to changes in Nodes, Namespaces and
Services.

* when *namespaces* change, *demo-controller* will update its subscription to *services* in that namespace
* when *services* change, *demo-controller* will announce the change and perform and necessary bookkeeping
* when *nodes* change, *ingress-mapper* will nnounce the change and perform and necessary bookkeeping

## Usage

## Application Notes

### General

The application is written in `go` and built using the [client-go](https://github.com/k8s.io/client-go) framework. It may either
be run outside of a cluster as a stand-alone application, or within a cluster in a container. The application uses the
(flag)[https://golang.org/pkg/flag/] package for handling command-line arguments and  uses the golang
[glog](https://godoc.org/github.com/golang/glog) logging package that enables multiple levels of logging: Info, Warning, Error,
and Fatal. The application is multi-threaded, so care is taken to ensure that updates to datastructures.

### Program Flow

The general flow of the program is as follows.


```
Initialize the configuation from the command line
Create An Instance of a DemoController
  Get existing Nodes
  Get existing Namespaces
  Create a kubernetes controller for Nodes
  Create a kubernetes controller for Namespaces
  Create a kubernetes controller Services per each Namespace
Run all of the controllers, each in its own thread and wait for the application to exit
```

All of the heavy-lifting is done by performing bookkeeping from controller `AddFunc`, `DeleteFunc` and `UpdateFunc` functions, 
then calling out to an external script that is customized to affect the appropriate changes in an external load-balancer.
