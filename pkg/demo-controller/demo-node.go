package controller

import (
	"reflect"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// Node Related Methods -------------------------------------------------------

func (dc *DemoController) createNodeController() (controller *cache.Controller) {
	watchlist := cache.NewListWatchFromClient(dc.clientset.Core().RESTClient(), "nodes", "",
		fields.Everything())
	_, controller = cache.NewInformer(
		watchlist,
		&v1.Node{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dc.nodeAdded(obj.(*v1.Node))
			},
			DeleteFunc: func(obj interface{}) {
				dc.nodeDeleted(obj.(*v1.Node))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				dc.nodeUpdated(oldObj.(*v1.Node), newObj.(*v1.Node))
			},
		},
	)
	return controller
}

func (dc *DemoController) nodeAdded(obj *v1.Node) {
	glog.Infof("%s %s added", reflect.TypeOf(obj), obj.ObjectMeta.Name)

	dc.Lock()
	if _, ok := dc.Nodes[obj.ObjectMeta.Name]; ok {
		glog.Infof("Node %s has already been added -- skipping", obj.ObjectMeta.Name)
		glog.Flush()
		dc.Unlock()
		return
	}
	dc.Nodes[obj.ObjectMeta.Name] = obj
	dc.Unlock()

	// TODO: Add application specific logic here
}

func (dc *DemoController) nodeDeleted(obj *v1.Node) {
	glog.Infof("%s %s deleted", reflect.TypeOf(obj), obj.ObjectMeta.Name)
	glog.Flush()
	dc.Lock()
	delete(dc.Nodes, obj.ObjectMeta.Name)
	dc.Unlock()

	// TODO: Add application specific logic here
}

func (dc *DemoController) nodeUpdated(oldObj *v1.Node, newObj *v1.Node) {
	glog.Infof("%s %s updated", reflect.TypeOf(newObj), newObj.ObjectMeta.Name)
	glog.Flush()
	// TODO Handle changes to Node readiness

	switch {
	case !nodeReady(oldObj) && nodeReady(newObj):
		glog.Infof("Node %s has become ready")
	case nodeReady(oldObj) && !nodeReady(newObj):
		glog.Infof("Node %s has become not ready")
	default:
		glog.Infof("Node %s update -- skipping", newObj.ObjectMeta.Name)
		return
	}

	// TODO: Add application specific logic here
}

// Return true iff the Node is in the NodeReady state
func nodeReady(node *v1.Node) bool {
	Debugf("Checking Node %s", node.ObjectMeta.Name)

	for _, condition := range node.Status.Conditions {
		Debugf("Node %s has condition = ", node.ObjectMeta.Name, condition)

		if condition.Type == v1.NodeReady {
			glog.Infof("Node %s is ready", node.ObjectMeta.Name)
			return true
		}
	}
	glog.Infof("Node %s is NOT ready", node.ObjectMeta.Name)
	glog.Flush()
	return false
}

func getNodeAddress(node *v1.Node, addressType v1.NodeAddressType) string {
	for _, address := range node.Status.Addresses {
		if address.Type == addressType {
			return address.Address
		}
	}
	return ""
}

func getAllNodeAddresses(nodes map[string]*v1.Node, addressType v1.NodeAddressType) (addresses []string) {
	for _, node := range nodes {
		Debugf("getAllNodeAddresses: Adding IP: %s", getNodeAddress(node, addressType))
		addresses = append(addresses, getNodeAddress(node, addressType))
	}
	return addresses
}
