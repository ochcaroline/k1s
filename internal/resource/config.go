package resource

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// UserColumn is a column entry in the user config file.
type UserColumn struct {
	Header string `yaml:"header"`
	Field  string `yaml:"field"` // dot-notation path, e.g. "metadata.name"
}

// UserConfig maps resource name → column list overrides.
type UserConfig map[string][]UserColumn

// ConfigPath returns the path for the user config file, respecting XDG_CONFIG_HOME.
func ConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "k1s", "columns.yaml")
}

// LoadConfig reads the user config from disk, creating a default template if absent.
func LoadConfig() (UserConfig, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if werr := writeDefaultConfig(path); werr != nil {
			// Non-fatal: couldn't write default, continue with built-in defaults.
			return nil, nil
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg UserConfig
	return cfg, yaml.Unmarshal(data, &cfg)
}

// ColumnsFor returns columns for a resource, checking user config first.
// User config entries replace the built-in defaults entirely for that resource.
func (cfg UserConfig) ColumnsFor(resource string) ([]Column, bool) {
	if cfg != nil {
		resource = strings.ToLower(strings.TrimSpace(resource))
		if uc, ok := cfg[resource]; ok {
			cols := make([]Column, len(uc))
			for i, c := range uc {
				cols[i] = Field(c.Header, c.Field)
			}
			return cols, true
		}
	}
	return ColumnsFor(resource)
}

// writeDefaultConfig creates the config directory and writes a documented template.
// The template is fully commented so built-in defaults remain active out of the box.
// Users uncomment and edit only the resources they want to customise.
func writeDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultConfigTemplate), 0o644)
}

const defaultConfigTemplate = `# k1s column configuration
# Location: $XDG_CONFIG_HOME/k1s/columns.yaml  (default: ~/.config/k1s/columns.yaml)
#
# Override which columns are displayed for any resource type.
# Each column needs:
#   header: display name shown in the table header
#   field:  dot-notation JSON path into the resource object (e.g. metadata.name)
#
# When a resource is listed here it fully replaces the built-in defaults for
# that resource. Built-in defaults (including computed columns such as READY,
# STATUS, RESTARTS, AGE) are used for any resource not listed below.
#
# Uncomment a block to activate it, then edit freely.
#
# ─── Examples ────────────────────────────────────────────────────────────────
#
# pods:
#   - header: NAME
#     field: metadata.name
#   - header: STATUS
#     field: status.phase
#   - header: IP
#     field: status.podIP
#   - header: NODE
#     field: spec.nodeName
#
# services:
#   - header: NAME
#     field: metadata.name
#   - header: TYPE
#     field: spec.type
#   - header: CLUSTER-IP
#     field: spec.clusterIP
#
# deployments:
#   - header: NAME
#     field: metadata.name
#   - header: REPLICAS
#     field: status.replicas
#   - header: AVAILABLE
#     field: status.availableReplicas
#
# nodes:
#   - header: NAME
#     field: metadata.name
#   - header: VERSION
#     field: status.nodeInfo.kubeletVersion
#   - header: OS
#     field: status.nodeInfo.osImage
`
