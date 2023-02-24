package resolver

type Resolver interface {
	ResolveHostname(hostname string) (string, bool)
}
