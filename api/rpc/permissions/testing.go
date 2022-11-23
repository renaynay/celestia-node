package permissions

import "github.com/gbrlsnchs/jwt/v3"

func DummySecret() *jwt.HMACSHA {
	return jwt.NewHS256(make([]byte, 32))
}
