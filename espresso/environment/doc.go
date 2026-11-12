// Package environment package is a collection of files that assist with the
// creation, management, and easy configuration of the Espresso chain in
// conjunction with the Optimism E2E (end-to-end) local testing devnet.
//
// This package contains a lot of helper functions, and utilities that allow
// for the easy setup, with sensible defaults, and the configuration of an
// `espresso-dev-node`. This process is predominately automatic, and isolated,
// with piecemeal parts of the Espresso Chain being launched in a Docker
// Container, with support for ports that are automatically mapped to external
// ports without explicit configuration.
package environment
