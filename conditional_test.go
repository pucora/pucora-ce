package pucora

import (
	"context"
	"testing"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/proxy"
)

type mockProxy struct {
	response *proxy.Response
	err      error
	called   bool
}

func (m *mockProxy) Call(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
	m.called = true
	return m.response, m.err
}

func TestBackendConditional_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cond     config.BackendConditional
		expected bool
	}{
		{
			name:     "header strategy is valid",
			cond:     config.BackendConditional{Strategy: "header"},
			expected: true,
		},
		{
			name:     "policy strategy is valid",
			cond:     config.BackendConditional{Strategy: "policy"},
			expected: true,
		},
		{
			name:     "fallback strategy is valid",
			cond:     config.BackendConditional{Strategy: "fallback"},
			expected: true,
		},
		{
			name:     "empty strategy is invalid",
			cond:     config.BackendConditional{Strategy: ""},
			expected: false,
		},
		{
			name:     "unknown strategy is invalid",
			cond:     config.BackendConditional{Strategy: "unknown"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.cond.IsValid()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestBackendConditional_IsFallback(t *testing.T) {
	tests := []struct {
		name     string
		cond     config.BackendConditional
		expected bool
	}{
		{
			name:     "fallback strategy",
			cond:     config.BackendConditional{Strategy: "fallback"},
			expected: true,
		},
		{
			name:     "header strategy",
			cond:     config.BackendConditional{Strategy: "header"},
			expected: false,
		},
		{
			name:     "policy strategy",
			cond:     config.BackendConditional{Strategy: "policy"},
			expected: false,
		},
		{
			name:     "empty strategy",
			cond:     config.BackendConditional{Strategy: ""},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.cond.IsFallback()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseBackendConditional(t *testing.T) {
	tests := []struct {
		name     string
		extra    map[string]interface{}
		expected *config.BackendConditional
	}{
		{
			name: "valid header strategy",
			extra: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "header",
					"name":     "X-Test",
					"value":    "A",
				},
			},
			expected: &config.BackendConditional{
				Strategy: "header",
				Name:     "X-Test",
				Value:    "A",
			},
		},
		{
			name: "valid policy strategy",
			extra: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "policy",
					"value":    "hasHeader('X-Test')",
				},
			},
			expected: &config.BackendConditional{
				Strategy: "policy",
				Value:    "hasHeader('X-Test')",
			},
		},
		{
			name: "valid fallback strategy",
			extra: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "fallback",
				},
			},
			expected: &config.BackendConditional{
				Strategy: "fallback",
			},
		},
		{
			name:     "no conditional config",
			extra:    map[string]interface{}{},
			expected: nil,
		},
		{
			name: "invalid strategy",
			extra: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "invalid",
				},
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseBackendConditional(tc.extra)
			if tc.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Errorf("expected %+v, got nil", tc.expected)
				return
			}
			if result.Strategy != tc.expected.Strategy {
				t.Errorf("expected Strategy=%s, got %s", tc.expected.Strategy, result.Strategy)
			}
			if result.Name != tc.expected.Name {
				t.Errorf("expected Name=%s, got %s", tc.expected.Name, result.Name)
			}
			if result.Value != tc.expected.Value {
				t.Errorf("expected Value=%s, got %s", tc.expected.Value, result.Value)
			}
		})
	}
}

