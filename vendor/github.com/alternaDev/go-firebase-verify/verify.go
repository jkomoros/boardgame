package firebase

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/dgrijalva/jwt-go.v3"
)

const (
	clientCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

func VerifyIDToken(idToken string, googleProjectID string) (string, error) {
	keys, err := fetchPublicKeys()

	if err != nil {
		return "", err
	}

	parsedToken, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid := token.Header["kid"]

		certPEM := string(*keys[kid.(string)])
		certPEM = strings.Replace(certPEM, "\\n", "\n", -1)
		certPEM = strings.Replace(certPEM, "\"", "", -1)
		block, _ := pem.Decode([]byte(certPEM))
		var cert *x509.Certificate
		cert, _ = x509.ParseCertificate(block.Bytes)
		rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

		return rsaPublicKey, nil
	})

	if err != nil {
		return "", err
	}

	errMessage := ""

	claims := parsedToken.Claims.(jwt.MapClaims)

	if claims["aud"].(string) != googleProjectID {
		errMessage = "Firebase Auth ID token has incorrect 'aud' claim: " + claims["aud"].(string)
	} else if claims["iss"].(string) != "https://securetoken.google.com/"+googleProjectID {
		errMessage = "Firebase Auth ID token has incorrect 'iss' claim"
	} else if claims["sub"].(string) == "" || len(claims["sub"].(string)) > 128 {
		errMessage = "Firebase Auth ID token has invalid 'sub' claim"
	}

	if errMessage != "" {
		return "", errors.New(errMessage)
	}

	return string(claims["sub"].(string)), nil
}

func fetchPublicKeys() (map[string]*json.RawMessage, error) {
	resp, err := http.Get(clientCertURL)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var objmap map[string]*json.RawMessage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&objmap)

	return objmap, err
}
