package template

import (
	"strings"
)

func MakeGlobalMapFunction(global map[string]interface{}) func() map[string]interface{} {
	globalMap := map[string]interface{}{}
	if global != nil {
		globalMap = global
	}
	return func() map[string]interface{} {
		return globalMap
	}
}

func Dot(m map[string]interface{}, selector string) interface{} {
	if i := strings.Index(selector, "."); i > -1 && i+1 < len(selector) {
		if mm, ok := m[selector[:i]].(map[string]interface{}); ok {
			return Dot(mm, selector[i+1:])
		}
	}
	return Get(m, selector)
}

func Get(m map[string]interface{}, key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}
