package resolver

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"io"
	"os"
)

type JsonResolver struct {
	filepath string
	lookup   map[string]string
}

func (j JsonResolver) ResolveHostname(hostname string) (string, bool) {
	val, ok := j.lookup[hostname]
	return val, ok
}

func (j JsonResolver) MarshalZerologObject(e *zerolog.Event) {
	e.Str("Filepath", j.filepath)
}

func NewJsonResolver(filepath string) (*JsonResolver, error) {
	jsonFile, err := os.Open(filepath)
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
		filepath: filepath,
		lookup:   mapping,
	}, nil
}
