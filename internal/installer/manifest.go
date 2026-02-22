package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-cordier/sew/core"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	fieldManager = "sew"
)

// ManifestInstaller installs components from plain Kubernetes manifest files.
type ManifestInstaller struct{}

func getConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

func (m *ManifestInstaller) Install(ctx context.Context, comp core.Component, dir string) error {
	if comp.Manifest == nil {
		return fmt.Errorf("component %q has no manifest spec", comp.Name)
	}
	if len(comp.Manifest.Files) == 0 {
		return fmt.Errorf("component %q has no manifest files", comp.Name)
	}

	config, err := getConfig()
	if err != nil {
		return fmt.Errorf("kube config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("kubernetes client: %w", err)
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("dynamic client: %w", err)
	}
	discoveryMapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))

	namespace := comp.Namespace
	if namespace == "" {
		namespace = "default"
	}
	if namespace != "default" {
		_, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create namespace %q: %w", namespace, err)
		}
	}

	for _, f := range comp.Manifest.Files {
		path := filepath.Join(dir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %q: %w", path, err)
		}
		docs := splitYAMLDocuments(data)
		for _, doc := range docs {
			doc = strings.TrimSpace(doc)
			if doc == "" {
				continue
			}
			obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
			if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
				return fmt.Errorf("decoding manifest in %q: %w", path, err)
			}
			gvk := obj.GroupVersionKind()
			if gvk.Kind == "" {
				continue
			}
			mapping, err := discoveryMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
			if err != nil {
				return fmt.Errorf("mapping %s: %w", gvk, err)
			}
			if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
				obj.SetNamespace(namespace)
			}
			gvr := mapping.Resource
			var ri dynamic.ResourceInterface
			if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
				ri = dynClient.Resource(gvr).Namespace(obj.GetNamespace())
			} else {
				ri = dynClient.Resource(gvr)
			}
			_, err = ri.Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{FieldManager: fieldManager})
			if err != nil {
				return fmt.Errorf("apply %s %s: %w", gvk.Kind, obj.GetName(), err)
			}
		}
	}
	return nil
}

func splitYAMLDocuments(data []byte) []string {
	var docs []string
	var buf strings.Builder
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if buf.Len() > 0 {
				docs = append(docs, buf.String())
				buf.Reset()
			}
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(line)
	}
	if buf.Len() > 0 {
		docs = append(docs, buf.String())
	}
	return docs
}

// Uninstall re-reads manifest files and deletes the resources.
func (m *ManifestInstaller) Uninstall(_ context.Context, comp core.Component) error {
	if comp.Manifest == nil || len(comp.Manifest.Files) == 0 {
		return nil
	}
	// Resolve dir: we don't have it in Uninstall. We need resolved.Dir in cmd.
	// For now we only support uninstall when we can resolve context again, or we
	// could require absolute paths in manifest files. The cmd layer doesn't pass dir to Uninstall.
	// So we cannot re-read files here. Return a clear error or skip delete.
	return fmt.Errorf("manifest uninstall not implemented (component %q): re-run with context to delete", comp.Name)
}
