package main

import (
	"flag"
	"time"

	"github.com/golang/glog"

	controller "github.com/leopoldodonnell/demo-kubernetes-controller/pkg/demo-controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const ()

var (
	incluster  = flag.Bool("incluster", true, "when true it is assumed that the application is running inside a cluster")
	kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	debug      = flag.Bool("debug", false, "print debug output")
)

// --------------------------------------------------------

func main() {
	flag.Parse()

	var err error
	var config *rest.Config

	if *incluster == true {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	if err != nil {
		glog.Fatalf(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf(err.Error())
	}

	controller.Debug = *debug

	dc := controller.NewDemoController(clientset)
	dc.RunAll()

	for {
		time.Sleep(time.Second)
	}
}
