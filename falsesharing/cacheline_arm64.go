//go:build arm64

package falsesharing

// cacheLineSize is the unit the cache actually moves around, in bytes.
// Apple Silicon uses 128-byte lines, twice the x86-64 size.
const cacheLineSize = 128
