// Package parse provides parsers for converting configuration formats to normalized trees.
package parse

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/pfrederiksen/configdiff/tree"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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
		return ParseHCL(data)
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

// ParseHCL parses HCL data into a normalized tree.
func ParseHCL(data []byte) (*tree.Node, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, "config.hcl")
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	// Extract attributes into a map
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to extract HCL attributes: %s", diags.Error())
	}

	result := make(map[string]interface{})
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate HCL attribute %q: %s", name, diags.Error())
		}

		goVal, err := ctyToGo(val)
		if err != nil {
			return nil, fmt.Errorf("failed to convert HCL value for %q: %w", name, err)
		}
		result[name] = goVal
	}

	node, err := valueToNode(result)
	if err != nil {
		return nil, err
	}

	// Set canonical paths
	node.SetPaths("/")
	return node, nil
}

// ctyToGo converts a cty.Value to a Go interface{} value
func ctyToGo(val cty.Value) (interface{}, error) {
	if val.IsNull() {
		return nil, nil
	}

	typ := val.Type()
	switch {
	case typ == cty.Bool:
		var result bool
		if err := gocty.FromCtyValue(val, &result); err != nil {
			return nil, err
		}
		return result, nil

	case typ == cty.Number:
		var result float64
		if err := gocty.FromCtyValue(val, &result); err != nil {
			return nil, err
		}
		return result, nil

	case typ == cty.String:
		var result string
		if err := gocty.FromCtyValue(val, &result); err != nil {
			return nil, err
		}
		return result, nil

	case typ.IsListType() || typ.IsTupleType():
		list := make([]interface{}, 0, val.LengthInt())
		it := val.ElementIterator()
		for it.Next() {
			_, elem := it.Element()
			goElem, err := ctyToGo(elem)
			if err != nil {
				return nil, err
			}
			list = append(list, goElem)
		}
		return list, nil

	case typ.IsMapType() || typ.IsObjectType():
		result := make(map[string]interface{})
		it := val.ElementIterator()
		for it.Next() {
			key, elem := it.Element()
			keyStr := key.AsString()
			goElem, err := ctyToGo(elem)
			if err != nil {
				return nil, err
			}
			result[keyStr] = goElem
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported cty type: %s", typ.FriendlyName())
	}
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
