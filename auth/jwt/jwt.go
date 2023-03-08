package jwt

import "github.com/golang-jwt/jwt/v4"

func CreateToken(kid string, key []byte, method jwt.SigningMethod, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = kid

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
