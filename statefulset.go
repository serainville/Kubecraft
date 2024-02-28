package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func fetchMinecraftServers() {
	/*
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
	*/
}

func MinecraftServer() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	svcClient := clientset.CoreV1().Services(corev1.NamespaceDefault)
	svc := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "http-80",
					Protocol: "TCP",
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
			Selector: map[string]string{"app": "t1", "s2iBuilder": "t1-s2i-1x55", "version": "v1"},
			Type:     corev1.ServiceTypeClusterIP,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "t1-v1",
			Labels: map[string]string{"app": "t1", "s2iBuilder": "t1-s2i-1x55", "version": "v1"},
		},
	}
	// Create svc
	fmt.Println("Creating svc...")
	result1, err1 := svcClient.Create(context.TODO(), svc, metav1.CreateOptions{})
	if err1 != nil {
		panic(err1)
	}
	fmt.Printf("Created svc %q.\n", result1.GetObjectMeta().GetName())

	stsClient := clientset.AppsV1().StatefulSets(corev1.NamespaceDefault)
	var replicas int32 = 3
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "t1-v1",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "t1", "s2iBuilder": "t1-s2i-1x55", "version": "v1"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "t1", "s2iBuilder": "t1-s2i-1x55", "version": "v1"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						// Image Command should be modefied, use memcached just for test
						Image: "nginx:latest",
						Name:  "container",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http-80",
							Protocol:      "TCP",
						}},
					}},
					RestartPolicy:      "Always",
					DNSPolicy:          "ClusterFirst",
					ServiceAccountName: "default",
				},
			},
		},
	}
	// Create sts
	fmt.Println("Creating sts...")
	result, err := stsClient.Create(context.TODO(), sts, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created sts %q.\n", result.GetObjectMeta().GetName())
}
