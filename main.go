package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"

	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func handleSignals(stopCh chan struct{}) {
	chSigs = make(chan os.Signal)
	signal.Notify(chSigs, os.Interrupt)
	<-chSigs
	close(stopCh)
}

var (
	stopCh      chan struct{}
	chSigs      chan os.Signal
	cfgInformer cache.SharedIndexInformer
	// recorder    events.EventRecorder
	recorder record.EventRecorder
	queue    chan RecordObj
)

type RecordObj struct {
	object  runtime.Object
	reason  string
	message string
}

func recordDaemon(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case r := <-queue:
			fmt.Printf("%#v\n", r)
			// recorder.Eventf(r.object, nil, v1.EventTypeNormal, r.reason, "", r.message)
			recorder.Event(r.object, v1.EventTypeNormal, r.reason, r.message)
		}
	}

}
func main() {
	fmt.Println("Program started.")
	queue = make(chan RecordObj, 50)
	stopCh = make(chan struct{})
	go handleSignals(stopCh)

	if len(os.Args) != 2 {
		fmt.Println("Usage: <program> <kube.config>")
		return
	}

	kubeConf := os.Args[1]
	config, err := clientcmd.BuildConfigFromFlags("", kubeConf)
	if nil != err {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// matchLabelSelector := func(opts *metav1.ListOptions) {
	// 	// opts.LabelSelector = "somelabel=xyz"
	// }
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second,
		// informers.WithNamespace("default"), // for test: limit resource scope to default.
		// informers.WithTweakListOptions(matchLabelSelector),
	)

	cfgInformer = sharedInformerFactory.Core().V1().ConfigMaps().Informer()

	cfgInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			abc := obj.(*v1.ConfigMap)
			fmt.Printf("add %s %s\n", abc.Namespace, abc.Name)
			queue <- RecordObj{
				object:  abc,
				reason:  "Deployed",
				message: "object addition has been notified by informer",
			}
			abc.GetObjectKind()
			// recorder.Eventf(abc, nil, v1.EventTypeNormal, "Deployed", "", "object addition has been notified by informer")
		},
		UpdateFunc: func(old, cur interface{}) {
			orig := old.(*v1.ConfigMap)
			newa := cur.(*v1.ConfigMap)
			if orig.GetUID() != newa.GetUID() || orig.GetResourceVersion() != newa.GetResourceVersion() {
				fmt.Printf("%s %s -> %s %s\n", orig.Namespace, orig.Name, newa.Namespace, newa.Name)

				queue <- RecordObj{
					object:  newa,
					reason:  "Updated",
					message: "object update has been notified by informer",
				}
				// recorder.Eventf(newa, nil, v1.EventTypeNormal, "Updated", "", "object update has been notified by informer")
			}
		},
		DeleteFunc: func(obj interface{}) {
			abc := obj.(*v1.ConfigMap)
			fmt.Printf("delete %s %s\n", abc.Namespace, abc.Name)
			queue <- RecordObj{
				object:  abc,
				reason:  "Deleted",
				message: "object deletion has been notified by informer",
			}
			// recorder.Eventf(abc, nil, v1.EventTypeNormal, "Deleted", "", "object deletion has been notified by informer")
		},
	})

	// eba := events.NewEventBroadcasterAdapter(clientset)
	// recorder = eba.NewRecorder("my-event-recorder")
	// eba.StartRecordingToSink(stopCh)

	eba := record.NewBroadcaster()
	eba.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	recorder = eba.NewRecorder(scheme.Scheme, v1.EventSource{Component: "my-event-recorder"})

	sharedInformerFactory.Start(stopCh)

	go recordDaemon(stopCh)
	doNilLoop()
}

func doNilLoop() {
	for {
		select {
		case <-stopCh:
			return
		case <-time.After(1 * time.Second):
			// do nothing
		}
	}
}
