//go:build !amd64 && !arm64

package falsesharing

// cacheLineSize is the unit the cache actually moves around, in bytes.
const cacheLineSize = 64
