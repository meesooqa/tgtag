package web

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewDefaultTemplate verifies that NewDefaultTemplate creates a template with the default code.
func TestNewDefaultTemplate(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	if tmpl == nil {
		t.Fatal("NewDefaultTemplate() returned nil")
	}
	// Check if GetTemplatesLocation returns the correct path
	expected := "templates/default"
	if got := tmpl.GetTemplatesLocation(); got != expected {
		t.Errorf("GetTemplatesLocation() = %v, want %v", got, expected)
	}
}

// TestGetTemplatesLocation verifies the GetTemplatesLocation method.
func TestGetTemplatesLocation(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	expected := "templates/default"
	if got := tmpl.GetTemplatesLocation(); got != expected {
		t.Errorf("GetTemplatesLocation() = %v, want %v", got, expected)
	}
}

// TestGetStaticLocation verifies the GetStaticLocation method.
func TestGetStaticLocation(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	expected := "templates/default/static"
	if got := tmpl.GetStaticLocation(); got != expected {
		t.Errorf("GetStaticLocation() = %v, want %v", got, expected)
	}
}

// TestGetLayoutTpl verifies the GetLayoutTpl method.
func TestGetLayoutTpl(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	expected := "layout.html"
	if got := tmpl.GetLayoutTpl(); got != expected {
		t.Errorf("GetLayoutTpl() = %v, want %v", got, expected)
	}
}

// TestGetDefaultContentTpl verifies the GetDefaultContentTpl method.
func TestGetDefaultContentTpl(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	expected := "content/default.html"
	if got := tmpl.GetDefaultContentTpl(); got != expected {
		t.Errorf("GetDefaultContentTpl() = %v, want %v", got, expected)
	}
}

// TestGetDefaultTitle verifies the getDefaultTitle unexported method.
func TestGetDefaultTitle(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	expected := "tgtag"
	if got := tmpl.getDefaultTitle(); got != expected {
		t.Errorf("getDefaultTitle() = %v, want %v", got, expected)
	}
}

// TestShallowMapMerge verifies the shallowMapMerge method.
func TestShallowMapMerge(t *testing.T) {
	t.Parallel()
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)

	tests := []struct {
		name     string
		map1     map[string]any
		map2     map[string]any
		expected map[string]any
	}{
		{
			name:     "no overlapping keys",
			map1:     map[string]any{"a": 1, "b": "test"},
			map2:     map[string]any{"c": true},
			expected: map[string]any{"a": 1, "b": "test", "c": true},
		},
		{
			name:     "overwriting existing keys",
			map1:     map[string]any{"a": 1, "b": "old"},
			map2:     map[string]any{"b": "new", "c": 42},
			expected: map[string]any{"a": 1, "b": "new", "c": 42},
		},
		{
			name:     "merge with empty map1",
			map1:     map[string]any{},
			map2:     map[string]any{"a": "hello", "b": 123},
			expected: map[string]any{"a": "hello", "b": 123},
		},
		{
			name:     "merge with empty map2",
			map1:     map[string]any{"a": 1},
			map2:     map[string]any{},
			expected: map[string]any{"a": 1},
		},
		{
			name:     "nil map2 (no panic)",
			map1:     map[string]any{"a": 1},
			map2:     nil,
			expected: map[string]any{"a": 1},
		},
		{
			name:     "shallow merge (no nested processing)",
			map1:     map[string]any{"nested": map[string]any{"a": 1}},
			map2:     map[string]any{"nested": "overwrite"},
			expected: map[string]any{"nested": "overwrite"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Prepare a clean copy to avoid test pollution
			map1Copy := cloneMap(tt.map1)

			tmpl.shallowMapMerge(map1Copy, tt.map2)
			assert.Equal(t, tt.expected, map1Copy)
		})
	}
}

func TestShallowMerge_NilMap1Panics(t *testing.T) {
	logger := slog.Default()
	tmpl := NewDefaultTemplate(logger, nil)
	assert.Panics(t, func() {
		tmpl.shallowMapMerge(nil, map[string]any{"a": 1})
	})
}

func cloneMap(m map[string]any) map[string]any {
	cp := make(map[string]any, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
