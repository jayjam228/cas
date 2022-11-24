package utils

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

var JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT = "360"

func GenerateNewAccessToken() (string, error){
	
	rand.Seed(time.Now().UnixNano())
    JWT_SECRET_KEY :=RandomString(40)
	secret := os.Getenv(JWT_SECRET_KEY)

	minutesCount, _ := strconv.Atoi(JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT)

    claims := jwt.StandardClaims{
		Id: "637b3ca2c84b3c8d7828e66e",
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
	}

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    t, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", err
    }

    return t, nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func RandomString(n int) string {
    sb := strings.Builder{}
    sb.Grow(n)
    for i := 0; i < n; i++ {
        sb.WriteByte(charset[rand.Intn(len(charset))])
    }
    return sb.String()
}

