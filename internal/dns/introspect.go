package dns

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/kind/pkg/cluster"
)

var (
	gatewayGVR = schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gateways",
	}
	httpRouteGVR = schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "httproutes",
	}
)

const defaultPollInterval = 2 * time.Second

// IntrospectCluster queries the named Kind cluster for Gateway and HTTPRoute
// resources, maps HTTPRoute hostnames to Gateway addresses, and writes the
// result as a record file to recordDir/<clusterName>.json.
//
// Gateway .status.addresses may not be populated immediately after deployment;
// this function polls until at least one Gateway has an address or the timeout
// expires.
func IntrospectCluster(ctx context.Context, clusterName, recordDir string, pollTimeout time.Duration) error {
	restCfg, err := introspectRESTConfig(clusterName)
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("creating dynamic client: %w", err)
	}

	gwAddrs, err := pollGatewayAddresses(ctx, dynClient, pollTimeout)
	if err != nil {
		return err
	}
	if len(gwAddrs) == 0 {
		klog.Info("no Gateways with addresses found; skipping DNS record file")
		return nil
	}

	records, err := buildRecords(ctx, dynClient, gwAddrs)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		klog.Info("no HTTPRoute hostnames found; skipping DNS record file")
		return nil
	}

	klog.Infof("writing %d DNS record(s) for cluster %q", len(records), clusterName)
	return WriteRecordFile(recordDir, clusterName, records)
}

type gatewayKey struct {
	Namespace string
	Name      string
}

// pollGatewayAddresses lists Gateways across all namespaces, waiting until at
// least one has a populated .status.addresses field or the timeout expires.
func pollGatewayAddresses(ctx context.Context, client dynamic.Interface, timeout time.Duration) (map[gatewayKey]string, error) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		addrs, err := listGatewayAddresses(ctx, client)
		if err != nil {
			return nil, err
		}
		if len(addrs) > 0 {
			return addrs, nil
		}

		klog.V(2).Info("no Gateway addresses yet, waiting...")
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			klog.Warning("timed out waiting for Gateway addresses")
			return nil, nil
		case <-ticker.C:
		}
	}
}

// listGatewayAddresses returns a map from Gateway namespace/name to the first
// IPAddress-type value from .status.addresses.
func listGatewayAddresses(ctx context.Context, client dynamic.Interface) (map[gatewayKey]string, error) {
	list, err := client.Resource(gatewayGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing Gateways: %w", err)
	}

	addrs := make(map[gatewayKey]string)
	for i := range list.Items {
		gw := &list.Items[i]
		ip := extractGatewayIP(gw)
		if ip == "" {
			continue
		}
		key := gatewayKey{
			Namespace: gw.GetNamespace(),
			Name:      gw.GetName(),
		}
		addrs[key] = ip
		klog.V(2).Infof("Gateway %s/%s → %s", key.Namespace, key.Name, ip)
	}
	return addrs, nil
}

// extractGatewayIP returns the first IPAddress-type value from
// .status.addresses, or the first address value if no type is specified.
func extractGatewayIP(gw *unstructured.Unstructured) string {
	addresses, found, err := unstructured.NestedSlice(gw.Object, "status", "addresses")
	if err != nil || !found || len(addresses) == 0 {
		return ""
	}
	for _, addr := range addresses {
		m, ok := addr.(map[string]interface{})
		if !ok {
			continue
		}
		addrType, _, _ := unstructured.NestedString(m, "type")
		if addrType != "" && addrType != "IPAddress" {
			continue
		}
		val, _, _ := unstructured.NestedString(m, "value")
		if val != "" {
			return val
		}
	}
	return ""
}

// buildRecords lists all HTTPRoutes and maps their hostnames to Gateway IPs
// using parentRefs.
func buildRecords(ctx context.Context, client dynamic.Interface, gwAddrs map[gatewayKey]string) (map[string]string, error) {
	list, err := client.Resource(httpRouteGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing HTTPRoutes: %w", err)
	}

	records := make(map[string]string)
	for i := range list.Items {
		route := &list.Items[i]
		hostnames := extractHostnames(route)
		if len(hostnames) == 0 {
			continue
		}

		parentIPs := resolveParentIPs(route, gwAddrs)
		if len(parentIPs) == 0 {
			klog.V(2).Infof("HTTPRoute %s/%s: no matching Gateway addresses for parentRefs",
				route.GetNamespace(), route.GetName())
			continue
		}

		ip := parentIPs[0]
		for _, h := range hostnames {
			records[strings.ToLower(h)] = ip
			klog.V(2).Infof("  %s → %s", h, ip)
		}
	}
	return records, nil
}

// extractHostnames returns .spec.hostnames from an HTTPRoute.
func extractHostnames(route *unstructured.Unstructured) []string {
	vals, found, err := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
	if err != nil || !found {
		return nil
	}
	return vals
}

// resolveParentIPs looks up each parentRef of an HTTPRoute in the Gateway
// address map and returns the corresponding IPs.
func resolveParentIPs(route *unstructured.Unstructured, gwAddrs map[gatewayKey]string) []string {
	parentRefs, found, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
	if err != nil || !found {
		return nil
	}

	routeNS := route.GetNamespace()
	var ips []string
	seen := make(map[string]bool)
	for _, ref := range parentRefs {
		m, ok := ref.(map[string]interface{})
		if !ok {
			continue
		}

		group, _, _ := unstructured.NestedString(m, "group")
		kind, _, _ := unstructured.NestedString(m, "kind")
		if group != "" && group != "gateway.networking.k8s.io" {
			continue
		}
		if kind != "" && kind != "Gateway" {
			continue
		}

		name, _, _ := unstructured.NestedString(m, "name")
		if name == "" {
			continue
		}
		ns, _, _ := unstructured.NestedString(m, "namespace")
		if ns == "" {
			ns = routeNS
		}

		key := gatewayKey{Namespace: ns, Name: name}
		if ip, ok := gwAddrs[key]; ok && !seen[ip] {
			ips = append(ips, ip)
			seen[ip] = true
		}
	}
	return ips
}

func introspectRESTConfig(clusterName string) (*rest.Config, error) {
	provider := cluster.NewProvider()
	kubeConfig, err := provider.KubeConfig(clusterName, false)
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig for %q: %w", clusterName, err)
	}
	restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("parsing kubeconfig: %w", err)
	}
	return restCfg, nil
}
