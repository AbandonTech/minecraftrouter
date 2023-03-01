package resolver

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"log"
	"net/http"
)

type ApiResolver struct {
	apiUrl string
}

func (a ApiResolver) ResolveHostname(hostname string) (string, bool) {
	resp, err := http.Get(a.apiUrl)
	if err != nil {
		log.Fatal(err, "Unable to request from api")
	}

	defer resp.Body.Close()
	var data map[string]string

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatal(err, "Unable to decode response from api")
	}

	val, ok := data[hostname]
	return val, ok
}

func (a ApiResolver) MarshalZerologObject(e *zerolog.Event) {
	e.Str("ApiUrl", a.apiUrl)
}

func NewApiResolver(apiUrl string) ApiResolver {
	return ApiResolver{
		apiUrl: apiUrl,
	}
}
