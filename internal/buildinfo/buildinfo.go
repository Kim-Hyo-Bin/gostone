// Package buildinfo holds release metadata (overridden via -ldflags at link time).
package buildinfo

// Version is the release / image tag (e.g. 1.2.3 or dev).
var Version = "dev"

// Commit is a short VCS revision (e.g. git describe / rev-parse).
var Commit = "unknown"
