package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meesooqa/tgtag/internal/web"
)

// dummyDataProvider implements ControllerDataProvider for testing.
type dummyDataProvider struct{}

func (d *dummyDataProvider) GetApiData(r *http.Request) map[string]any {
	return map[string]any{"key": "apiValue"}
}

func (d *dummyDataProvider) GetTplData(r *http.Request) map[string]any {
	return map[string]any{"Title": "Test Title"}
}

// dummyTemplate implements web.Template for testing.
type dummyTemplate struct {
	dir string
}

func (d *dummyTemplate) GetTemplatesLocation() string {
	return d.dir
}

func (d *dummyTemplate) GetStaticLocation() string {
	return d.GetTemplatesLocation() + "/static"
}

func (d *dummyTemplate) GetLayoutTpl() string {
	return "layout.html"
}

func (d *dummyTemplate) GetDefaultContentTpl() string {
	return "content/default.html"
}

func (d *dummyTemplate) GetData(r *http.Request, contentData map[string]any) (map[string]any, error) {
	return nil, nil
}

// dummyChild is used to test that child controllers get their Router method called.
type dummyChild struct {
	routerCalled bool
}

func (d *dummyChild) Router(log *slog.Logger, mux *http.ServeMux, tpl web.Template) {
	d.routerCalled = true
}

func (d *dummyChild) GetChildren() []Controller {
	return nil
}

func (d *dummyChild) AddChildren(cc ...Controller) {}

func (d *dummyChild) GetRoute() string {
	return ""
}

func (d *dummyChild) GetTitle() string {
	return ""
}

// createTempTemplates creates a temporary directory with minimal template files.
// It creates "layout.html" at the top level and "content/default.html" in a subdirectory.
func createTempTemplates(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "templates")
	require.NoError(t, err)

	// Create layout.html at top level
	layoutPath := filepath.Join(tempDir, "layout.html")
	// The layout template outputs the "Title" value from the provided data.
	layoutContent := `{{with .}}{{index . "Title"}}{{end}}`
	err = os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	require.NoError(t, err)

	// Create content directory and default.html inside it.
	contentDir := filepath.Join(tempDir, "content")
	err = os.Mkdir(contentDir, 0755)
	require.NoError(t, err)

	contentPath := filepath.Join(contentDir, "default.html")
	// Content file can have arbitrary content since layout.html is used for rendering.
	contentContent := `default content`
	err = os.WriteFile(contentPath, []byte(contentContent), 0644)
	require.NoError(t, err)

	return tempDir
}

// TestAddAndGetChildren verifies that AddChildren and GetChildren work as expected.
func TestAddAndGetChildren(t *testing.T) {
	bc := &BaseController{}
	child1 := &BaseController{}
	child2 := &BaseController{}

	// Initially, Children should be empty.
	assert.Empty(t, bc.GetChildren(), "Expected no Children initially")

	// Add Children and check.
	bc.AddChildren(child1, child2)
	children := bc.GetChildren()
	assert.Len(t, children, 2, "Expected two Children")
	assert.Contains(t, children, child1, "Child1 should be in Children")
	assert.Contains(t, children, child2, "Child2 should be in Children")
}

// TestHandleApiWrongMethod verifies that handleApi returns 405 for an incorrect HTTP method.
func TestHandleApiWrongMethod(t *testing.T) {
	logger := slog.Default()
	bc := &BaseController{
		Self:      &dummyDataProvider{},
		MethodApi: "POST",
		RouteApi:  "/api",
		Log:       logger,
	}

	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()

	bc.handleApi(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "Expected status 405 for wrong method")
}

// TestHandleApiSuccess verifies that handleApi returns the correct JSON data.
func TestHandleApiSuccess(t *testing.T) {
	logger := slog.Default()
	bc := &BaseController{
		Self:      &dummyDataProvider{},
		MethodApi: "POST",
		RouteApi:  "/api",
		Log:       logger,
	}

	req := httptest.NewRequest("POST", "/api", nil)
	w := httptest.NewRecorder()

	bc.handleApi(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 for correct method")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Expected JSON content type")

	var data map[string]any
	err := json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err, "Decoding JSON should not error")
	assert.Equal(t, "apiValue", data["key"], "Expected API data to match")
}

// TestHandlePageWrongMethod verifies that handlePage returns 405 for an incorrect HTTP method.
func TestHandlePageWrongMethod(t *testing.T) {
	logger := slog.Default()
	bc := &BaseController{
		Self:   &dummyDataProvider{},
		Method: "GET",
		Route:  "/page",
		Log:    logger,
	}

	req := httptest.NewRequest("POST", "/page", nil)
	w := httptest.NewRecorder()

	bc.handlePage(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "Expected status 405 for wrong method in page handler")
}

// TestHandlePageSuccess verifies that handlePage renders the template correctly when using the proper HTTP method.
func TestHandlePageSuccess_001(t *testing.T) {
	logger := slog.Default()
	tempDir := createTempTemplates(t)
	defer os.RemoveAll(tempDir)

	// Create dummy template with temporary directory.
	dt := &dummyTemplate{dir: tempDir}

	bc := &BaseController{
		Self:   &dummyDataProvider{},
		Method: "GET",
		Route:  "/page",
		Log:    logger,
		// ContentTpl is empty so that initTemplates sets it to the default.
	}

	mux := http.NewServeMux()
	// Call Router which initializes templates and registers handlers.
	bc.Router(logger, mux, dt)

	// Create a GET request to the page route.
	req := httptest.NewRequest("GET", "/page", nil)
	w := httptest.NewRecorder()

	// Serve the request using the mux.
	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Reading response body should not error")

	// The layout template renders the "Title" from dummyDataProvider ("Test Title").
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 for correct method in page handler")
	assert.Contains(t, string(body), "Test Title", "Response body should contain the title from template data")
}

// TestRouterRegistersChildren verifies that Router calls the Router method of child controllers.
func TestRouterRegistersChildren(t *testing.T) {
	logger := slog.Default()
	tempDir := createTempTemplates(t)
	defer os.RemoveAll(tempDir)

	dt := &dummyTemplate{dir: tempDir}

	bc := &BaseController{
		Self:   &dummyDataProvider{},
		Method: "GET",
		Route:  "/parent",
		Log:    logger,
	}

	child := &dummyChild{}
	bc.AddChildren(child)

	mux := http.NewServeMux()
	bc.Router(logger, mux, dt)

	// The Router method should have called the child's Router.
	assert.True(t, child.routerCalled, "Expected child's Router to be called")
}
