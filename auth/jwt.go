package auth

import (
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// 这是加密秘钥
var jwtSecret = []byte("2020jxc")

type Claims struct {
	Username string `json:"username"`
	UserId   int64  `json:"user_id"`
	ComId    int64  `json:"com_id"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

// 利用时间戳生成token
func GenerateToken(username string, userId int64, comId int64, admin bool) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(3 * time.Hour)

	claims := Claims{
		username,
		userId,
		comId,
		admin,
		jwt.StandardClaims{
			Subject:   "",
			ExpiresAt: expireTime.Unix(),
			Issuer:    "jxc",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
