package rollup

// IsEspressoEnforcement returns true if the Espresso enforcement upgrade is
// active at or past the given L2 block timestamp. When active, the derivation
// pipeline runs all Espresso-specific semantics (event-based batch
// authentication via the BatchAuthenticator contract). When inactive, the
// pipeline behaves exactly as upstream Optimism.
func (c *Config) IsEspressoEnforcement(timestamp uint64) bool {
	return c.EspressoEnforcementTime != nil && timestamp >= *c.EspressoEnforcementTime
}
