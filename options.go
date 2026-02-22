package templr

import "html/template"

// Option is a function that configures the Renderer.
type Option func(*Renderer)

// WithFuncs adds a custom template FuncMap.
func WithFuncs(fm template.FuncMap) Option {
	return func(r *Renderer) {
		r.funcs = fm
	}
}

// WithAutoReload enables auto reloading of templates on every render.
// This is especially helpful during development.
func WithAutoReload() Option {
	return func(r *Renderer) {
		r.autoReload = true
	}
}
