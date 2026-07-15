package app

import (
	"bytes"
	"sync"
	"testing"
)

// resetMemCache empties the in-memory cache so each test starts clean.
func resetMemCache() {
	mx.Lock()
	defer mx.Unlock()
	tempFS = make(map[[20]byte]*file)
}

func key(b byte) [20]byte {
	var k [20]byte
	k[0] = b
	return k
}

// TestMemCacheConcurrentAccess hammers the in-memory cache from many goroutines
// while the janitor ages it, which is what a media flood does on an instance
// with memcache enabled.
//
// This is the regression test for the readers that touched tempFS without
// holding mx: concurrently with the janitor's delete that is a concurrent map
// read and map write, which the runtime reports as a fatal error that no
// recover can catch. Run under -race to also catch the unsynchronised field
// access that does not happen to trip the map check.
func TestMemCacheConcurrentAccess(t *testing.T) {
	resetMemCache()
	defer resetMemCache()

	const workers, rounds = 24, 200
	body := []byte("not-really-an-image")

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for i := range rounds {
				// Overlapping keys, so goroutines contend for the same entries.
				k := key(byte((w + i) % 8)) //nolint:gosec // G115: (w+i)%8 is 0-7
				memPut(k, body)
				memGet(k)
			}
		})
	}

	// Age the cache underneath the readers and writers: this is the delete that
	// the old per-entry goroutines raced against.
	wg.Go(func() {
		for range rounds {
			ageMemCache()
		}
	})

	wg.Wait()
}

// TestMemGetReturnsStoredBody covers the plain hit and miss paths.
func TestMemGetReturnsStoredBody(t *testing.T) {
	resetMemCache()
	defer resetMemCache()

	k := key(1)
	if got := memGet(k); got != nil {
		t.Fatalf("empty cache: got %q, want nil", got)
	}

	want := []byte("body")
	memPut(k, want)

	got := memGet(k)
	if !bytes.Equal(got, want) {
		t.Fatalf("after put: got %q, want %q", got, want)
	}
}

// TestMemPutIgnoresEmptyBody stops a failed fetch from caching a zero-length
// image that would then be served to everyone until it aged out.
func TestMemPutIgnoresEmptyBody(t *testing.T) {
	resetMemCache()
	defer resetMemCache()

	k := key(2)
	memPut(k, nil)
	memPut(k, []byte{})

	if got := memGet(k); got != nil {
		t.Fatalf("empty body was cached: got %q, want nil", got)
	}
}

// TestAgeMemCacheEvicts checks that a cold entry is dropped while a hot one
// survives, since that scoring is the only bound on the cache's memory use.
func TestAgeMemCacheEvicts(t *testing.T) {
	resetMemCache()
	defer resetMemCache()

	cold, hot := key(3), key(4)
	memPut(cold, []byte("cold"))
	memPut(hot, []byte("hot"))

	// A hit raises the hot entry's score above zero.
	memGet(hot)

	ageMemCache()

	if got := memGet(cold); got != nil {
		t.Errorf("cold entry survived aging: got %q, want nil", got)
	}
	if got := memGet(hot); got == nil {
		t.Error("hot entry was evicted after a hit, want it kept")
	}
}
