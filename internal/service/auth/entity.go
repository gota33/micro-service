package auth

import (
	"context"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gota33/errors"
)

type ctxKey int

const (
	unknown ctxKey = iota
	userKey
)

var parser = &jwt.Parser{}

type User struct {
	jwt.StandardClaims
	Authorization string `json:"-"`
	Nick          string `json:"nick,omitempty"`
	// TODO: More user fields
}

func (u *User) FromJWT(str string) (err error) {
	u.Authorization = str
	token := strings.TrimPrefix(str, "Bearer ")
	_, _, err = parser.ParseUnverified(token, u)
	return
}

func (u User) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func (u *User) FromContext(ctx context.Context) (err error) {
	var ok bool
	if *u, ok = ctx.Value(userKey).(User); !ok {
		return errors.Unauthenticated
	}
	return
}
