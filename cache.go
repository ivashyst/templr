package templr

import (
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"strings"
)

const (
	pathPrefix = "templates/"
	pathSuffix = ".html"

	patternLayouts    = pathPrefix + "layouts/*" + pathSuffix
	patternPages      = pathPrefix + "pages/*" + pathSuffix
	patternErrors     = pathPrefix + "errors/*" + pathSuffix
	patternComponents = pathPrefix + "components/*" + pathSuffix
)

type templateCache map[string]*template.Template

func buildCache(fsys fs.FS, extraFuncs template.FuncMap) (templateCache, error) {
	result := make(templateCache)

	// 1. Initialize the root template with merged Funcs
	merged := make(template.FuncMap)
	for k, v := range funcMap {
		merged[k] = v
	}
	for k, v := range extraFuncs {
		merged[k] = v
	}
	root := template.New("root").Funcs(merged)

	// 2. Read and parse all the layout templates manually to control names
	layoutFiles, err := glob(fsys, patternLayouts, patternComponents)
	if err != nil {
		return nil, fmt.Errorf("list layouts: %w", err)
	}

	for _, path := range layoutFiles {
		name, err := templateNameFromPath(path)
		if err != nil {
			return nil, fmt.Errorf("extract template name from %s: %w", path, err)
		}

		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("read layout %s: %w", path, err)
		}

		// The layout is added to the 'root' template
		// therefore we do not need to store the instance
		_, err = root.New(name).Parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("parse layout %s: %w", path, err)
		}
	}

	// 3. Read and parse all the page templates manually to control names
	pageFiles, err := glob(fsys, patternPages, patternErrors)
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}

	for _, path := range pageFiles {
		name, err := templateNameFromPath(path)
		if err != nil {
			return nil, fmt.Errorf("extract template name from %s: %w", path, err)
		}

		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("read page %s: %w", path, err)
		}

		pageTmpl, err := root.Clone()
		if err != nil {
			return nil, fmt.Errorf("clone layouts: %w", err)
		}

		_, err = pageTmpl.New(name).Parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", path, err)
		}

		result[name] = pageTmpl
	}

	return result, nil
}

func glob(fsys fs.FS, patterns ...string) ([]string, error) {
	var result []string

	for _, pattern := range patterns {
		files, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, fmt.Errorf("list files in %s: %w", pattern, err)
		}

		result = append(result, files...)
	}

	return result, nil
}

// templateNameFromPath extracts template name from path
// by trimming the "templates/" prefix and "*.html" suffix
func templateNameFromPath(p string) (string, error) {
	p = path.Clean(p)

	if !strings.HasPrefix(p, pathPrefix) {
		return "", fmt.Errorf("invalid template path: %s", p)
	}

	if path.Ext(p) != pathSuffix {
		return "", fmt.Errorf("not an HTML template: %s", p)
	}

	return strings.TrimSuffix(
		strings.TrimPrefix(p, pathPrefix),
		pathSuffix,
	), nil
}
