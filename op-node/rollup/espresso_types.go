package rollup

// IsEspresso returns true if the Espresso upgrade is active at or past the
// given timestamp. EspressoTime is conceptually an L2-timestamp fork activation
// time, but the derivation pipeline calls this with the L1 origin time of the
// enclosing L1 block (mirroring upstream's ecotoneTime treatment), so the fork
// is effectively gated per L2 epoch. When active, the derivation pipeline runs
// all Espresso-specific semantics (event-based batch authentication via the
// BatchAuthenticator contract). When inactive, the pipeline behaves exactly as
// upstream Optimism.
func (c *Config) IsEspresso(timestamp uint64) bool {
	return c != nil && c.EspressoTime != nil && timestamp >= *c.EspressoTime
}
