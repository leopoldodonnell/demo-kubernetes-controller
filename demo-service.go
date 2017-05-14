
package main

import (
	"reflect"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/tools/cache"
)


// Service Related Methods ----------------------------------------------------

// Services are stored as <namespace>-<servicename> to avoid collision between namespaces
func ServiceHash(obj *v1.Service) string {
	return obj.ObjectMeta.Namespace + "-" + obj.ObjectMeta.Name
}

func (dc * DemoController) createServiceController(namespace string) (controller *cache.Controller) {
	watchlist := cache.NewListWatchFromClient(dc.clientset.CoreV1().RESTClient(), "services", namespace, fields.Everything())
	_, controller = cache.NewInformer(
		watchlist,
		&v1.Service{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dc.serviceAdded(obj.(*v1.Service))
			},
			DeleteFunc: func(obj interface{}) {
				dc.serviceDeleted(obj.(*v1.Service))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				dc.serviceUpdated(oldObj.(*v1.Service), newObj.(*v1.Service))
			},
		},
	)
	return controller
}

func (dc * DemoController) serviceAdded(obj *v1.Service) {
	glog.Infof("%s %s added", reflect.TypeOf(obj), obj.ObjectMeta.Name)

	// Always add a Service if it isn't in the list, but do no processing if
	// it doesn't have any NodePorts

	dc.Lock()
	if _, ok := dc.Services[ServiceHash(obj)]; ok {
		glog.Infof("Service %s has already been added -- skipping", ServiceHash(obj))
		glog.Flush()
		dc.Unlock()
		return
	}
	dc.Services[ServiceHash(obj)] = obj
	dc.Unlock()

	if !hasNodePort(obj) {
		glog.Infof("Service %s has no NodePorts, skipping", ServiceHash(obj))
		glog.Flush()
		return
	}

	nodePorts := servicePorts(obj)
	Debugf("Service %s has the following ports: %s", ServiceHash(obj), nodePorts)

  // TODO: Add application specific logic here
}

func (dc * DemoController) serviceUpdated(oldObj *v1.Service, newObj *v1.Service) {
	glog.Infof("%s %s updated", reflect.TypeOf(newObj), ServiceHash(newObj))
	if !(hasNodePort(oldObj) || hasNodePort(newObj)) {
		glog.Infof("Service %s has no nodeports -- Skipping Update", ServiceHash(newObj))
		glog.Flush()
		return
	}

	// TODO Check if port type and/or ports have changed and update the load-balancer
	// Ports added/removed?
	oldPorts := servicePorts(oldObj)
	newPorts := servicePorts(newObj)

	// Changed Ports
	changedPorts := servicePortsChanged(oldPorts, newPorts)

	// Added Ports, Removed Ports
	addedPorts := servicePortAdded(oldPorts, newPorts)
	removedPorts := servicePortAdded(newPorts, oldPorts)

	if len(changedPorts) > 0 {
		Debugf("Service %s - the following ports were changed %s", ServiceHash(newObj), changedPorts)
	}
	if len(addedPorts) > 0 {
		Debugf("Service %s - the following ports were added %s", ServiceHash(newObj), addedPorts)
	}
	if len(removedPorts) > 0 {
		Debugf("Service %s - the following ports were removed %s", ServiceHash(newObj), removedPorts)
	}

  // TODO: Add application specific logic here
}

func (dc * DemoController) serviceDeleted(obj *v1.Service) {
	glog.Infof("%s %s deleted", reflect.TypeOf(obj), ServiceHash(obj))
	if !hasNodePort(obj) {
		glog.Infof("Service %s has no nodeports -- Skipping Update", ServiceHash(obj))
		glog.Flush()
		return
	}

	dc.Lock()
	// TODO Remove the Service from the load-balancer
	delete(dc.Services, ServiceHash(obj))
	dc.Unlock()

  // TODO: Add application specific logic here
}

func hasNodePort(svc *v1.Service) bool {
	return svc.Spec.Type == v1.ServiceTypeNodePort
}

func servicePortsChanged(oldPorts map[string]int32, newPorts map[string]int32) map[string]int32 {
	changed := make(map[string]int32)

	for port, _ := range newPorts {
		if oldPorts[port] != newPorts[port] {
			changed[port] = newPorts[port]
		}
	}
	return changed
}

// Return any ports in s2 that weren't found in s1
func servicePortAdded(svc1 map[string]int32, svc2 map[string]int32) map[string]int32 {
	added := make(map[string]int32)

	for port, value := range svc2 {
		if _, ok := svc1[port]; ok {
			added[port] = value
		}
	}
	return added
}

func servicePorts(svc *v1.Service) map[string]int32 {
	nodePorts := make(map[string]int32)

	for _, port := range svc.Spec.Ports {
		nodePorts[port.Name] = port.NodePort
	}
	return nodePorts
}
