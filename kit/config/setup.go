package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrKeyEmpty      = errors.New("key cannot be empty")
	ErrReadFile      = errors.New("failed to read config file")
	ErrMarshalStruct = errors.New("failed to marshal struct")
	ErrParseYAML     = errors.New("failed to parse YAML")
	ErrMapKeyType    = errors.New("map key must be string")
)

const (
	defaultSeparator = "."
)

type Options struct {
	Separator string
}

type Config struct {
	data map[string]interface{}
	opts Options
	mu   sync.RWMutex
}

func New(opts Options) *Config {
	if opts.Separator == "" {
		opts.Separator = defaultSeparator
	}
	return &Config{
		data: make(map[string]interface{}),
		opts: opts,
	}
}

func (c *Config) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReadFile, err)
	}

	m, err := unmarshalToMap(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	deepMerge(c.data, m)
	return nil
}

func (c *Config) LoadStruct(s interface{}) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalStruct, err)
	}

	m, err := unmarshalToMap(data)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	deepMerge(c.data, m)
	return nil
}

func deepMerge(dst, src map[string]interface{}) {
	for k, srcVal := range src {
		if dstVal, exists := dst[k]; exists {
			srcMap, srcIsMap := srcVal.(map[string]interface{})
			dstMap, dstIsMap := dstVal.(map[string]interface{})
			if srcIsMap && dstIsMap {
				deepMerge(dstMap, srcMap)
				continue
			}
		}
		dst[k] = srcVal
	}
}

func unmarshalToMap(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseYAML, err)
	}
	return m, nil
}
