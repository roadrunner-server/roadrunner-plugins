package directives

import (
	"testing"

	"github.com/spiral/roadrunner/v2/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var noop = zap.NewNop() //nolint:gochecknoglobals

func TestParseRequest(t *testing.T) {
	cc := "foo=bar"
	rq := &Req{}
	ParseRequestCacheControl(cc, noop, rq)

	require.False(t, rq.NoCache)
	require.False(t, rq.NoTransform)
	require.False(t, rq.NoStore)
	require.False(t, rq.OnlyIfCached)
	require.Nil(t, rq.MinFresh)
	require.Nil(t, rq.MaxAge)
	require.Nil(t, rq.MaxStale)

	cc = "max-age="
	rq = &Req{}
	ParseRequestCacheControl(cc, noop, rq)

	require.False(t, rq.NoCache)
	require.False(t, rq.NoTransform)
	require.False(t, rq.NoStore)
	require.False(t, rq.OnlyIfCached)
	require.Nil(t, rq.MinFresh)
	require.Nil(t, rq.MaxAge)
	require.Nil(t, rq.MaxStale)

	cc = "max-age=100, max-stale=100, min-fresh=100, no-cache, no-transform, no-store, only-if-cached"
	rq = &Req{}
	ParseRequestCacheControl(cc, noop, rq)

	require.True(t, rq.NoCache)
	require.True(t, rq.NoTransform)
	require.True(t, rq.NoStore)
	require.True(t, rq.OnlyIfCached)
	require.Equal(t, utils.Uint64(100), rq.MinFresh)
	require.Equal(t, utils.Uint64(100), rq.MaxStale)
	require.Equal(t, utils.Uint64(100), rq.MaxAge)

	cc = "public, max-age=15"
	rq = &Req{}
	ParseRequestCacheControl(cc, noop, rq)

	require.False(t, rq.NoCache)
	require.False(t, rq.NoTransform)
	require.False(t, rq.NoStore)
	require.False(t, rq.OnlyIfCached)
	require.Equal(t, utils.Uint64(15), rq.MaxAge)
	require.Nil(t, rq.MinFresh)
	require.Nil(t, rq.MaxStale)
}

// BenchmarkParseRequest-32    	 2172706	       542.6 ns/op	     376 B/op	      14 allocs/op
// BAD, should not be allocations in that function. TODO(rustatian): rewrite with own lexer and tokenizer.
func BenchmarkParseRequest(b *testing.B) {
	cc := "max-age=100, max-stale=100, min-fresh=100, no-cache, no-transform, no-store, only-if-cached"
	rq := &Req{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ParseRequestCacheControl(cc, noop, rq)
	}
}
