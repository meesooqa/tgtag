package extensions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meesooqa/tgtag/internal/web"
	"github.com/meesooqa/tgtag/pkg/controllers"
)

// dummyController is used to test that Router is called during route registration.
type dummyController struct {
	routerCalled bool
}

func (d *dummyController) Router(log *slog.Logger, mux *http.ServeMux, tpl web.Template) {
	d.routerCalled = true
}

func (d *dummyController) GetChildren() []controllers.Controller {
	return nil
}

func (d *dummyController) AddChildren(cc ...controllers.Controller) {}

func (d *dummyController) GetRoute() string {
	return ""
}

func (d *dummyController) GetTitle() string {
	return ""
}

// dummyTemplate is a test implementation for the web.Template interface.
type dummyTemplate struct {
	dir string
}

func (d *dummyTemplate) GetTemplatesLocation() string {
	return d.dir
}

func (d *dummyTemplate) GetStaticLocation() string {
	return d.dir + "/static"
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

// TestID verifies that ID returns a valid non-empty UUID.
func TestID(t *testing.T) {
	ext := &BaseExtension{}
	id := ext.ID()
	// Check that the ID is not empty.
	assert.NotEmpty(t, id, "ID should not be empty")
	// Check that the ID is a valid UUID.
	_, err := uuid.Parse(id)
	assert.NoError(t, err, "ID should be a valid UUID")
}

// TestRegisterRoutes verifies that RegisterRoutes calls Router on all controllers.
func TestRegisterRoutes(t *testing.T) {
	logger := slog.Default()
	mux := http.NewServeMux()
	dt := &dummyTemplate{dir: "dummy"}

	dummyCtrl1 := &dummyController{}
	dummyCtrl2 := &dummyController{}

	ext := &BaseExtension{
		Controllers: []controllers.Controller{dummyCtrl1, dummyCtrl2},
	}

	ext.RegisterRoutes(logger, mux, dt)

	// Verify that Router was called for each controller.
	assert.True(t, dummyCtrl1.routerCalled, "dummyCtrl1 Router should be called")
	assert.True(t, dummyCtrl2.routerCalled, "dummyCtrl2 Router should be called")
}

// TestRegisterRoutesEmpty verifies that RegisterRoutes does not panic when there are no controllers.
func TestRegisterRoutesEmpty(t *testing.T) {
	logger := slog.Default()
	mux := http.NewServeMux()
	dt := &dummyTemplate{dir: "dummy"}

	ext := &BaseExtension{
		Controllers: []controllers.Controller{},
	}

	// This should not panic.
	ext.RegisterRoutes(logger, mux, dt)
	// No further assertions are necessary.
}

// TestStaticHandler verifies that StaticHandler returns the correct path and serves static files.
func TestStaticHandler(t *testing.T) {
	// Create a temporary directory structure with "templates/default/static" and a test file.
	tempDir, err := os.MkdirTemp("", "static_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create the directory structure: tempDir/templates/default/static
	staticDir := filepath.Join(tempDir, "templates", "default", "static")
	err = os.MkdirAll(staticDir, 0755)
	require.NoError(t, err)

	// Create a test file inside the static directory.
	testFilePath := filepath.Join(staticDir, "test.txt")
	testContent := "Hello, world!"
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Save the current working directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)
	// Change working directory to tempDir so that "./templates/default/static" is resolved correctly.
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	// Ensure we change back to the original directory after the test.
	defer func() {
		_ = os.Chdir(origDir)
	}()

	ext := &BaseExtension{}
	prefix, handler := ext.StaticHandler()

	// Wrap the returned handler with http.StripPrefix to remove the URL prefix,
	// which is necessary for correct file resolution.
	wrappedHandler := http.StripPrefix(prefix, handler)

	// Create a test HTTP request to the static file.
	req := httptest.NewRequest("GET", prefix+"test.txt", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	// Verify that the file server returns the content of the test file.
	assert.Equal(t, testContent, string(body), "StaticHandler should serve the correct file content")
}
