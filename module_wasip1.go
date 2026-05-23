//go:build wasip1

package client

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"unsafe"
)

//go:wasmimport looking-glass render
func hostRender(ptr unsafe.Pointer, length uint32)

// Module represents a running looking-glass module instance.
type Module struct {
	name string
}

// NewModule returns a new Module. The module name is read from the
// MODULE_NAME environment variable set by the host.
func NewModule() (*Module, error) {
	name := os.Getenv("MODULE_NAME")
	if name == "" {
		return nil, fmt.Errorf("MODULE_NAME not set")
	}

	return &Module{name: name}, nil
}

// Name returns the module name.
func (m *Module) Name() string {
	return m.name
}

// ParseConfig decodes the host-provided configuration into v.
// If no configuration was provided, v is left unchanged.
func (m *Module) ParseConfig(v any) error {
	cfgStr := os.Getenv("MODULE_CONFIG")
	if cfgStr == "" {
		return nil
	}

	return json.Unmarshal([]byte(cfgStr), v)
}

// Asset returns the contents of the named file from the module's assets directory.
func (m *Module) Asset(path string) ([]byte, error) {
	return os.ReadFile("/assets/" + path)
}

// Render sends a widget tree to the display. The host replaces the current
// content with the new tree on every call.
func (m *Module) Render(w Widget) {
	data, err := xml.Marshal(w)
	if err != nil {
		return
	}

	payload := make([]byte, len(m.name)+1+len(data))
	copy(payload, m.name)
	payload[len(m.name)] = 0
	copy(payload[len(m.name)+1:], data)

	hostRender(unsafe.Pointer(&payload[0]), uint32(len(payload)))
}
