package auth

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

const TokenDuration = time.Hour * 24

var ErrorAuth = fmt.Errorf("auth error")
var ErrorCreateUser = fmt.Errorf("can't create user")

type Config struct {
	PublicKeyPath        string
	PrivateKeyPath       string
	DefaultAdminPassword string
}

type Token struct {
	AccessToken string `json:"access_token"`
}

type JWTAuth struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	cfg        *Config
	db         *DB
}

func NewAuth(conf Config, db *DB) (*JWTAuth, error) {
	jwtAuth := &JWTAuth{
		cfg: &conf,
		db:  db,
	}

	var err error
	if conf.PrivateKeyPath != "" {
		jwtAuth.privateKey, err = getPrivateKey(conf)
		if err != nil {
			err = fmt.Errorf("error loading private auth key: %s", err)
			return nil, err
		} else {
			logrus.Info("private auth key loaded")
		}
	}

	if conf.PublicKeyPath != "" {
		jwtAuth.PublicKey, err = getPublicKey(conf)
		if err != nil {
			err = fmt.Errorf("error loading public auth key: %s", err)
			return nil, err
		} else {
			logrus.Info("public auth key loaded")
		}
	}

	count := 0
	resDB := db.DB().Table("users").Count(&count)
	if resDB.Error != nil {
		err = fmt.Errorf("failed to count users: %s", resDB.Error)
		return nil, err
	}

	if count == 0 {
		password := conf.DefaultAdminPassword
		username := "admin"

		// create admin user automatically
		_, _, err = jwtAuth.Register(username, password)
		if err != nil {
			err = fmt.Errorf("failed to register user: %s", err)
			return nil, err
		}
		logrus.Warningf("new user created:\nUsername: %s\nPassword:%s", username, password)
	}

	return jwtAuth, nil
}

func (j *JWTAuth) GenerateToken(userID uint) (string, error) {
	return j.generateExpToken(userID, TokenDuration)
}

func (j *JWTAuth) generateExpToken(userID uint, duration time.Duration) (string, error) {
	t := jwt.New(jwt.SigningMethodRS512)
	claims := t.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(duration).Unix()
	claims["iat"] = time.Now().Unix()
	claims["sub"] = userID

	t.Claims = claims

	if j.privateKey == nil {
		err := fmt.Errorf("can't generate token - private key not loaded")
		logrus.Error(err)
		return "", err
	}

	tokenString, err := t.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTAuth) Authenticate(username, password string) (res Token, err error) {
	user, err := j.db.UserByName(username)
	if err != nil {
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil {
		res.AccessToken, err = j.GenerateToken(user.ID)
		return
	}

	err = ErrorAuth

	return
}

func (j *JWTAuth) Validate(req *http.Request) (bool, uint64) {

	token, err := j.ParseTokenFromRequest(req)

	if err != nil || !token.Valid {
		err = fmt.Errorf("token not valid: %s", err)
		logrus.Debug(err)
		return false, 0
	}

	ts := token.Claims.(jwt.MapClaims)["token_signature"]
	if ts != nil {
		return false, 0
	}

	userID := uint64(token.Claims.(jwt.MapClaims)["sub"].(float64))

	return true, userID
}

func (j *JWTAuth) ParseTokenFromRequest(req *http.Request) (*jwt.Token, error) {
	if j.PublicKey == nil {
		err := fmt.Errorf("can't validate token - public key not loaded")
		logrus.Error(err)
		return nil, err
	}

	token, err := request.ParseFromRequest(req, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else {
			return j.PublicKey, nil
		}
	})

	return token, err
}

func (j *JWTAuth) ParseToken(token string) (*jwt.Token, error) {
	if j.PublicKey == nil {
		err := fmt.Errorf("can't validate token - public key not loaded")
		logrus.Error(err)
		return nil, err
	}

	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else {
			return j.PublicKey, nil
		}
	})
}

func getPrivateKey(cfg Config) (*rsa.PrivateKey, error) {
	privateKeyFile, err := os.Open(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	pemFileInfo, _ := privateKeyFile.Stat()
	size := pemFileInfo.Size()
	pemBytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pemBytes)

	data, _ := pem.Decode([]byte(pemBytes))

	err = privateKeyFile.Close()
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, err
	}

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		return nil, err
	}

	return privateKeyImported, nil
}

func getPublicKey(cfg Config) (*rsa.PublicKey, error) {
	publicKeyFile, err := os.Open(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	err = publicKeyFile.Close()
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, err
	}

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)
	if !ok {
		return nil, err
	}

	return rsaPub, nil
}

func (j *JWTAuth) Register(username, password string) (*Token, *User, error) {
	// поиск юзера по логину и паролю

	newUser := User{}
	newUser.Username = username

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		logrus.Error(err)
		return nil, nil, ErrorCreateUser
	}
	newUser.PasswordHash = string(hash)

	user, err := j.db.UserCreate(&newUser)
	if err != nil {
		logrus.Error(err)
		return nil, nil, ErrorCreateUser

	}

	// отдаем токены для нового клиента
	t, err := j.GenerateToken(user.ID)

	if err != nil {
		logrus.Error(err)
		return nil, nil, ErrorCreateUser
	}

	return &Token{AccessToken: t}, user, nil
}
