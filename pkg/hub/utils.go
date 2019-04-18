package hub

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
)

// Claims define jwt claims information
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

const tokenDuration time.Duration = 5 * time.Minute
const jwtIssuer string = "tungle.local"
const jwtSubj string = "VGU Robocon 2019 Login"

func newJWTToken(signKey *rsa.PrivateKey, creds *ClientCredentials) (string, error) {
	jwtIssuedAt := time.Now()
	jwtExpiresAt := jwtIssuedAt.Add(tokenDuration)

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			Issuer:    jwtIssuer,
			Subject:   jwtSubj,
			ExpiresAt: jwtExpiresAt.Unix(),
			IssuedAt:  jwtIssuedAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func httpWriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not encode json; got %v", err)
		return
	}
}
