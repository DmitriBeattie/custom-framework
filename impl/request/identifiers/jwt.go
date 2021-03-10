package identifiers

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

type JWTKey struct {
	key *rsa.PublicKey
	ignoreErrors []uint32
}

func JWTKeyFromPath(path string, ignoreErrors []uint32) (*JWTKey, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cc, err := x509.ParseCertificates(b)
	if err != nil {
		return nil, err
	}

	for _, c := range cc {
		if c.PublicKey == nil {
			continue
		}

		if pub, ok := c.PublicKey.(*rsa.PublicKey); ok {
			return &JWTKey{
				key: pub,
				ignoreErrors: ignoreErrors,
			}, nil
		}
	}

	return nil, errors.New("could not find rsa256 public key")
}

func (j *JWTKey) calculateIgnoreErrorMask() uint32 {
	var jwtIgnoreErrorAccum uint32 = 0

	for _, err := range j.ignoreErrors {
		jwtIgnoreErrorAccum |= err
	}

	return jwtIgnoreErrorAccum
}

func (j *JWTKey) Identify(r *http.Request) ([]string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errors.New("Не передан header")
	}

	token := strings.Split(header, " ")
	if len(token) < 2 {
		return nil, errors.New("Токен не найден")
	}

	if j.key == nil {
		return nil, errors.New("Не удалось найти публичный ключ для проверки токена")
	}

	tok, err := jwt.Parse(token[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("Parse token error")
		}
		return j.key, nil
	})
	if err != nil {
		if len(j.ignoreErrors) == 0 {
			return nil, err
		}

		jwtError, ok := err.(*jwt.ValidationError)
		if !ok {
			return nil, err
		}

		ignoreErrorAccum := j.calculateIgnoreErrorMask()

		if ignoreErrorAccum | jwtError.Errors != ignoreErrorAccum {
			return  nil, err
		}

		tok.Valid = true
	}
	if !tok.Valid {
		return nil, errors.New("Токен не валиден")
	}

	payload, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Не удалось прочитать необходимую информацию из токена")
	}

	if employeeID, ok := payload["id"].(string); ok {
		context.Set(r, "employeeID", employeeID)
	} else {
		context.Set(r, "employeeID", "undefined")
	}

	if scopes, ok := payload["scope"].([]string); ok {
		return scopes, nil
	}

	return nil, nil
}
