package resource

import (
	"fmt"
	"strings"
	"time"
)

// Column defines one table column for a resource type.
type Column struct {
	Header  string
	extract func(obj map[string]any) string
}

// Value extracts the column value from a JSON object.
func (c Column) Value(obj map[string]any) string {
	return c.extract(obj)
}

// Field returns a Column that reads a dot-notation path from the JSON object.
func Field(header, path string) Column {
	parts := strings.Split(path, ".")
	return Column{Header: header, extract: func(obj map[string]any) string {
		return getPath(obj, parts...)
	}}
}

// Computed returns a Column with a custom extraction function.
func Computed(header string, fn func(map[string]any) string) Column {
	return Column{Header: header, extract: fn}
}

// ColumnsFor returns the default column set for a resource, and whether one is defined.
func ColumnsFor(resource string) ([]Column, bool) {
	cols, ok := defaults[resource]
	return cols, ok
}

// ─── Default column sets ──────────────────────────────────────────────────────

var defaults = map[string][]Column{
	"pods":                     podColumns,
	"services":                 serviceColumns,
	"deployments":              deploymentColumns,
	"daemonsets":               daemonsetColumns,
	"statefulsets":             statefulsetColumns,
	"replicasets":              replicasetColumns,
	"namespaces":               namespaceColumns,
	"nodes":                    nodeColumns,
	"configmaps":               configmapColumns,
	"secrets":                  secretColumns,
	"ingresses":                ingressColumns,
	"persistentvolumeclaims":   pvcColumns,
	"persistentvolumes":        pvColumns,
	"serviceaccounts":          saColumns,
	"jobs":                     jobColumns,
	"cronjobs":                 cronjobColumns,
	"horizontalpodautoscalers": hpaColumns,
	"events":                   eventColumns,
}

var podColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("READY", podReady),
	Computed("STATUS", podStatus),
	Computed("RESTARTS", podRestarts),
	Computed("AGE", ageField("metadata.creationTimestamp")),
	Field("IP", "status.podIP"),
	Field("NODE", "spec.nodeName"),
}

var serviceColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("TYPE", "spec.type"),
	Field("CLUSTER-IP", "spec.clusterIP"),
	Computed("EXTERNAL-IP", svcExternalIP),
	Computed("PORT(S)", svcPorts),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var deploymentColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("READY", func(o map[string]any) string {
		ready := getPath(o, "status", "readyReplicas")
		total := getPath(o, "spec", "replicas")
		if ready == "" {
			ready = "0"
		}
		if total == "" {
			total = "0"
		}
		return ready + "/" + total
	}),
	Field("UP-TO-DATE", "status.updatedReplicas"),
	Field("AVAILABLE", "status.availableReplicas"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var daemonsetColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("DESIRED", "status.desiredNumberScheduled"),
	Field("CURRENT", "status.currentNumberScheduled"),
	Field("READY", "status.numberReady"),
	Field("UP-TO-DATE", "status.updatedNumberScheduled"),
	Field("AVAILABLE", "status.numberAvailable"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var statefulsetColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("READY", func(o map[string]any) string {
		ready := getPath(o, "status", "readyReplicas")
		total := getPath(o, "spec", "replicas")
		if ready == "" {
			ready = "0"
		}
		if total == "" {
			total = "0"
		}
		return ready + "/" + total
	}),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var replicasetColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("DESIRED", "spec.replicas"),
	Field("CURRENT", "status.replicas"),
	Field("READY", "status.readyReplicas"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var namespaceColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("STATUS", "status.phase"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var nodeColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("STATUS", nodeStatus),
	Computed("ROLES", nodeRoles),
	Computed("AGE", ageField("metadata.creationTimestamp")),
	Field("VERSION", "status.nodeInfo.kubeletVersion"),
}

var configmapColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("DATA", func(o map[string]any) string {
		d, _ := o["data"].(map[string]any)
		return fmt.Sprintf("%d", len(d))
	}),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var secretColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("TYPE", "type"),
	Computed("DATA", func(o map[string]any) string {
		d, _ := o["data"].(map[string]any)
		return fmt.Sprintf("%d", len(d))
	}),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var ingressColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("CLASS", "spec.ingressClassName"),
	Computed("HOSTS", ingressHosts),
	Computed("ADDRESS", ingressAddress),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var pvcColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("STATUS", "status.phase"),
	Field("VOLUME", "spec.volumeName"),
	Field("CAPACITY", "status.capacity.storage"),
	Computed("ACCESS MODES", func(o map[string]any) string {
		return joinStringSlice(getSlice(o, "spec", "accessModes"), ",")
	}),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var pvColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("CAPACITY", "spec.capacity.storage"),
	Computed("ACCESS MODES", func(o map[string]any) string {
		return joinStringSlice(getSlice(o, "spec", "accessModes"), ",")
	}),
	Field("RECLAIM POLICY", "spec.persistentVolumeReclaimPolicy"),
	Field("STATUS", "status.phase"),
	Field("CLAIM", "spec.claimRef.name"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var saColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("SECRETS", func(o map[string]any) string {
		s, _ := o["secrets"].([]any)
		return fmt.Sprintf("%d", len(s))
	}),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var jobColumns = []Column{
	Field("NAME", "metadata.name"),
	Computed("COMPLETIONS", func(o map[string]any) string {
		succeeded := getPath(o, "status", "succeeded")
		total := getPath(o, "spec", "completions")
		if succeeded == "" {
			succeeded = "0"
		}
		if total == "" {
			total = "1"
		}
		return succeeded + "/" + total
	}),
	Computed("DURATION", jobDuration),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var cronjobColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("SCHEDULE", "spec.schedule"),
	Field("SUSPEND", "spec.suspend"),
	Computed("ACTIVE", func(o map[string]any) string {
		a, _ := o["status"].(map[string]any)
		if a == nil {
			return "0"
		}
		active, _ := a["active"].([]any)
		return fmt.Sprintf("%d", len(active))
	}),
	Field("LAST SCHEDULE", "status.lastScheduleTime"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var hpaColumns = []Column{
	Field("NAME", "metadata.name"),
	Field("REFERENCE", "spec.scaleTargetRef.name"),
	Field("MIN PODS", "spec.minReplicas"),
	Field("MAX PODS", "spec.maxReplicas"),
	Field("REPLICAS", "status.currentReplicas"),
	Computed("AGE", ageField("metadata.creationTimestamp")),
}

var eventColumns = []Column{
	Computed("LAST SEEN", ageField("lastTimestamp")),
	Field("TYPE", "type"),
	Field("REASON", "reason"),
	Field("OBJECT", "involvedObject.name"),
	Computed("MESSAGE", func(o map[string]any) string {
		msg := getPath(o, "message")
		if len(msg) > 80 {
			return msg[:77] + "..."
		}
		return msg
	}),
}

// ─── Computed helpers ─────────────────────────────────────────────────────────

func podReady(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "0/0"
	}
	statuses, _ := status["containerStatuses"].([]any)
	total := len(statuses)
	if total == 0 {
		// count initContainerStatuses if no containers yet
		inits, _ := status["initContainerStatuses"].([]any)
		total = len(inits)
	}
	ready := 0
	for _, cs := range statuses {
		c, _ := cs.(map[string]any)
		if r, _ := c["ready"].(bool); r {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

func podStatus(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "Unknown"
	}
	// Check waiting reason of first container first
	if css, _ := status["containerStatuses"].([]any); len(css) > 0 {
		if cs, _ := css[0].(map[string]any); cs != nil {
			if state, _ := cs["state"].(map[string]any); state != nil {
				if waiting, _ := state["waiting"].(map[string]any); waiting != nil {
					if reason := getPath(waiting, "reason"); reason != "" {
						return reason
					}
				}
				if terminated, _ := state["terminated"].(map[string]any); terminated != nil {
					if reason := getPath(terminated, "reason"); reason != "" {
						return reason
					}
				}
			}
		}
	}
	if phase := getPath(status, "phase"); phase != "" {
		return phase
	}
	return "Unknown"
}

func podRestarts(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "0"
	}
	total := 0
	if css, _ := status["containerStatuses"].([]any); len(css) > 0 {
		for _, cs := range css {
			c, _ := cs.(map[string]any)
			if n, _ := c["restartCount"].(float64); n > 0 {
				total += int(n)
			}
		}
	}
	return fmt.Sprintf("%d", total)
}

func nodeStatus(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "Unknown"
	}
	conditions, _ := status["conditions"].([]any)
	for _, cond := range conditions {
		c, _ := cond.(map[string]any)
		if getPath(c, "type") == "Ready" {
			if getPath(c, "status") == "True" {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func nodeRoles(o map[string]any) string {
	labels, _ := o["metadata"].(map[string]any)
	if labels == nil {
		return "<none>"
	}
	lmap, _ := labels["labels"].(map[string]any)
	var roles []string
	for k := range lmap {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
		}
	}
	if len(roles) == 0 {
		return "<none>"
	}
	return strings.Join(roles, ",")
}

func svcExternalIP(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "<none>"
	}
	lb, _ := status["loadBalancer"].(map[string]any)
	if lb == nil {
		return "<none>"
	}
	ingress, _ := lb["ingress"].([]any)
	if len(ingress) == 0 {
		return "<none>"
	}
	first, _ := ingress[0].(map[string]any)
	if ip := getPath(first, "ip"); ip != "" {
		return ip
	}
	if host := getPath(first, "hostname"); host != "" {
		return host
	}
	return "<none>"
}

func svcPorts(o map[string]any) string {
	spec, _ := o["spec"].(map[string]any)
	if spec == nil {
		return ""
	}
	ports, _ := spec["ports"].([]any)
	var parts []string
	for _, p := range ports {
		pm, _ := p.(map[string]any)
		port := fmt.Sprintf("%v", pm["port"])
		proto := fmt.Sprintf("%v", pm["protocol"])
		nodePort, hasNodePort := pm["nodePort"]
		if hasNodePort {
			parts = append(parts, fmt.Sprintf("%s:%v/%s", port, nodePort, proto))
		} else {
			parts = append(parts, fmt.Sprintf("%s/%s", port, proto))
		}
	}
	return strings.Join(parts, ",")
}

func ingressHosts(o map[string]any) string {
	spec, _ := o["spec"].(map[string]any)
	if spec == nil {
		return "<none>"
	}
	rules, _ := spec["rules"].([]any)
	var hosts []string
	for _, r := range rules {
		rm, _ := r.(map[string]any)
		if h := getPath(rm, "host"); h != "" {
			hosts = append(hosts, h)
		}
	}
	if len(hosts) == 0 {
		return "*"
	}
	return strings.Join(hosts, ",")
}

func ingressAddress(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return ""
	}
	lb, _ := status["loadBalancer"].(map[string]any)
	if lb == nil {
		return ""
	}
	ingress, _ := lb["ingress"].([]any)
	var addrs []string
	for _, i := range ingress {
		im, _ := i.(map[string]any)
		if ip := getPath(im, "ip"); ip != "" {
			addrs = append(addrs, ip)
		} else if h := getPath(im, "hostname"); h != "" {
			addrs = append(addrs, h)
		}
	}
	return strings.Join(addrs, ",")
}

func jobDuration(o map[string]any) string {
	status, _ := o["status"].(map[string]any)
	if status == nil {
		return "<unknown>"
	}
	start := getPath(status, "startTime")
	end := getPath(status, "completionTime")
	if start == "" {
		return "<unknown>"
	}
	t0, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return "<unknown>"
	}
	var t1 time.Time
	if end != "" {
		t1, err = time.Parse(time.RFC3339, end)
		if err != nil {
			t1 = time.Now()
		}
	} else {
		t1 = time.Now()
	}
	return formatAge(t1.Sub(t0))
}

// ageField returns a Computed column that formats the duration since the given dot-path timestamp.
func ageField(path string) func(map[string]any) string {
	parts := strings.Split(path, ".")
	return func(obj map[string]any) string {
		ts := getPath(obj, parts...)
		if ts == "" {
			return "<unknown>"
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
		return formatAge(time.Since(t))
	}
}

func formatAge(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd%dh", days, hours)
}

// ─── JSON field helpers ───────────────────────────────────────────────────────

// getPath traverses a nested map[string]any following the given keys.
func getPath(obj map[string]any, keys ...string) string {
	var cur any = obj
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[k]
	}
	if cur == nil {
		return ""
	}
	switch v := cur.(type) {
	case string:
		return v
	case bool:
		return fmt.Sprintf("%v", v)
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getSlice(obj map[string]any, keys ...string) []any {
	var cur any = obj
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[k]
	}
	s, _ := cur.([]any)
	return s
}

func joinStringSlice(items []any, sep string) string {
	var parts []string
	for _, i := range items {
		if s, ok := i.(string); ok {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, sep)
}
