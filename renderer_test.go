package templr

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// baseLayout is a minimal layout for tests.
const baseLayout = `
<!DOCTYPE html>
<html>
	<body>
		{{block "content" .}}[CONTENT]{{end}}
	</body>
</html>`

// homePage invokes the layout and defines the content block.
const homePage = `
{{template "layouts/base" .}}
{{define "content"}}Hello, World!{{end}}`

func makeFS(extra ...fstest.MapFS) fstest.MapFS {
	fsys := fstest.MapFS{
		"templates/layouts/base.html": {Data: []byte(baseLayout)},
		"templates/pages/home.html":   {Data: []byte(homePage)},
	}
	for _, m := range extra {
		for k, v := range m {
			fsys[k] = v
		}
	}
	return fsys
}

func TestBuildCache_PopulatesPageNames(t *testing.T) {
	cache, err := buildCache(makeFS(), nil)
	if err != nil {
		t.Fatalf("buildCache: %v", err)
	}

	if _, ok := cache["pages/home"]; !ok {
		t.Error("expected cache to contain 'pages/home'")
	}
}

func TestBuildCache_DoesNotExposeLayoutAsPage(t *testing.T) {
	cache, err := buildCache(makeFS(), nil)
	if err != nil {
		t.Fatalf("buildCache: %v", err)
	}

	if _, ok := cache["layouts/base"]; ok {
		t.Error("layouts should not be stored as standalone cache entries")
	}
}

func TestRender_WritesCorrectHTML(t *testing.T) {
	r, err := New(makeFS())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := r.Render(w, "pages/home", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Hello, World!") {
		t.Errorf("expected body to contain 'Hello, World!', got: %s", body)
	}
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("expected layout to be rendered, got: %s", body)
	}
}

func TestRender_UnknownTemplate(t *testing.T) {
	r, err := New(makeFS())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := r.Render(w, "pages/nonexistent", nil); err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
}

func TestDict_PassesDataToComponent(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/layouts/base.html": {Data: []byte(`
{{block "content" .}}{{end}}`)},
		"templates/components/card.html": {Data: []byte(`
<div>
	<h2>{{.Title}}</h2>
	<p>{{.Body}}</p>
</div>`)},
		"templates/pages/article.html": {Data: []byte(`
{{template "layouts/base" .}}
{{define "content"}}
{{template "components/card" (dict
		"Title" "My Title"
		"Body"  "My Body"
)}}
{{end}}`)},
	}

	r, err := New(fsys)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := r.Render(w, "pages/article", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "My Title") {
		t.Errorf("expected card title in body, got: %s", body)
	}
	if !strings.Contains(body, "My Body") {
		t.Errorf("expected card body in body, got: %s", body)
	}
}

func TestWithFuncs_CustomFuncAvailableInTemplate(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/layouts/base.html": {Data: []byte(`
{{block "content" .}}{{end}}`)},
		"templates/pages/greet.html": {Data: []byte(`
{{template "layouts/base" .}}
{{define "content"}}{{greet "World"}}{{end}}`)},
	}

	r, err := New(fsys, WithFuncs(template.FuncMap{
		"greet": func(name string) string { return "Hi, " + name + "!" },
	}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	w := httptest.NewRecorder()
	if err := r.Render(w, "pages/greet", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	if got := w.Body.String(); !strings.Contains(got, "Hi, World!") {
		t.Errorf("expected 'Hi, World!', got: %s", got)
	}
}

func TestWithAutoReload_PicksUpTemplateChanges(t *testing.T) {
	fsys := makeFS()

	r, err := New(fsys, WithAutoReload())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// First render — original content.
	w1 := httptest.NewRecorder()
	if err := r.Render(w1, "pages/home", nil); err != nil {
		t.Fatalf("first Render: %v", err)
	}
	if !strings.Contains(w1.Body.String(), "Hello, World!") {
		t.Fatalf("first render: expected 'Hello, World!', got: %s", w1.Body.String())
	}

	// Mutate the FS to simulate a file change.
	fsys["templates/pages/home.html"] = &fstest.MapFile{
		Data: []byte(`
{{template "layouts/base" .}}
{{define "content"}}Updated!{{end}}`),
	}

	// Second render — should pick up the new content.
	w2 := httptest.NewRecorder()
	if err := r.Render(w2, "pages/home", nil); err != nil {
		t.Fatalf("second Render: %v", err)
	}
	if !strings.Contains(w2.Body.String(), "Updated!") {
		t.Errorf("auto-reload: expected 'Updated!', got: %s", w2.Body.String())
	}
}
