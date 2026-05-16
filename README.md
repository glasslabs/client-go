![Logo](https://svg.wiersma.co.za/glasslabs/module?title=CLIENT-GO&tag=a%20WASM%20client%20library)

[![GitHub release](https://img.shields.io/github/release/glasslabs/client-go.svg)](https://github.com/glasslabs/client-go/releases)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/glasslabs/client-go/main/LICENSE)

`client-go` is the Go library for writing [looking-glass](https://github.com/glasslabs/looking-glass) modules. Modules are compiled to WebAssembly (`GOARCH=wasm GOOS=wasip1`) and run inside the looking-glass host. This library provides everything a module needs: identity, configuration, asset access, rendering, HTTP, and logging.

**Note:** This is in active development, the API may change.

---

## How a module works

A looking-glass module is a standard Go `main` package compiled to a WASI binary. The host launches one instance per module slot, injects configuration via environment variables, mounts an assets directory, and exposes a small set of host functions for rendering and HTTP. The module runs its own event loop and calls `mod.Content(widget)` whenever the display should update.

## Writing a module

### 1. Initialise the module

`client.NewModule()` reads the `MODULE_NAME` environment variable set by the host. It must be called before any other client functions that require the module identity.

```go
//go:build wasip1

package main

import "github.com/glasslabs/client-go"

func main() {
    log := client.NewLogger()

    mod, err := client.NewModule()
    if err != nil {
        log.Error("Could not create module", "error", err.Error())
        return
    }
    // ...
}
```

### 2. Parse configuration

The host encodes the module's configuration map as JSON and exposes it via `MODULE_CONFIG`. Call `mod.ParseConfig` with a pointer to your config struct; fields are populated from the JSON keys that match the struct tags. Set defaults before calling `ParseConfig` so that omitted keys retain sensible values.

```go
type Config struct {
    Interval int    `json:"interval"`
    Label    string `json:"label"`
}

cfg := Config{Interval: 60} // set defaults first
if err := mod.ParseConfig(&cfg); err != nil {
    log.Error("Could not parse config", "error", err.Error())
    return
}
```

### 3. Load assets

Static files (templates, images, SVG) are placed alongside the module binary. The host mounts them at `/assets` inside the WASM sandbox. Use `mod.Asset(path)` to read them; `path` is relative to that directory.

```go
data, err := mod.Asset("index.html")
```

### 4. Render content

Call `mod.Content(widget)` to push a new widget tree to the display. The host replaces the module's current content with the new tree on every call — there is no diffing. Call it once on startup and again whenever the display needs to change.

Widget trees are built from the composable types in this package:

| Widget | Behaviour |
|--------|-----------|
| `NewText(s, ...opts)` | A single styled line of text. |
| `NewSVG(markup)` | Raw SVG markup rasterised into the slot. |
| `NewVStack(children...)` | Lays out children vertically, top to bottom. |
| `NewHStack(children...)` | Lays out children horizontally, left to right. |
| `NewSpacer()` | Flexible empty space that fills remaining room in a stack. |
| `NewCanvas(w, h, ops...)` | A fixed logical viewport scaled to fit, drawn with `DrawOp` commands. |

`Text` can be styled with option functions: `WithColor`, `WithFontSize`, `WithLight`, `WithBold`, `WithItalic`, `WithCondensed`, and `WithAlign`.

```go
func render(mod *client.Module, label string) {
    mod.Content(client.NewVStack(
        client.NewText(label,
            client.WithColor("#ffffff"),
            client.WithFontSize(48),
            client.WithLight(),
        ),
        client.NewSpacer(),
    ))
}
```

#### Canvas drawing

`Canvas` renders a list of `DrawOp` commands within a logical coordinate space. The viewport is scaled uniformly to fit the allocated slot.

| DrawOp | Behaviour |
|--------|-----------|
| `NewRect(x, y, w, h, ...opts)` | Filled and/or stroked rectangle. Options: `WithFill`, `WithStroke`, `WithCornerRadius`. |
| `NewArc(cx, cy, radius, startAngle, sweepAngle, strokeWidth, color)` | Circular arc stroke. Angles are in degrees; 0° is right, clockwise positive. |
| `NewLabel(x, y, align, runs...)` | Baseline-aligned multi-run text. `align` is `"start"`, `"middle"`, or `"end"`. |
| `NewPath(x, y, scale, d, fill)` | SVG path `d` string placed at an offset with optional uniform scale. |

`Label` is built from `TextRun` segments created with `NewRun(content, ...opts)`. Options: `WithRunFontSize`, `WithRunBaselineShift`, `WithRunColor`.

### 5. Make HTTP requests

The host provides an HTTP transport that is automatically installed into `http.DefaultClient` when the module starts. Use the standard `net/http` package directly — no special client is required.

```go
resp, err := http.Get("https://example.com/api/data")
```

The transport supports streaming response bodies, so Server-Sent Events and other long-lived streams work as expected.

### 6. Log messages

`client.NewLogger()` returns a logger that writes logfmt lines to stderr. The host captures these lines and forwards them to its structured logging pipeline.

```go
log := client.NewLogger()
log.Info("Module ready", "module", mod.Name())
log.Error("Something failed", "error", err.Error())
```

Methods: `Debug`, `Info`, `Warn`, `Error`. Each accepts a message followed by alternating key/value string pairs.

---

## Building a module

Modules must be compiled for the `wasip1` target:

```bash
GOARCH=wasm GOOS=wasip1 go build -o my-module.wasm .
```

All source files that use `client-go` should carry the `//go:build wasip1` build constraint, as the host functions are only available inside the WASM sandbox.

## Minimal example

```go
//go:build wasip1

package main

import (
    "time"

    "github.com/glasslabs/client-go"
)

type Config struct {
    Format string `json:"format"`
}

func main() {
    log := client.NewLogger()

    mod, err := client.NewModule()
    if err != nil {
        log.Error("Could not create module", "error", err.Error())
        return
    }

    cfg := Config{Format: "15:04:05"}
    if err = mod.ParseConfig(&cfg); err != nil {
        log.Error("Could not parse config", "error", err.Error())
        return
    }

    log.Info("Module ready", "module", mod.Name())

    for {
        mod.Content(client.NewText(
            time.Now().Format(cfg.Format),
            client.WithColor("#ffffff"),
            client.WithFontSize(48),
        ))
        time.Sleep(time.Second)
    }
}
```

