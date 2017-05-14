/*

* Figure out templates and update scripts
* Figure out what to do about liveness/readiness
* Add vs Remove
* Adding/removing namespaces
* What about batching events?
* In Cluster vs Client for debugging
* Logging?
* Handling the program loop more elegantly (necessary?)
* Do we have to lock down NodePorts, or is there some way of logically mapping services?
  ADD NODE/SERVICE    - Node[] - Service{nodeport, namespace}
  UPDATE SERVICE
  DELETE NODE/SERVICE - Node[]

  Node Template: ADD/DELETE, Existing Nodes, Changing Node, Services{namespace, name, nodeport}
  Service Template: ADD/Update, Existing Nodes, Existing Services{namespace, name, nodeport}, Changing Service{namespace, name, nodeport}

TODO: 2 F5s - run two conrollers?, leave it up to the script and environment values?

Tests:
  * With, without config
  * Command line and in cluster
  * Node Up, Node issue, Node down
  * Service add, change, delete - test access at F5

*/

package main

import (
	"flag"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const ()

var (
	incluster    = flag.Bool("incluster", true, "when true it is assumed that the application is running inside a cluster")
	kubeconfig   = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	debug        = flag.Bool("debug", false, "print debug output")
)

// DemoController --------------------------------------------------------------

type DemoController struct {
	sync.RWMutex // Use this Mutex to Ensure that the DemoController values are consistent and updates to the F5s are idempotent

	stop chan struct{}

	clientset *kubernetes.Clientset

	Namespaces          map[string]*v1.Namespace // by Namespace Name
	NamespaceController *cache.Controller

	Nodes          map[string]*v1.Node // by Node Name
	NodeController *cache.Controller

	Services           map[string]*v1.Service       // by Service Name
	ServiceControllers map[string]*cache.Controller // by namespace
}

// TODO: Add an include/exclude feature for namespaces
func NewDemoController(clientset *kubernetes.Clientset) (dc * DemoController) {
	dc = &DemoController{
		clientset:          clientset,
		Namespaces:         make(map[string]*v1.Namespace),
		Nodes:              make(map[string]*v1.Node),
		Services:           make(map[string]*v1.Service),
		ServiceControllers: make(map[string]*cache.Controller),
	}

	var err error

	// Go grab the current state of the Nodes, Namespaces and Services
	nodelist, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		glog.Fatalf(err.Error())
	}

	Debugf("Got %d Nodes", len(nodelist.Items))

  // Go process the initial NodeList before addressing Node
  // status changes to avoid race conditions that could cause
  // missed service updates
	for key, node := range nodelist.Items {
		if nodeReady(&node) {
			dc.nodeAdded(&nodelist.Items[key])
		}
	}

  // Go grab all of the namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		glog.Fatalf(err.Error())
	}

	Debugf("There were %d namespaces returned", len(namespaces.Items))
	for _, namespace := range namespaces.Items {
		dc.Namespaces[namespace.ObjectMeta.Name] = &namespace
	}

	// Go setup the controllers for Nodes, Namespaces
	dc.NamespaceController = dc.createNamespaceController()
	dc.NodeController = dc.createNodeController()

	// Service Conrollers are per Namespace, make one for each
	for _, namespace := range namespaces.Items {
		Debugf("Added a ServiceController for %s", namespace.ObjectMeta.Name)
		dc.ServiceControllers[namespace.ObjectMeta.Name] = dc.createServiceController(namespace.ObjectMeta.Name)
	}

	return dc
}

// Start up all of the current list of controllers
func (dc * DemoController) RunAll() {
	dc.stop = make(chan struct{})

	dc.Lock()

	go dc.NamespaceController.Run(dc.stop)
	go dc.NodeController.Run(dc.stop)

	for namespace, _ := range dc.Namespaces {
		Debugf("Running a Service controller for %s", namespace)
		go dc.ServiceControllers[namespace].Run(dc.stop)
	}

	dc.Unlock()
}

// --------------------------------------------------------

func Debugf(format string, args ...interface{}) {
	if *debug == true {
		glog.Infof(format, args)
		glog.Flush()
	}
}

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

	dc := NewDemoController(clientset)
	dc.RunAll()

	for {
		time.Sleep(time.Second)
	}
}
