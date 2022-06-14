package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/events"
)

// 处理信号，并及时关闭stopCh
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
	recorder    events.EventRecorder
)

func main() {

	stopCh = make(chan struct{})
	go handleSignals(stopCh)

	if len(os.Args) != 2 {
		fmt.Println("Usage: <program> <kube.config>")
		return
	}

	// 创建clientset，共informer和events 共用。
	kubeConf := os.Args[1]
	config, err := clientcmd.BuildConfigFromFlags("", kubeConf)
	if nil != err {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Labelselector可以用来过滤仅需要关注的资源
	// matchLabelSelector := func(opts *metav1.ListOptions) {
	// 	// opts.LabelSelector = "somelabel=xyz"
	// }
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second,
		informers.WithNamespace("default"), // for test: limit resource scope to default.
		// informers.WithTweakListOptions(matchLabelSelector),
	)

	// 创建针对configmap的informer
	cfgInformer = sharedInformerFactory.Core().V1().ConfigMaps().Informer()

	// 为informer添加处理函数
	cfgInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			abc := obj.(*v1.ConfigMap)
			fmt.Printf("add %s %s\n", abc.Namespace, abc.Name)
			recorder.Eventf(abc, nil, v1.EventTypeNormal, "Deployed", "", "object addition has been notified by informer")
		},
		UpdateFunc: func(old, cur interface{}) {
			orig := old.(*v1.ConfigMap)
			newa := cur.(*v1.ConfigMap)
			if orig.GetUID() != newa.GetUID() || orig.GetResourceVersion() != newa.GetResourceVersion() {
				fmt.Printf("%s %s -> %s %s\n", orig.Namespace, orig.Name, newa.Namespace, newa.Name)
				recorder.Eventf(newa, nil, v1.EventTypeNormal, "Deployed", "", "object update has been notified by informer")
			}
		},
		DeleteFunc: func(obj interface{}) {
			abc := obj.(*v1.ConfigMap)
			fmt.Printf("delete %s %s\n", abc.Namespace, abc.Name)
			recorder.Eventf(abc, nil, v1.EventTypeNormal, "Deployed", "", "object deletion has been notified by informer")
		},
	})

	// 创建EventRecorder 过程。
	eba := events.NewEventBroadcasterAdapter(clientset)
	recorder = eba.NewRecorder("my-event-recorder")

	// 开始工作
	eba.StartRecordingToSink(stopCh)
	sharedInformerFactory.Start(stopCh)

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
