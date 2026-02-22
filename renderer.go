package templr

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

// Renderer parses and caches html/template files from an fs.FS and renders
// them into HTTP responses. Use New or MustNew to create one.
type Renderer struct {
	cache      templateCache
	funcs      template.FuncMap
	autoReload bool
	fsys       fs.FS
}

// New creates a Renderer backed by fsys, building the template cache eagerly.
// Returns an error if any template fails to parse.
func New(fsys fs.FS, opts ...Option) (*Renderer, error) {
	r := &Renderer{fsys: fsys}

	for _, opt := range opts {
		opt(r)
	}

	// Initial build of the cache
	var err error
	r.cache, err = buildCache(fsys, r.funcs)
	if err != nil {
		return nil, fmt.Errorf("build template cache: %w", err)
	}

	return r, nil
}

// Must panics if err is non-nil, otherwise returns r. Intended for use with New
// at program startup: Must(New(fsys, opts...)).
func Must(r *Renderer, err error) *Renderer {
	if err != nil {
		panic(err)
	}

	return r
}

// MustNew is shorthand for Must(New(fsys, opts...)). Panics on parse errors.
func MustNew(fsys fs.FS, opts ...Option) *Renderer {
	return Must(New(fsys, opts...))
}

// Render executes the named template and writes the result to w with a
// text/html content type. templateName must match a path under templates/pages/
// or templates/errors/ (e.g. "pages/home", "errors/404").
// The response is buffered so a template execution error does not produce a
// partial 200 response.
func (t *Renderer) Render(w http.ResponseWriter, templateName string, data any) error {
	cache := t.cache
	if t.autoReload {
		var err error
		cache, err = buildCache(t.fsys, t.funcs)
		if err != nil {
			return fmt.Errorf("reload template cache: %w", err)
		}
	}

	templ, ok := cache[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Use a buffer to avoid "half-rendered" 200 OK responses
	buf := new(bytes.Buffer)
	err := templ.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		return fmt.Errorf("execute template %s: %w", templateName, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	return err
}
