package util

import (
	"slices"

	"golang.org/x/exp/maps"
)

// Params are accessible in generator.Generate().
type Params map[string]any
type Option func(Params)

// Extract all parameters from the options passed as map[string]any.
func ToDict(opts ...Option) Params {
	params := Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

// Test if an option is present in params
func (p *Params) OptionExists(params Params, opt string) bool {
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

// Sugary generic function to get parameter from util.Params.
func Get[T any](params Params, key string) *T {
	if v, ok := params[key].(T); ok {
		return &v
	}
	if defaultValue, ok := params["default"].(T); ok {
		return &defaultValue
	}
	return nil
}

func GetOpt[T any](opts []Option, key string) *T {
	return Get[T](ToDict(opts...), "required_claims")
}

func (p Params) GetVerbose() bool {
	if verbose, ok := p["verbose"].(bool); ok {
		return verbose
	}

	// default setting
	return false
}
