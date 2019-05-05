package hub

import (
	"context"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
)

type ctxKey string

const usernameKey = ctxKey("username")
const roleKey = ctxKey("role")

// JWTProtect verifies jwt before continue
func (h *Hub) JWTProtect(role string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// get jwt from header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusBadRequest)
				log.Infof("no authorization found")
				return
			}

			reqToken := strings.Split(auth, "Bearer")[1]
			reqToken = strings.TrimSpace(reqToken)
			if reqToken == "" {
				w.WriteHeader(http.StatusBadRequest)
				log.Infof("no token found")
				return
			}

			// parse and validate token
			token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
				// since we only use the one private key to sign the tokens,
				// we also only use its public counter part to verify
				return h.jwtVerify, nil
			})

			// check for token error
			switch err.(type) {
			case nil: // no error
				if !token.Valid {
					w.WriteHeader(http.StatusUnauthorized)
					log.Infof("invalid token")
					return
				}
			case *jwt.ValidationError:
				vErr := err.(*jwt.ValidationError)
				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					w.WriteHeader(http.StatusUnauthorized)
					log.Infof("token expired")
					return

				default:
					w.WriteHeader(http.StatusInternalServerError)
					log.Errorf("could not parse token; got %v", vErr)
					return
				}
			default: // something else went wrong
				w.WriteHeader(http.StatusInternalServerError)
				log.Errorf("could not parse token; got %v", err)
				return
			}

			// parse claims to get client username
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				log.Errorf("could not parse claims")
				return
			}

			claimUsername, ok := claims["username"].(string)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				log.Infof("no username found in jwt claims")
				return
			}

			claimRole, ok := claims["role"].(string)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				log.Infof("no role found in jwt claims")
				return
			}

			if claimRole != role {
				w.WriteHeader(http.StatusBadRequest)
				log.Infof("invalid role for path")
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, claimUsername)
			ctx = context.WithValue(ctx, roleKey, claimRole)
			next(w, r.WithContext(ctx))
		}
	}
}
