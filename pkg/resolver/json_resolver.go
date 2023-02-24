package resolver

import (
	"encoding/json"
	"io"
	"os"
)

type JsonResolver struct {
	lookup map[string]string
}

func (r JsonResolver) ResolveHostname(hostname string) (string, bool) {
	val, ok := r.lookup[hostname]
	return val, ok
}

func NewJsonResolver(filename string) (*JsonResolver, error) {
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

	return &JsonResolver{
		lookup: mapping,
	}, nil
}
