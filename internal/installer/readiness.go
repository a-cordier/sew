package installer

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// WaitForReady waits until all pods for the release in the namespace are Ready or timeout.
func WaitForReady(ctx context.Context, releaseName, namespace string, timeout time.Duration, matchLabels map[string]string) error {
	config, err := getKubeConfig()
	if err != nil {
		return fmt.Errorf("kube config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("kubernetes client: %w", err)
	}
	if namespace == "" {
		namespace = "default"
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for %q to be ready", releaseName)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ready, err := checkPodsReady(ctx, clientset, releaseName, namespace, matchLabels)
			if err != nil {
				return err
			}
			if ready {
				return nil
			}
		}
	}
}

func getKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

func checkPodsReady(ctx context.Context, clientset *kubernetes.Clientset, releaseName, namespace string, matchLabels map[string]string) (bool, error) {
	var selector string
	if len(matchLabels) > 0 {
		selector = labels.SelectorFromSet(matchLabels).String()
	} else {
		selector = "app.kubernetes.io/instance=" + releaseName
	}
	list, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return false, fmt.Errorf("listing pods: %w", err)
	}
	if len(list.Items) == 0 {
		return false, nil
	}
	for _, pod := range list.Items {
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				return false, nil
			}
		}
	}
	return true, nil
}
