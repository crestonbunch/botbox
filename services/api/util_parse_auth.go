package api

import (
	"net/http"
	"strings"
)

const AuthorizationHeader = "Authorization"
const AuthorizationPrefix = "Bearer"

func ParseAuthSecret(header http.Header) string {
	contents := header.Get(AuthorizationHeader)
	return strings.TrimPrefix(contents, AuthorizationPrefix+" ")
}
