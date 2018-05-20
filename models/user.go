package models

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Email    string
	Password string
	Tokens   []string
	Profile
}

type Profile struct {
	LastName  string
	FirstName string
}

type ConfigMongo struct {
	Mongo struct {
		Host       string
		DB         string
		Collection string
	}
}

var SignKey = []byte("AllYourBase")

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func ValidatePass(user *User) error {
	if len(user.Password) < 10 {
		return errors.New("passord is small")
	}
	return nil
}

func CreateToken(c *mgo.Collection, user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"mls":   time.Now().UnixNano() / int64(time.Millisecond),
	})
	tokenString, err := token.SignedString(SignKey)
	if err != nil {
		log.Println(err)
	}
	user.Tokens = append(user.Tokens, tokenString)
	return tokenString, err
}

func InsertTokenToDb(c *mgo.Collection, user *User) error {
	// Update
	colQuerier := bson.M{"email": user.Email}
	change := bson.M{"$set": bson.M{"Tokens": user.Tokens}}
	return c.Update(colQuerier, change)
}

func NewMongoDbCollection(cfg *ConfigMongo) (*mgo.Collection, error) {
	session, err := mgo.Dial(cfg.Mongo.Host)
	if err != nil {
		return nil, nil
	}
	return session.DB(cfg.Mongo.DB).C(cfg.Mongo.Collection), err
}

func GetUserByEmail(c *mgo.Collection, userdata *User) (*User, error) {
	user := User{}
	err := c.Find(bson.M{"email": userdata.Email}).One(&user)
	return &user, err
}

func FindByEmail(c *mgo.Collection, userdata *User) *mgo.Query {
	q := c.Find(bson.M{"email": userdata.Email})
	return q
}

func Insert(c *mgo.Collection, user *User) error {
	return c.Insert(user)
}

func Singout(c *mgo.Collection) (*User, error) {
	user := &User{}
	return user, nil
}

func EditProfile(c *mgo.Collection, user *User) (*User, error) {
	user = &User{}
	return user, nil
}
