package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client wraps kubectl invocations.
type Client struct {
	Context   string
	Namespace string
}

func (c *Client) baseArgs() []string {
	var args []string
	if c.Context != "" {
		args = append(args, "--context", c.Context)
	}
	return args
}

// List returns the header line and data rows from kubectl get.
func (c *Client) List(resource string) (header string, rows []string, err error) {
	if resource == "_contexts" {
		return c.listContexts()
	}

	args := c.baseArgs()
	args = append(args, "get", resource)

	if !IsClusterScoped(resource) {
		if c.Namespace == "all" {
			args = append(args, "--all-namespaces")
		} else {
			ns := c.Namespace
			if ns == "" {
				ns = "default"
			}
			args = append(args, "-n", ns)
		}
	}
	args = append(args, "-o", "wide")

	out, cmdErr := exec.Command("kubectl", args...).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if cmdErr != nil {
		return "", nil, fmt.Errorf("%s", output)
	}

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return "", nil, nil
	}
	return lines[0], lines[1:], nil
}

// ListJSON fetches resources as structured JSON objects.
// Returns nil items (no error) for pseudo-resources like _contexts.
func (c *Client) ListJSON(resource string) ([]map[string]any, error) {
	args := c.baseArgs()
	args = append(args, "get", resource, "-o", "json")

	if !IsClusterScoped(resource) {
		if c.Namespace == "all" {
			args = append(args, "--all-namespaces")
		} else {
			ns := c.Namespace
			if ns == "" {
				ns = "default"
			}
			args = append(args, "-n", ns)
		}
	}

	out, cmdErr := exec.Command("kubectl", args...).CombinedOutput()
	if cmdErr != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}

	var result struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("json parse: %w", err)
	}
	return result.Items, nil
}

func (c *Client) listContexts() (string, []string, error) {
	out, err := exec.Command("kubectl", "config", "get-contexts").CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return "", nil, nil
	}
	return lines[0], lines[1:], nil
}

// Describe runs kubectl describe for a resource.
func (c *Client) Describe(resource, namespace, name string) (string, error) {
	args := c.baseArgs()
	args = append(args, "describe", resource, name)
	args = append(args, c.nsArgs(namespace)...)
	return kubectlOutput(args...)
}

// YAML returns the resource manifest as YAML.
func (c *Client) YAML(resource, namespace, name string) (string, error) {
	args := c.baseArgs()
	args = append(args, "get", resource, name, "-o", "yaml")
	args = append(args, c.nsArgs(namespace)...)
	return kubectlOutput(args...)
}

// Logs returns pod logs (last 500 lines).
func (c *Client) Logs(namespace, name string) (string, error) {
	args := c.baseArgs()
	args = append(args, "logs", name, "--tail=500")
	args = append(args, c.nsArgs(namespace)...)
	return kubectlOutput(args...)
}

// Edit opens an interactive kubectl edit session, connecting stdin/stdout/stderr
// directly to the terminal. Must be called inside tview's Suspend callback.
func (c *Client) Edit(resource, namespace, name string) error {
	args := c.baseArgs()
	args = append(args, "edit", resource, name)
	args = append(args, c.nsArgs(namespace)...)
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Delete deletes a resource.
func (c *Client) Delete(resource, namespace, name string) (string, error) {
	args := c.baseArgs()
	args = append(args, "delete", resource, name)
	args = append(args, c.nsArgs(namespace)...)
	return kubectlOutput(args...)
}

func (c *Client) nsArgs(namespace string) []string {
	if namespace != "" {
		return []string{"-n", namespace}
	}
	if c.Namespace != "" && c.Namespace != "all" {
		return []string{"-n", c.Namespace}
	}
	return nil
}

// SwitchContext runs kubectl config use-context to make name the active context.
func (c *Client) SwitchContext(name string) error {
	out, err := exec.Command("kubectl", "config", "use-context", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
func (c *Client) CurrentContext() string {
	if c.Context != "" {
		return c.Context
	}
	args := c.baseArgs()
	args = append(args, "config", "current-context")
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// CurrentNamespace returns the namespace from kubeconfig if none was specified.
func (c *Client) CurrentNamespace() string {
	if c.Namespace != "" {
		return c.Namespace
	}
	args := c.baseArgs()
	args = append(args, "config", "view", "--minify", "-o", "jsonpath={..namespace}")
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "default"
	}
	return strings.TrimSpace(string(out))
}

func kubectlOutput(args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run()
	return buf.String(), nil
}
