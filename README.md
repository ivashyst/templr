# templr

[![pipeline status](https://github.com/ivashyst/templr/actions/workflows/ci.yml/badge.svg)](https://github.com/ivashyst/templr/actions/workflows/ci.yml)

A lightweight `html/template` renderer for Go with layout inheritance, template caching, and zero external dependencies.

## What it is

`templr` wraps the standard `html/template` package to add layout inheritance and a build-once template cache. You organize templates into layouts, pages, components, and error pages using a simple directory convention. Each page template renders inside a layout by invoking it via `{{template "layouts/base" .}}`, and the renderer handles cloning, caching, and buffered output automatically.

## Template layout conventions

Templates live under a `templates/` directory (embedded via `embed.FS` or any `fs.FS`):

```
templates/
├── layouts/
│   └── base.html        # Named "layouts/base"
├── components/
│   └── nav.html         # Named "components/nav"
├── pages/
│   └── home.html        # Named "pages/home"
└── errors/
    └── 404.html         # Named "errors/404"
```

- **`layouts/`** and **`components/`** — shared templates loaded into every page's clone
- **`pages/`** and **`errors/`** — standalone pages stored in the cache, each rendered individually

Template names in Go code use the slash-separated path without the `templates/` prefix or `.html` suffix (e.g. `"pages/home"`, `"errors/404"`).

## Quick start

```bash
go get github.com/ivashyst/templr
```

```go
import (
    "embed"
    "net/http"

    "github.com/ivashyst/templr"
)

//go:embed templates
var templateFS embed.FS

func main() {
    r := templr.MustNew(templateFS)

    http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        if err := r.Render(w, "pages/home", nil); err != nil {
            http.Error(w, err.Error(), 500)
        }
    })

    http.ListenAndServe(":8080", nil)
}
```

## Layout inheritance

Define the layout in `templates/layouts/base.html`:

```html
<!DOCTYPE html>
<html>
  <head><title>My App</title></head>
  <body>
    {{block "content" .}}{{end}}
  </body>
</html>
```

Invoke it from a page in `templates/pages/home.html`:

```html
{{template "layouts/base" .}}
{{define "content"}}
  <h1>Hello, World!</h1>
{{end}}
```

When `r.Render(w, "pages/home", data)` is called, the page template is executed, which triggers the layout, which renders the `content` block defined by the page.

## Options

### `WithFuncs` — custom template functions

```go
r, err := templr.New(fsys, templr.WithFuncs(template.FuncMap{
    "upper": strings.ToUpper,
    "greet": func(name string) string { return "Hello, " + name + "!" },
}))
```

Custom functions are merged with the built-in ones and available in all templates.

### `WithAutoReload` — development mode

```go
r, err := templr.New(fsys, templr.WithAutoReload())
```

When enabled, the template cache is rebuilt on every `Render` call instead of once at startup. Useful during development so template changes are picked up without restarting the server. Disable in production.

## Built-in template functions

| Function | Signature | Output |
|---|---|---|
| `formatDate` | `formatDate time.Time` | `2006-01-02` |
| `formatDateTime` | `formatDateTime time.Time` | `2006-01-02 15:04` |
| `formatTime` | `formatTime time.Time` | `15:04` |
| `dict` | `dict "key" value ...` | `map[string]any` for passing named data to sub-templates |

Example usage in a template:

```html
<time>{{formatDate .CreatedAt}}</time>

{{template "components/card" (dict
        "Title" .Title
        "Body"  .Body
)}}
```