func TestHasConditionalBackends(t *testing.T) {
	tests := []struct {
		name     string
		backends []*config.Backend
		expected bool
	}{
		{
			name:     "no backends",
			backends: []*config.Backend{},
			expected: false,
		},
		{
			name: "backend without conditional",
			backends: []*config.Backend{
				{ExtraConfig: map[string]interface{}{}},
			},
			expected: false,
		},
		{
			name: "backend with header conditional",
			backends: []*config.Backend{
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "header",
							"name":     "X-Test",
							"value":    "A",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "backend with policy conditional",
			backends: []*config.Backend{
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "policy",
							"value":    "hasHeader('X-Test')",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "backend with fallback conditional",
			backends: []*config.Backend{
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "fallback",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "one conditional, one regular",
			backends: []*config.Backend{
				{ExtraConfig: map[string]interface{}{}},
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "header",
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hasConditionalBackends(tc.backends)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestIsConditionalBackend(t *testing.T) {
	tests := []struct {
		name     string
		backend  *config.Backend
		expected bool
	}{
		{
			name:     "nil extra config",
			backend:  &config.Backend{ExtraConfig: nil},
			expected: false,
		},
		{
			name:     "empty extra config",
			backend:  &config.Backend{ExtraConfig: map[string]interface{}{}},
			expected: false,
		},
		{
			name: "header strategy",
			backend: &config.Backend{
				ExtraConfig: map[string]interface{}{
					"backend/conditional": map[string]interface{}{
						"strategy": "header",
					},
				},
			},
			expected: true,
		},
		{
			name: "policy strategy",
			backend: &config.Backend{
				ExtraConfig: map[string]interface{}{
					"backend/conditional": map[string]interface{}{
						"strategy": "policy",
					},
				},
			},
			expected: true,
		},
		{
			name: "fallback strategy",
			backend: &config.Backend{
				ExtraConfig: map[string]interface{}{
					"backend/conditional": map[string]interface{}{
						"strategy": "fallback",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid strategy",
			backend: &config.Backend{
				ExtraConfig: map[string]interface{}{
					"backend/conditional": map[string]interface{}{
						"strategy": "invalid",
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isConditionalBackend(tc.backend)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionalRouter_Execute(t *testing.T) {
	backends := []*config.Backend{
		{
			URLPattern: "/matched",
			ExtraConfig: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "header",
					"name":     "X-Test",
					"value":    "A",
				},
			},
		},
	}

	bf := func(b *config.Backend) proxy.Proxy {
		return func(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{Data: map[string]interface{}{"url": b.URLPattern}}, nil
		}
	}

	condProxy := newConditionalProxy(nil, &config.EndpointConfig{Backend: backends}, bf)

	req := &proxy.Request{
		Headers: map[string][]string{"X-Test": {"A"}},
	}

	resp, err := condProxy(context.Background(), req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Data["url"] != "/matched" {
		t.Errorf("expected url=/matched, got %v", resp.Data)
	}
}

func TestConditionalRouter_NoMatch(t *testing.T) {
	backends := []*config.Backend{
		{
			URLPattern: "/conditional",
			ExtraConfig: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "header",
					"name":     "X-Test",
					"value":    "A",
				},
			},
		},
	}

	bf := func(b *config.Backend) proxy.Proxy {
		return func(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{Data: map[string]interface{}{"url": b.URLPattern}}, nil
		}
	}

	condProxy := newConditionalProxy(nil, &config.EndpointConfig{Backend: backends}, bf)

	req := &proxy.Request{
		Headers: map[string][]string{"X-Test": {"B"}},
	}

	resp, err := condProxy(context.Background(), req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response when no condition matches, got %+v", resp)
	}
}

func TestConditionalRouter_Fallback(t *testing.T) {
	backends := []*config.Backend{
		{
			URLPattern: "/conditional",
			ExtraConfig: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "header",
					"name":     "X-Test",
					"value":    "A",
				},
			},
		},
		{
			URLPattern: "/fallback",
			ExtraConfig: map[string]interface{}{
				"backend/conditional": map[string]interface{}{
					"strategy": "fallback",
				},
			},
		},
	}

	bf := func(b *config.Backend) proxy.Proxy {
		return func(ctx context.Context, r *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{Data: map[string]interface{}{"url": b.URLPattern}}, nil
		}
	}

	condProxy := newConditionalProxy(nil, &config.EndpointConfig{Backend: backends}, bf)

	req := &proxy.Request{
		Headers: map[string][]string{"X-Test": {"B"}},
	}

	resp, err := condProxy(context.Background(), req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Data["url"] != "/fallback" {
		t.Errorf("expected url=/fallback, got %v", resp.Data)
	}
}

func TestGetRequiredHeaders(t *testing.T) {
	tests := []struct {
		name     string
		backends []*config.Backend
		expected []string
	}{
		{
			name:     "no backends",
			backends: []*config.Backend{},
			expected: nil,
		},
		{
			name: "header conditions with different headers",
			backends: []*config.Backend{
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "header",
							"name":     "X-Header-A",
						},
					},
				},
				{
					ExtraConfig: map[string]interface{}{
						"backend/conditional": map[string]interface{}{
							"strategy": "header",
							"name":     "X-Header-B",
						},
					},
				},
			},
			expected: []string{"x-header-a", "x-header-b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getRequiredHeaders(tc.backends)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d headers, got %d", len(tc.expected), len(result))
				return
			}
			for i, h := range result {
				if h != tc.expected[i] {
					t.Errorf("expected header[%d]=%s, got %s", i, tc.expected[i], h)
				}
			}
		})
	}
}