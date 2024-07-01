package util

import (
	"slices"

	"golang.org/x/exp/maps"
)

type Params = map[string]any
type Option func(Params)

func GetParams(opts ...Option) Params {
	params := Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

func OptionExists(params Params, opt string) bool {
	var k []string = maps.Keys(params)
	return slices.Contains(k, opt)
}

// Assert that the options exists within the params map
func AssertOptionsExist(params Params, opts ...string) []string {
	foundKeys := []string{}
	for k := range params {
		index := slices.IndexFunc(opts, func(s string) bool {
			return s == k
		})
		if index >= 0 {
			foundKeys = append(foundKeys, k)
		}
	}
	return foundKeys
}

func WithDefault[T any](v T) Option {
	return func(p Params) {
		p["default"] = v
	}
}

// Syntactic sugar generic function to get parameter from util.Params.
func Get[T any](params Params, key string, opts ...Option) *T {
	if v, ok := params[key].(T); ok {
		return &v
	}
	if defaultValue, ok := params["default"].(T); ok {
		return &defaultValue
	}
	return nil
}
