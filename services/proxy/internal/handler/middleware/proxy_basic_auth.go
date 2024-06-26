package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ProxyUserCreditanalsKey is key for user creditanals in request context
const proxyUserProxyBasicAuthKey key = "proxy_user_creditanals"

var ErrGetProxyUserCreditanals = errors.New("failed get proxy user creditanals")

type Creditanals struct {
	Username string
	Password string
}

// ProxyBasicAuth is middleware for proxy basic authorization
// if Proxy-Authorization header not exists or basic auth is invalid
// then response with Proxy Auth required with given realm.
// Otherwise user creditanals go along with context value under the key ProxyUserCreditanals
func ProxyBasicAuth(realm string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := getProxyBasicAuth(r)
			if !ok {
				ProxyBasicAuthFailed(w, realm)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				proxyUserProxyBasicAuthKey,
				Creditanals{
					Username: username,
					Password: password,
				})
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func ProxyUserCreditanalsFromContext(ctx context.Context) (Creditanals, error) {
	if creditanals, ok := ctx.Value(proxyUserProxyBasicAuthKey).(Creditanals); ok {
		return creditanals, nil
	}

	return Creditanals{}, ErrGetProxyUserCreditanals
}

func ProxyBasicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("Proxy-Authenticate", fmt.Sprintf("Basic realm=%s", realm))
	w.WriteHeader(http.StatusProxyAuthRequired)
}

func getProxyBasicAuth(r *http.Request) (string, string, bool) {
	proxyBasicAuth := r.Header.Get("Proxy-Authorization")
	if proxyBasicAuth == "" {
		return "", "", false
	}

	splitProxyAuth := strings.Split(proxyBasicAuth, " ")
	if len(splitProxyAuth) != 2 || splitProxyAuth[0] != "Basic" {
		return "", "", false
	}

	basicAuth, err := base64.StdEncoding.DecodeString(splitProxyAuth[1])
	if err != nil {
		return "", "", false
	}

	creditanals := strings.Split(string(basicAuth), ":")
	if len(creditanals) != 2 {
		return "", "", false
	}

	return creditanals[0], creditanals[1], true
}
