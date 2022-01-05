package cache

/*
   In the initial release, the cache middleware doesn't support the Range and Content-Range headers.
   https://datatracker.ietf.org/doc/html/rfc7234#section-3.1
   a cache MUST NOT store
   incomplete or partial-content responses if it does not support the
   Range and Content-Range header fields or if it does not understand
   the range units used in those fields.


   https://datatracker.ietf.org/doc/html/rfc7234#section-4
   cache should not use stored responses
*/
