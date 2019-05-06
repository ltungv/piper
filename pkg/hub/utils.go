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
	Role     string `json:"role"`
	jwt.StandardClaims
}

const contestantTokenDuration time.Duration = 5 * time.Minute
const adminTokenDuration time.Duration = 48 * time.Hour
const jwtIssuer string = "REC"
const jwtSubj string = "VGU Robocon 2019 Login"

// create new jwt from rsa private key and user credentials
func newJWTToken(signKey *rsa.PrivateKey, creds *ClientCredentials) (string, error) {
	jwtIssuedAt := time.Now()

	var tokenDuration time.Duration
	switch creds.Role {
	case "contestant":
		tokenDuration = contestantTokenDuration
	case "admin":
		tokenDuration = adminTokenDuration
	}

	jwtExpiresAt := jwtIssuedAt.Add(tokenDuration)

	claims := &Claims{
		Username: creds.Username,
		Role:     creds.Role,
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

// json http response helper
func httpWriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("could not encode json; got %v", err)
		return
	}
}
