package pkg

import (
	"encoding/json"
	"io"
	"os"
)

type Resolver struct {
	lookup map[string]string
}

func (r Resolver) ResolveHostname(hostname string) (string, bool) {
	val, ok := r.lookup[hostname]
	return val, ok
}

func NewResolver(filename string) (*Resolver, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, _ := io.ReadAll(jsonFile)

	var mapping map[string]string
	err = json.Unmarshal(bytes, &mapping)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		lookup: mapping,
	}, nil
}
