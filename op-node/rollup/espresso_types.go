package rollup

// IsEspresso returns true if the Espresso upgrade is active at or past the
// given L2 block timestamp. When active, the derivation pipeline runs all
// Espresso-specific semantics (event-based batch authentication via the
// BatchAuthenticator contract). When inactive, the pipeline behaves exactly
// as upstream Optimism.
func (c *Config) IsEspresso(timestamp uint64) bool {
	return c.EspressoTime != nil && timestamp >= *c.EspressoTime
}
