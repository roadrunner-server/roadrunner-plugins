package cache

/*
Cache-Control keys and values https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#cache_directives
*/

const (
	auth         string = "Authorization"
	cacheControl string = "Cache-Control"
	// ----- REQUEST -----
	maxAge       string = "max-age"   //nolint:deadcode,varcheck
	maxStale     string = "max-stale" //nolint:deadcode,varcheck
	minFresh     string = "min-fresh" //nolint:deadcode,varcheck
	noCache      string = "no-cache"  //nolint:deadcode,varcheck
	noStore      string = "no-store"
	noTransform  string = "no-transform"   //nolint:deadcode,varcheck
	onlyIfCached string = "only-if-cached" //nolint:deadcode,varcheck
	staleIfErr   string = "stale-if-error" //nolint:deadcode,varcheck

	vary  string = "Vary" //nolint:deadcode,varcheck
	empty string = ""
)
