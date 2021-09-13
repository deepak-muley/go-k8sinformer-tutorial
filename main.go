package main

import (
	"fmt"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	POD_LABEL = "name=sherlock-operators"
)

var (
	totalPodAddCount    = 0
	totalPodUpdateCount = 0
	totalPodDeleteCount = 0

	totalCMAddCount    = 0
	totalCMUpdateCount = 0
	totalCMDeleteCount = 0

	totalUnknownAddCount    = 0
	totalUnknownUpdateCount = 0
	totalUnknownDeleteCount = 0
)

func main() {
	log.Print("Shared Informer app started")
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	defer runtime.HandleCrash()

	stopper := make(chan struct{})
	defer close(stopper)

	podInformer := AddPodInformer(clientset, stopper)
	configMapInformer := AddConfigMapInformer(clientset, stopper)

	if !cache.WaitForCacheSync(stopper, podInformer.HasSynced, configMapInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}

func AddPodInformer(clientset *kubernetes.Clientset, stopper chan struct{}) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	go informer.Run(stopper)
	return informer
}

func AddConfigMapInformer(clientset *kubernetes.Clientset, stopper chan struct{}) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	go informer.Run(stopper)
	return informer
}

// onAdd is the function executed when the kubernetes informer notified the
// presence of a new kubernetes Pod in the cluster
func onAdd(obj interface{}) {
	switch v := obj.(type) {
	case *corev1.Pod:
		totalPodAddCount += 1
		fmt.Printf("[%d] Added Pod Namespace: %s | Name: %s\n", totalPodAddCount, v.GetNamespace(), v.GetName())
	case *corev1.ConfigMap:
		totalCMAddCount += 1
		fmt.Printf("[%d] Added ConfigMap Namespace: %s | Name: %s\n", totalCMAddCount, v.GetNamespace(), v.GetName())
	default:
		totalUnknownAddCount += 1
		fmt.Printf("[%d] Added Unknown type object: %#v", totalUnknownAddCount, v)
	}
}

// onUpdate is the function executed when the kubernetes informer notified the
// update of a  kubernetes Pod in the cluster
func onUpdate(oldObj, newObj interface{}) {
	switch v := oldObj.(type) {
	case *corev1.Pod:
		totalPodUpdateCount += 1
		// Cast the obj as Pod
		oldPod := oldObj.(*corev1.Pod)
		newPod := newObj.(*corev1.Pod)
		fmt.Printf("[%d] Updated Old Pod Namespace: %s | Name: %s\n", totalPodUpdateCount, oldPod.GetNamespace(), oldPod.GetName())
		fmt.Printf("[%d] Updated New Pod Namespace: %s | Name: %s\n", totalPodUpdateCount, newPod.GetNamespace(), newPod.GetName())
	case *corev1.ConfigMap:
		totalCMUpdateCount += 1
		// Cast the obj as ConfigMap
		oldCM := oldObj.(*corev1.ConfigMap)
		newCM := newObj.(*corev1.ConfigMap)
		fmt.Printf("[%d] Updated Old ConfigMap Namespace: %s | Name: %s\n", totalCMUpdateCount, oldCM.GetNamespace(), oldCM.GetName())
		fmt.Printf("[%d] Updated New ConfigMap Namespace: %s | Name: %s\n", totalCMUpdateCount, newCM.GetNamespace(), newCM.GetName())
	default:
		totalUnknownUpdateCount += 1
		fmt.Printf("[%d] Updated Unknown type object: %#v", totalUnknownUpdateCount, v)
	}
}

// onDelete is the function executed when the kubernetes informer notified the
// deletion of a new kubernetes Pod in the cluster
func onDelete(obj interface{}) {
	switch v := obj.(type) {
	case *corev1.Pod:
		totalPodDeleteCount += 1
		fmt.Printf("[%d] Deleted Pod Namespace: %s | Name: %s\n", totalPodDeleteCount, v.GetNamespace(), v.GetName())
	case *corev1.ConfigMap:
		totalCMDeleteCount += 1
		fmt.Printf("[%d] Deleted ConfigMap Namespace: %s | Name: %s\n", totalCMDeleteCount, v.GetNamespace(), v.GetName())
	default:
		totalUnknownDeleteCount += 1
		fmt.Printf("[%d] Deleted Unknown type object: %#v", totalUnknownDeleteCount, v)
	}
}
