
package main

import (
	"reflect"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// Namespace Related Methods --------------------------------------------------

func (dc * DemoController) createNamespaceController() (controller *cache.Controller) {
	watchlist := cache.NewListWatchFromClient(dc.clientset.Core().RESTClient(), "namespaces", "", fields.Everything())
	_, controller = cache.NewInformer(
		watchlist,
		&v1.Namespace{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dc.namespaceAdded(obj.(*v1.Namespace))
			},
			DeleteFunc: func(obj interface{}) {
				dc.namespaceDeleted(obj.(*v1.Namespace))
			},
		},
	)
	return controller
}

func (dc * DemoController) namespaceAdded(obj *v1.Namespace) {
	glog.Infof("%s %s added", reflect.TypeOf(obj), obj.ObjectMeta.Name)

	dc.Lock()
	if _, ok := dc.Namespaces[obj.ObjectMeta.Name]; ok {
		glog.Infof("Namespace %s already exists - skipping", obj.ObjectMeta.Name)
		glog.Flush()
		dc.Unlock()
		return
	}
	dc.Namespaces[obj.ObjectMeta.Name] = obj
	dc.Unlock()

	// There's a new Namespace, add Service controller to the DemoController
	controller := dc.createServiceController(obj.ObjectMeta.Name)

	dc.Lock()
	dc.ServiceControllers[obj.ObjectMeta.Name] = controller
	dc.Unlock()

	go controller.Run(dc.stop)
}

func (dc * DemoController) namespaceDeleted(obj *v1.Namespace) {
	glog.Infof("%s %s deleted", reflect.TypeOf(obj), obj.ObjectMeta.Name)
	glog.Flush()
	dc.Lock()
	// TODO: Check if there are any services attached to this Namespace, stop the
	// the threads they're running on (requires a per thread channel for Services)
	// remove them from the DemoController
	delete(dc.Namespaces, obj.ObjectMeta.Name)
	dc.Unlock()

  // TODO: Add application specific logic here
}
