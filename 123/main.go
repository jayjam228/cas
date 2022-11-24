package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"workspace/configs"
	"workspace/middleware"
	"workspace/models"
	"workspace/utils"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)
var ctx = context.Background()

var client = redis.NewClient(&redis.Options{
    Addr: "192.168.31.34:6379",
    Password: "",
    DB: 10,
    
})


var UserCollection *mongo.Collection = configs.GetCollection(configs.DB, "Users")
func CreateUser(c *fiber.Ctx) error {
    ctx ,cancel := context.WithTimeout(context.Background(), 10*time.Second)
    var user models.User
    defer cancel()

    if err := c.BodyParser(&user); 
	err != nil{
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"data": err.Error()})
	}

    newUser := models.User{
        Username: user.Username,
        Password: user.Password,
    }

    fmt.Println(newUser.Username)

    result, err := UserCollection.InsertOne(ctx, newUser)
    if err != nil {
        panic(err)
    }

    return c.Status(http.StatusAccepted).JSON(&fiber.Map{"data": result})
}

func GenerateTokenForUser(c *fiber.Ctx)  error{
    ctx ,cancel := context.WithTimeout(context.Background(), 10*time.Second)
    var user models.User
    defer cancel()
    username := c.Params("username")

    mongoErr := UserCollection.FindOne(ctx,bson.M{"username" : username}).Decode(&user)
    println(mongoErr)
    userId:=user.Id.Hex()
    println(userId)
    rand.Seed(time.Now().UnixNano())
    JWT_SECRET_KEY :=utils.RandomString(40)
	secret := os.Getenv(JWT_SECRET_KEY)

	minutesCount, _ := strconv.Atoi(utils.JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT)

    claims := jwt.StandardClaims{
		Id: userId,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix(),
	}

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    t, err := token.SignedString([]byte(secret))
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"data": mongoErr.Error()})
    }else{
        return c.Status(http.StatusAccepted).JSON(&fiber.Map{"data":t})
    }
}

func FindUser(c *fiber.Ctx) error {
    ctx ,cancel := context.WithTimeout(context.Background(), 10*time.Second)
    var user models.User
    defer cancel()
    userId := c.Params("userId")
    objId, _ := primitive.ObjectIDFromHex(userId)

    mongoErr := UserCollection.FindOne(ctx,bson.M{"_id" : objId}).Decode(&user)
    println(mongoErr)
	if mongoErr != nil {
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"data": mongoErr.Error()})
	} else {
		return c.Status(http.StatusAccepted).JSON(&fiber.Map{"data":user})
	}

}
const JWTAuthScheme string = "Bearer"
func Token(c *fiber.Ctx) error{
    auth, _:=jwtFromHeader(c, fiber.HeaderAuthorization, JWTAuthScheme)
    
    ctx ,cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    rand.Seed(time.Now().UnixNano())
    JWT_SECRET_KEY :=utils.RandomString(40)
    token := auth
    var user models.User 
    type MyClaims struct {
        jwt.StandardClaims
        MyField string `json:"my_field"`
    }

    tokenString := token

    // pass your custom claims to the parser function
    token3, _ := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(JWT_SECRET_KEY), nil
    })
    fmt.Println(token3)
    // type-assert `Claims` into a variable of the appropriate type
    myClaims := token3.Claims.(*MyClaims)
    fmt.Println("id",myClaims.Id)
    userId:=myClaims.Id

    objId, _ := primitive.ObjectIDFromHex(userId)


    mongoErr := UserCollection.FindOne(ctx,bson.M{"_id" : objId}).Decode(&user)
    println(mongoErr)
	if mongoErr != nil {
		return c.Status(http.StatusBadRequest).JSON(&fiber.Map{"data": mongoErr.Error()})
	} else {
		return c.Status(http.StatusAccepted).JSON(&fiber.Map{"data":user})
	}
}
//637b3ca2c84b3c8d7828e66e
func jwtFromHeader(c *fiber.Ctx, header string, authScheme string) (string, error) {
	auth := c.Get(header)
	l := len(authScheme)
	if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
		return auth[l+1:], nil
	}
        return "", errors.New("missing or malformed JWT")
	
}

func main() {
    
    fmt.Println("Testing Golang Redis")
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
    println(utils.GenerateNewAccessToken())
    token, _ := utils.GenerateNewAccessToken()
    rand.Seed(time.Now().UnixNano())
    JWT_SECRET_KEY :=utils.RandomString(40)

    


    configs.ConnectDB()

    app := fiber.New()
    app.Post("/Create/User", CreateUser)
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })
    app.Get("/Get/User/:userId", middleware.JWTProtected(), FindUser)
    app.Post("/Token",Token)
    app.Post("/Get/Token/:username",GenerateTokenForUser)
    app.Listen(":3000")

    minutesCount,_:= strconv.Atoi(utils.JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT)
    fmt.Println(minutesCount)
    min := time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix()
    sec:=strconv.Itoa(int(min))
    
    date := time.Now().Unix()
    date1 := (min-date)*1000000000
    fmt.Println(date1)
    
    
    json, err := json.Marshal(models.JWT{Token: token, Secret_key: JWT_SECRET_KEY, Algorithm:"HS256", Expire: sec})
    

	err = client.Set("id1234", json, time.Duration(date1)).Err()
    if err != nil {
        fmt.Println(err)
    }
    val, err := client.Get("id1234").Result()
    if err != nil {
        fmt.Println(err)
    }

	fmt.Println(val)

    
}


 