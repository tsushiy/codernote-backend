package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

const (
	audience = "codernote-project"
	issuer   = "https://securetoken.google.com/codernote-project"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicKeyMap, err := fetchPublicKeyMap()
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch public key", http.StatusInternalServerError)
			return
		}

		token, err := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("invalid signing method")
			}
			if token.Method != jwt.SigningMethodRS256 {
				return nil, errors.New("invalid signing method")
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("token header should have kid field")
			}
			pubKey, ok := publicKeyMap[kid]
			if !ok {
				return nil, errors.New("invalid public key id")
			}

			return pubKey, nil
		})
		if err != nil {
			log.Println(err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if !isValid(token) {
			log.Println(errors.New("invalid token"))
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		uid := claims["sub"].(string)
		if uid == "" {
			return
		}
		ctx := context.WithValue(r.Context(), "uid", uid)
		r = r.WithContext(ctx)

		// log.Printf("Got a valid token. Header: %v Claims: %v", token.Header, token.Claims)
		next.ServeHTTP(w, r)
	})
}

type publicKeyMap map[string]*rsa.PublicKey

func fetchPublicKeyMap() (km publicKeyMap, err error) {
	const srcURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

	resp, err := http.Get(srcURL)
	if err != nil {
		return nil, errors.New("error fetching public key: " + err.Error())
	}
	defer resp.Body.Close()

	d := make(map[string]string)
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, errors.New("error decoding public key http response body: " + err.Error())
	}

	keyMap := make(map[string]*rsa.PublicKey)
	for k, v := range d {
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(v))
		if err != nil {
			return nil, errors.New("error parsing public key: " + err.Error())
		}
		keyMap[k] = key
	}

	return keyMap, nil
}

func isValid(token *jwt.Token) bool {
	now := time.Now().Unix()
	if !token.Valid {
		return false
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims.VerifyExpiresAt(now, true) == false {
		return false
	}
	if claims.VerifyIssuedAt(now, true) == false {
		return false
	}
	if claims.VerifyAudience(audience, true) == false {
		return false
	}
	if claims.VerifyIssuer(issuer, true) == false {
		return false
	}
	if sub, ok := claims["sub"].(string); !ok || sub == "" {
		return false
	}
	if authTime, ok := claims["auth_time"].(float64); !ok || int64(authTime) > now {
		return false
	}
	return true
}
