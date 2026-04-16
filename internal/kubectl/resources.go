package kubectl

import "strings"

// resourceAliases maps short names / aliases to canonical kubectl resource names.
var resourceAliases = map[string]string{
	"po":             "pods",
	"pod":            "pods",
	"svc":            "services",
	"service":        "services",
	"deploy":         "deployments",
	"deployment":     "deployments",
	"ds":             "daemonsets",
	"daemonset":      "daemonsets",
	"sts":            "statefulsets",
	"statefulset":    "statefulsets",
	"rs":             "replicasets",
	"replicaset":     "replicasets",
	"ns":             "namespaces",
	"namespace":      "namespaces",
	"no":             "nodes",
	"node":           "nodes",
	"cm":             "configmaps",
	"configmap":      "configmaps",
	"secret":         "secrets",
	"ing":            "ingresses",
	"ingress":        "ingresses",
	"pvc":            "persistentvolumeclaims",
	"pv":             "persistentvolumes",
	"sa":             "serviceaccounts",
	"serviceaccount": "serviceaccounts",
	"ep":             "endpoints",
	"endpoint":       "endpoints",
	"job":            "jobs",
	"cj":             "cronjobs",
	"cronjob":        "cronjobs",
	"hpa":            "horizontalpodautoscalers",
	"ev":             "events",
	"event":          "events",
	"crd":            "customresourcedefinitions",
	"ctx":            "_contexts",
	"context":        "_contexts",
	"contexts":       "_contexts",
}

// clusterScopedResources lists resources that are not namespace-scoped.
var clusterScopedResources = map[string]bool{
	"nodes":                     true,
	"namespaces":                true,
	"persistentvolumes":         true,
	"clusterroles":              true,
	"clusterrolebindings":       true,
	"storageclasses":            true,
	"ingressclasses":            true,
	"customresourcedefinitions": true,
	"priorityclasses":           true,
}

// ResolveResource returns the canonical resource name for a given alias.
func ResolveResource(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if canonical, ok := resourceAliases[name]; ok {
		return canonical
	}
	return name
}

// IsClusterScoped reports whether a resource is cluster-scoped (not namespaced).
func IsClusterScoped(resource string) bool {
	return clusterScopedResources[resource]
}
