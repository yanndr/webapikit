package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
)

const (

	// ClaimsContextKey holds the key used to store the JWT Claims in the
	// context.
	ClaimsContextKey string = "JWTClaims"
)

var (

	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("JWT was invalid")

	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("JWT is expired")

	// ErrTokenMalformed denotes a token was not formatted as a JWT.
	ErrTokenMalformed = errors.New("JWT is malformed")

	// ErrTokenNotActive denotes a token's not before header (nbf) is in the
	// future.
	ErrTokenNotActive = errors.New("token is not valid yet")

	// ErrUnexpectedSigningMethod denotes a token was signed with an unexpected
	// signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

func IsAuthorized(keyFunc jwt.Keyfunc, method jwt.SigningMethod, newClaims func() jwt.Claims) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Header["Authorization"] == nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, "Not Authorized")
				return
			}

			reqToken := r.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, "Not Authorized")
				return
			}
			tokenString := splitToken[1]
			claims := newClaims()
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if token.Method != method {
					return nil, ErrUnexpectedSigningMethod
				}

				return keyFunc(token)
			})
			if err != nil {
				var errorMsg string
				if e, ok := err.(*jwt.ValidationError); ok {
					switch {
					case e.Errors&jwt.ValidationErrorMalformed != 0:
						// Token is malformed
						errorMsg = ErrTokenMalformed.Error()
					case e.Errors&jwt.ValidationErrorExpired != 0:
						// Token is expired
						errorMsg = ErrTokenExpired.Error()
					case e.Errors&jwt.ValidationErrorNotValidYet != 0:
						// Token is not active yet
						errorMsg = ErrTokenNotActive.Error()
					case e.Inner != nil:
						// report e.Inner
						errorMsg = e.Inner.Error()
					}
					// We have a ValidationError but have no specific Go kit error for it.
					// Fall through to return original error.
				} else {
					errorMsg = err.Error()
				}

				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, errorMsg)
				return
			}

			if !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, ErrTokenInvalid.Error())
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
