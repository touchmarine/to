package template

import (
	"strings"
)

// MakeGlobalMapFunction returns a function that returns a map (via closure).
// A map is made if the given map is nil.
//
// MakeGlobalMapFunction is used to create global maps in templates.
func MakeGlobalMapFunction(base map[string]interface{}) func() map[string]interface{} {
	m := base
	if m == nil {
		m = map[string]interface{}{}
	}
	return func() map[string]interface{} {
		return m
	}
}

// Dot implements dot notation for accessing map values.
func Dot(m map[string]interface{}, selector string) interface{} {
	if i := strings.Index(selector, "."); i > -1 && i+1 < len(selector) {
		if mm, ok := m[selector[:i]].(map[string]interface{}); ok {
			return Dot(mm, selector[i+1:])
		}
	}
	return Get(m, selector)
}

// Get returns a map value or an empty string if entry doesn't exist.
func Get(m map[string]interface{}, key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}

// Set sets a map entry. A map is made if the given map is nil.
func Set(m map[string]interface{}, key string, v interface{}) map[string]interface{} {
	if m == nil {
		m = map[string]interface{}{}
	}
	m[key] = v
	return m
}

// SetDefault sets a map entry if the entry doesn't exist.
func SetDefault(m map[string]interface{}, key string, v interface{}) map[string]interface{} {
	if m != nil {
		if _, ok := m[key]; ok {
			// an entry already exists
			return m
		}
	}
	return Set(m, key, v)
}
