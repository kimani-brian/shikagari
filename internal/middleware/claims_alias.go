package middleware

import jwtpkg "github.com/shikagari/api/pkg/jwt"

// Claims is a re-export of jwt.Claims for use by handler packages
// without them needing to import the jwt package directly.
type Claims = jwtpkg.Claims
