package assets

import "embed"

// Files contains the immutable document definitions bundled with the backend.
//
//go:embed templates/**
var Files embed.FS
