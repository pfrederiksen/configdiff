package parse

import (
	"os"
	"testing"

	"github.com/pfrederiksen/configdiff/tree"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *tree.Node)
	}{
		{
			name:    "null",
			input:   `null`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNull {
					t.Errorf("Kind = %v, want %v", n.Kind, tree.KindNull)
				}
			},
		},
		{
			name:    "boolean true",
			input:   `true`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindBool || n.Value != true {
					t.Errorf("Node = %v, want bool true", n.Value)
				}
			},
		},
		{
			name:    "boolean false",
			input:   `false`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindBool || n.Value != false {
					t.Errorf("Node = %v, want bool false", n.Value)
				}
			},
		},
		{
			name:    "integer",
			input:   `42`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber || n.Value != 42.0 {
					t.Errorf("Node = %v, want number 42", n.Value)
				}
			},
		},
		{
			name:    "float",
			input:   `3.14`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber || n.Value != 3.14 {
					t.Errorf("Node = %v, want number 3.14", n.Value)
				}
			},
		},
		{
			name:    "string",
			input:   `"hello"`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindString || n.Value != "hello" {
					t.Errorf("Node = %v, want string 'hello'", n.Value)
				}
			},
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject || len(n.Object) != 0 {
					t.Errorf("Kind = %v, len = %v, want empty object", n.Kind, len(n.Object))
				}
			},
		},
		{
			name:    "simple object",
			input:   `{"name": "test", "value": 123}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if len(n.Object) != 2 {
					t.Fatalf("Object len = %v, want 2", len(n.Object))
				}
				if n.Object["name"].Kind != tree.KindString || n.Object["name"].Value != "test" {
					t.Errorf("name = %v, want 'test'", n.Object["name"].Value)
				}
				if n.Object["value"].Kind != tree.KindNumber || n.Object["value"].Value != 123.0 {
					t.Errorf("value = %v, want 123", n.Object["value"].Value)
				}
			},
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindArray || len(n.Array) != 0 {
					t.Errorf("Kind = %v, len = %v, want empty array", n.Kind, len(n.Array))
				}
			},
		},
		{
			name:    "simple array",
			input:   `[1, 2, 3]`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindArray {
					t.Fatalf("Kind = %v, want array", n.Kind)
				}
				if len(n.Array) != 3 {
					t.Fatalf("Array len = %v, want 3", len(n.Array))
				}
				for i := 0; i < 3; i++ {
					expected := float64(i + 1)
					if n.Array[i].Value != expected {
						t.Errorf("Array[%d] = %v, want %v", i, n.Array[i].Value, expected)
					}
				}
			},
		},
		{
			name: "nested structure",
			input: `{
				"spec": {
					"replicas": 3,
					"containers": [
						{
							"name": "nginx",
							"image": "nginx:latest"
						}
					]
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				spec := n.Object["spec"]
				if spec == nil {
					t.Fatal("spec is nil")
				}
				if spec.Object["replicas"].Value != 3.0 {
					t.Errorf("replicas = %v, want 3", spec.Object["replicas"].Value)
				}
				containers := spec.Object["containers"]
				if containers == nil || containers.Kind != tree.KindArray {
					t.Fatal("containers is not an array")
				}
				if len(containers.Array) != 1 {
					t.Fatalf("containers len = %v, want 1", len(containers.Array))
				}
				container := containers.Array[0]
				if container.Object["name"].Value != "nginx" {
					t.Errorf("name = %v, want 'nginx'", container.Object["name"].Value)
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "paths are set",
			input:   `{"a": {"b": "c"}}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Path != "/" {
					t.Errorf("root path = %v, want /", n.Path)
				}
				if n.Object["a"].Path != "/a" {
					t.Errorf("a path = %v, want /a", n.Object["a"].Path)
				}
				if n.Object["a"].Object["b"].Path != "/a/b" {
					t.Errorf("b path = %v, want /a/b", n.Object["a"].Object["b"].Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("ParseJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseJSON() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, node)
			}
		})
	}
}

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *tree.Node)
	}{
		{
			name:    "null",
			input:   `null`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNull {
					t.Errorf("Kind = %v, want %v", n.Kind, tree.KindNull)
				}
			},
		},
		{
			name:    "boolean true",
			input:   `true`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindBool || n.Value != true {
					t.Errorf("Node = %v, want bool true", n.Value)
				}
			},
		},
		{
			name:    "integer",
			input:   `42`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber || n.Value != 42.0 {
					t.Errorf("Node = %v, want number 42", n.Value)
				}
			},
		},
		{
			name:    "string",
			input:   `hello`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindString || n.Value != "hello" {
					t.Errorf("Node = %v, want string 'hello'", n.Value)
				}
			},
		},
		{
			name: "simple object",
			input: `
name: test
value: 123`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if len(n.Object) != 2 {
					t.Fatalf("Object len = %v, want 2", len(n.Object))
				}
				if n.Object["name"].Kind != tree.KindString || n.Object["name"].Value != "test" {
					t.Errorf("name = %v, want 'test'", n.Object["name"].Value)
				}
				if n.Object["value"].Kind != tree.KindNumber || n.Object["value"].Value != 123.0 {
					t.Errorf("value = %v, want 123", n.Object["value"].Value)
				}
			},
		},
		{
			name: "simple array",
			input: `
- 1
- 2
- 3`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindArray {
					t.Fatalf("Kind = %v, want array", n.Kind)
				}
				if len(n.Array) != 3 {
					t.Fatalf("Array len = %v, want 3", len(n.Array))
				}
				for i := 0; i < 3; i++ {
					expected := float64(i + 1)
					if n.Array[i].Value != expected {
						t.Errorf("Array[%d] = %v, want %v", i, n.Array[i].Value, expected)
					}
				}
			},
		},
		{
			name: "nested structure",
			input: `
spec:
  replicas: 3
  containers:
    - name: nginx
      image: nginx:latest`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				spec := n.Object["spec"]
				if spec == nil {
					t.Fatal("spec is nil")
				}
				if spec.Object["replicas"].Value != 3.0 {
					t.Errorf("replicas = %v, want 3", spec.Object["replicas"].Value)
				}
				containers := spec.Object["containers"]
				if containers == nil || containers.Kind != tree.KindArray {
					t.Fatal("containers is not an array")
				}
				if len(containers.Array) != 1 {
					t.Fatalf("containers len = %v, want 1", len(containers.Array))
				}
				container := containers.Array[0]
				if container.Object["name"].Value != "nginx" {
					t.Errorf("name = %v, want 'nginx'", container.Object["name"].Value)
				}
			},
		},
		{
			name: "YAML-specific features",
			input: `
key: &anchor
  value: shared
copy: *anchor`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				// Both key and copy should have the same structure
				if n.Object["key"].Object["value"].Value != "shared" {
					t.Errorf("key.value = %v, want 'shared'", n.Object["key"].Object["value"].Value)
				}
				if n.Object["copy"].Object["value"].Value != "shared" {
					t.Errorf("copy.value = %v, want 'shared'", n.Object["copy"].Object["value"].Value)
				}
			},
		},
		{
			name:    "invalid YAML",
			input:   ":\ninvalid",
			wantErr: true,
		},
		{
			name: "paths are set",
			input: `
a:
  b: c`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Path != "/" {
					t.Errorf("root path = %v, want /", n.Path)
				}
				if n.Object["a"].Path != "/a" {
					t.Errorf("a path = %v, want /a", n.Object["a"].Path)
				}
				if n.Object["a"].Object["b"].Path != "/a/b" {
					t.Errorf("b path = %v, want /a/b", n.Object["a"].Object["b"].Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseYAML([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("ParseYAML() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseYAML() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, node)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		format  Format
		wantErr bool
	}{
		{
			name:    "JSON format",
			data:    `{"key": "value"}`,
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "YAML format",
			data:    "key: value",
			format:  FormatYAML,
			wantErr: false,
		},
		{
			name:    "HCL format",
			data:    `key = "value"`,
			format:  FormatHCL,
			wantErr: false,
		},
		{
			name:    "unsupported format",
			data:    "data",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.data), tt.format)
			if tt.wantErr {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Parse() error = %v", err)
				}
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Format
		wantErr bool
	}{
		{
			name:    "JSON object",
			data:    `{"key": "value"}`,
			want:    FormatJSON,
			wantErr: false,
		},
		{
			name:    "JSON array",
			data:    `[1, 2, 3]`,
			want:    FormatJSON,
			wantErr: false,
		},
		{
			name:    "YAML simple",
			data:    "key: value",
			want:    FormatYAML,
			wantErr: false,
		},
		{
			name: "YAML multi-line",
			data: `
key1: value1
key2: value2`,
			want:    FormatYAML,
			wantErr: false,
		},
		{
			name:    "invalid data",
			data:    ":\n:",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Error("DetectFormat() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("DetectFormat() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueToNode_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		check   func(*testing.T, *tree.Node)
	}{
		{
			name:    "int types",
			value:   int(42),
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber || n.Value != 42.0 {
					t.Errorf("int conversion failed: %v", n.Value)
				}
			},
		},
		{
			name:    "uint types",
			value:   uint(42),
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber || n.Value != 42.0 {
					t.Errorf("uint conversion failed: %v", n.Value)
				}
			},
		},
		{
			name:    "float32",
			value:   float32(3.14),
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindNumber {
					t.Errorf("float32 conversion failed: kind = %v", n.Kind)
				}
			},
		},
		{
			name:    "unsupported type",
			value:   struct{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := valueToNode(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Error("valueToNode() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("valueToNode() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, node)
			}
		})
	}
}

func TestNormalizeYAMLValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(*testing.T, interface{})
	}{
		{
			name:  "map[interface{}]interface{}",
			input: map[interface{}]interface{}{"key": "value"},
			check: func(t *testing.T, v interface{}) {
				m, ok := v.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T", v)
				}
				if m["key"] != "value" {
					t.Errorf("value = %v, want 'value'", m["key"])
				}
			},
		},
		{
			name:  "nested map",
			input: map[interface{}]interface{}{"outer": map[interface{}]interface{}{"inner": "value"}},
			check: func(t *testing.T, v interface{}) {
				m, ok := v.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T", v)
				}
				outer := m["outer"].(map[string]interface{})
				if outer["inner"] != "value" {
					t.Errorf("inner value = %v, want 'value'", outer["inner"])
				}
			},
		},
		{
			name:  "array with maps",
			input: []interface{}{map[interface{}]interface{}{"key": "value"}},
			check: func(t *testing.T, v interface{}) {
				arr, ok := v.([]interface{})
				if !ok {
					t.Fatalf("expected []interface{}, got %T", v)
				}
				m := arr[0].(map[string]interface{})
				if m["key"] != "value" {
					t.Errorf("value = %v, want 'value'", m["key"])
				}
			},
		},
		{
			name:  "passthrough scalar",
			input: "string",
			check: func(t *testing.T, v interface{}) {
				if v != "string" {
					t.Errorf("value = %v, want 'string'", v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeYAMLValue(tt.input)
			tt.check(t, result)
		})
	}
}
func TestParseHCL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *tree.Node)
	}{
		{
			name:    "boolean true",
			input:   `enabled = true`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if n.Object["enabled"].Kind != tree.KindBool || n.Object["enabled"].Value != true {
					t.Errorf("enabled = %v, want bool true", n.Object["enabled"].Value)
				}
			},
		},
		{
			name:    "boolean false",
			input:   `disabled = false`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if n.Object["disabled"].Kind != tree.KindBool || n.Object["disabled"].Value != false {
					t.Errorf("disabled = %v, want bool false", n.Object["disabled"].Value)
				}
			},
		},
		{
			name:    "integer",
			input:   `count = 42`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if n.Object["count"].Kind != tree.KindNumber || n.Object["count"].Value != 42.0 {
					t.Errorf("count = %v, want number 42", n.Object["count"].Value)
				}
			},
		},
		{
			name:    "float",
			input:   `ratio = 3.14`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if n.Object["ratio"].Kind != tree.KindNumber || n.Object["ratio"].Value != 3.14 {
					t.Errorf("ratio = %v, want number 3.14", n.Object["ratio"].Value)
				}
			},
		},
		{
			name:    "string",
			input:   `name = "test"`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if n.Object["name"].Kind != tree.KindString || n.Object["name"].Value != "test" {
					t.Errorf("name = %v, want string 'test'", n.Object["name"].Value)
				}
			},
		},
		{
			name: "object",
			input: `config = {
  host = "localhost"
  port = 8080
}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				configNode := n.Object["config"]
				if configNode.Kind != tree.KindObject {
					t.Fatalf("config.Kind = %v, want object", configNode.Kind)
				}
				if configNode.Object["host"].Kind != tree.KindString || configNode.Object["host"].Value != "localhost" {
					t.Errorf("config.host = %v, want 'localhost'", configNode.Object["host"].Value)
				}
				if configNode.Object["port"].Kind != tree.KindNumber || configNode.Object["port"].Value != 8080.0 {
					t.Errorf("config.port = %v, want 8080", configNode.Object["port"].Value)
				}
			},
		},
		{
			name:    "list of strings",
			input:   `tags = ["prod", "web"]`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				tagsNode := n.Object["tags"]
				if tagsNode.Kind != tree.KindArray {
					t.Fatalf("tags.Kind = %v, want array", tagsNode.Kind)
				}
				if len(tagsNode.Array) != 2 {
					t.Fatalf("tags len = %v, want 2", len(tagsNode.Array))
				}
				if tagsNode.Array[0].Kind != tree.KindString || tagsNode.Array[0].Value != "prod" {
					t.Errorf("tags[0] = %v, want 'prod'", tagsNode.Array[0].Value)
				}
				if tagsNode.Array[1].Kind != tree.KindString || tagsNode.Array[1].Value != "web" {
					t.Errorf("tags[1] = %v, want 'web'", tagsNode.Array[1].Value)
				}
			},
		},
		{
			name: "list of objects",
			input: `servers = [
  {
    name = "web1"
    ip   = "10.0.1.1"
  },
  {
    name = "web2"
    ip   = "10.0.1.2"
  }
]`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				serversNode := n.Object["servers"]
				if serversNode.Kind != tree.KindArray {
					t.Fatalf("servers.Kind = %v, want array", serversNode.Kind)
				}
				if len(serversNode.Array) != 2 {
					t.Fatalf("servers len = %v, want 2", len(serversNode.Array))
				}
				server1 := serversNode.Array[0]
				if server1.Kind != tree.KindObject {
					t.Fatalf("server1.Kind = %v, want object", server1.Kind)
				}
				if server1.Object["name"].Kind != tree.KindString || server1.Object["name"].Value != "web1" {
					t.Errorf("server1.name = %v, want 'web1'", server1.Object["name"].Value)
				}
				if server1.Object["ip"].Kind != tree.KindString || server1.Object["ip"].Value != "10.0.1.1" {
					t.Errorf("server1.ip = %v, want '10.0.1.1'", server1.Object["ip"].Value)
				}
			},
		},
		{
			name: "nested objects",
			input: `database = {
  connection = {
    host = "localhost"
    port = 5432
  }
  credentials = {
    username = "admin"
    password = "secret"
  }
}`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				dbNode := n.Object["database"]
				if dbNode.Kind != tree.KindObject {
					t.Fatalf("database.Kind = %v, want object", dbNode.Kind)
				}
				connNode := dbNode.Object["connection"]
				if connNode.Kind != tree.KindObject {
					t.Fatalf("connection.Kind = %v, want object", connNode.Kind)
				}
				if connNode.Object["host"].Kind != tree.KindString || connNode.Object["host"].Value != "localhost" {
					t.Errorf("connection.host = %v, want 'localhost'", connNode.Object["host"].Value)
				}
			},
		},
		{
			name: "multiple top-level attributes",
			input: `name = "app"
version = "1.0.0"
enabled = true`,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if len(n.Object) != 3 {
					t.Fatalf("Object len = %v, want 3", len(n.Object))
				}
				if n.Object["name"].Kind != tree.KindString || n.Object["name"].Value != "app" {
					t.Errorf("name = %v, want 'app'", n.Object["name"].Value)
				}
				if n.Object["version"].Kind != tree.KindString || n.Object["version"].Value != "1.0.0" {
					t.Errorf("version = %v, want '1.0.0'", n.Object["version"].Value)
				}
				if n.Object["enabled"].Kind != tree.KindBool || n.Object["enabled"].Value != true {
					t.Errorf("enabled = %v, want true", n.Object["enabled"].Value)
				}
			},
		},
		{
			name:    "invalid HCL",
			input:   `invalid = [[[`,
			wantErr: true,
		},
		{
			name:    "empty HCL",
			input:   ``,
			wantErr: false,
			check: func(t *testing.T, n *tree.Node) {
				if n.Kind != tree.KindObject {
					t.Fatalf("Kind = %v, want object", n.Kind)
				}
				if len(n.Object) != 0 {
					t.Errorf("Object len = %v, want 0", len(n.Object))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseHCL([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("ParseHCL() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseHCL() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, node)
			}
			// Verify paths are set
			if node.Path == "" {
				t.Error("ParseHCL() did not set paths")
			}
		})
	}
}

// Integration tests using testdata files
func TestParseHCL_Integration(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectedKeys []string
	}{
		{
			name:         "simple HCL file",
			file:         "../testdata/hcl/simple.hcl",
			expectedKeys: []string{"name", "enabled", "count", "ratio", "tags"},
		},
		{
			name:         "complex HCL file",
			file:         "../testdata/hcl/complex.hcl",
			expectedKeys: []string{"name", "version", "config", "servers", "metadata"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", tt.file, err)
			}

			node, err := ParseHCL(data)
			if err != nil {
				t.Fatalf("ParseHCL() error = %v", err)
			}

			if node.Kind != tree.KindObject {
				t.Fatalf("Kind = %v, want object", node.Kind)
			}

			for _, key := range tt.expectedKeys {
				if _, ok := node.Object[key]; !ok {
					t.Errorf("Expected key %q not found in parsed HCL", key)
				}
			}
		})
	}
}
