//go:build js && wasm

package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
	"honnef.co/go/js/dom/v2"
)

// Module bridges the gap between WASM modules and looking glass.
type Module struct {
	name string
	root dom.Element
}

// NewModule returns a module.
func NewModule() (*Module, error) {
	name := os.Args[0]

	root := dom.GetWindow().Document().QuerySelector("#" + name + ".module")
	if root == nil {
		return nil, fmt.Errorf("module %q not found", name)
	}

	return &Module{
		name: name,
		root: root,
	}, nil
}

// Name returns the module name.
func (m *Module) Name() string {
	return m.name
}

// ParseConfig parse the config for the module into v.
func (m *Module) ParseConfig(v any) error {
	if len(os.Args) <= 1 {
		return nil
	}

	return yaml.Unmarshal([]byte(os.Args[1]), v)
}

// Asset returns the path in the configured asset directory.
func (m *Module) Asset(path string) ([]byte, error) {
	assetPath := os.Getenv("ASSETS_URL")
	u, err := url.Parse(os.Getenv("ASSETS_URL"))
	if err != nil {
		return nil, fmt.Errorf("parsing assest path %q: %w", assetPath, err)
	}
	u = u.JoinPath(path)

	//nolint:noctx // There is not context here anyway.
	resp, err := http.DefaultClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("getting asset: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// LoadCSS loads the given styles into the module.
func (m *Module) LoadCSS(styles ...string) error {
	for _, style := range styles {
		styleElem := dom.GetWindow().Document().CreateElement("style")
		styleElem.SetID(m.name)
		styleElem.SetTextContent(style)

		headElem := dom.GetWindow().Document().QuerySelector("head")
		if headElem == nil {
			return errors.New("head element not found")
		}
		headElem.AppendChild(styleElem)
	}
	return nil
}

// Element returns the modules root DOM element.
func (m *Module) Element() dom.Element {
	return m.root
}
