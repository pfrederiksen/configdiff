// Package parse provides parsers for converting configuration formats to normalized trees.
package parse

import (
	"encoding/json"
	"fmt"

	"github.com/pfrederiksen/configdiff/tree"
	"gopkg.in/yaml.v3"
)

// Format represents a supported configuration format.
type Format string

const (
	// FormatYAML represents YAML format.
	FormatYAML Format = "yaml"

	// FormatJSON represents JSON format.
	FormatJSON Format = "json"

	// FormatHCL represents HCL format (experimental).
	FormatHCL Format = "hcl"
)

// Parse parses configuration data in the specified format into a normalized tree.
func Parse(data []byte, format Format) (*tree.Node, error) {
	switch format {
	case FormatYAML:
		return ParseYAML(data)
	case FormatJSON:
		return ParseJSON(data)
	case FormatHCL:
		return nil, fmt.Errorf("HCL format not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ParseYAML parses YAML data into a normalized tree.
func ParseYAML(data []byte) (*tree.Node, error) {
	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// YAML unmarshals into map[interface{}]interface{}, need to normalize
	normalized := normalizeYAMLValue(v)
	node, err := valueToNode(normalized)
	if err != nil {
		return nil, err
	}

	// Set canonical paths
	node.SetPaths("/")
	return node, nil
}

// ParseJSON parses JSON data into a normalized tree.
func ParseJSON(data []byte) (*tree.Node, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	node, err := valueToNode(v)
	if err != nil {
		return nil, err
	}

	// Set canonical paths
	node.SetPaths("/")
	return node, nil
}

// normalizeYAMLValue converts YAML's map[interface{}]interface{} to map[string]interface{}
// for consistent handling with JSON.
func normalizeYAMLValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{})
		for k, v := range val {
			keyStr := fmt.Sprintf("%v", k)
			normalized[keyStr] = normalizeYAMLValue(v)
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, len(val))
		for i, item := range val {
			normalized[i] = normalizeYAMLValue(item)
		}
		return normalized
	default:
		return val
	}
}

// valueToNode converts a Go value to a tree.Node.
func valueToNode(v interface{}) (*tree.Node, error) {
	if v == nil {
		return tree.NewNull(), nil
	}

	switch val := v.(type) {
	case bool:
		return tree.NewBool(val), nil

	case int:
		return tree.NewNumber(float64(val)), nil
	case int8:
		return tree.NewNumber(float64(val)), nil
	case int16:
		return tree.NewNumber(float64(val)), nil
	case int32:
		return tree.NewNumber(float64(val)), nil
	case int64:
		return tree.NewNumber(float64(val)), nil
	case uint:
		return tree.NewNumber(float64(val)), nil
	case uint8:
		return tree.NewNumber(float64(val)), nil
	case uint16:
		return tree.NewNumber(float64(val)), nil
	case uint32:
		return tree.NewNumber(float64(val)), nil
	case uint64:
		return tree.NewNumber(float64(val)), nil
	case float32:
		return tree.NewNumber(float64(val)), nil
	case float64:
		return tree.NewNumber(val), nil

	case string:
		return tree.NewString(val), nil

	case map[string]interface{}:
		obj := make(map[string]*tree.Node)
		for k, v := range val {
			node, err := valueToNode(v)
			if err != nil {
				return nil, err
			}
			obj[k] = node
		}
		return tree.NewObject(obj), nil

	case []interface{}:
		arr := make([]*tree.Node, len(val))
		for i, item := range val {
			node, err := valueToNode(item)
			if err != nil {
				return nil, err
			}
			arr[i] = node
		}
		return tree.NewArray(arr), nil

	default:
		return nil, fmt.Errorf("unsupported value type: %T", v)
	}
}

// DetectFormat attempts to detect the format based on content.
// Returns the detected format or an error if detection fails.
func DetectFormat(data []byte) (Format, error) {
	// Try JSON first (stricter format)
	var jsonVal interface{}
	if err := json.Unmarshal(data, &jsonVal); err == nil {
		return FormatJSON, nil
	}

	// Try YAML
	var yamlVal interface{}
	if err := yaml.Unmarshal(data, &yamlVal); err == nil {
		return FormatYAML, nil
	}

	return "", fmt.Errorf("unable to detect format")
}
