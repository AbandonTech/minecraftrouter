package resolver

import "github.com/rs/zerolog"

type Resolver interface {
	zerolog.LogObjectMarshaler

	ResolveHostname(hostname string) (string, bool)
}
