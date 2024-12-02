package token

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Token interface {
	GenerateAccess(userId, userIP string) (string, error)
	GenerateRefresh() (string, error)
	GetHashRefresh(refresh string) (string, error)
	IsActualRefresh(createdTime time.Time) bool
	IsSame(refreshHash, refresh string) bool
}

type Config struct {
	TtlAccess  int
	TtlRefresh int
	Key        string
}

type tokenImpl struct {
	accessToken  *jwt.Token
	refreshToken string
	cfg          Config
}

func New(cfg Config) *tokenImpl {
	return &tokenImpl{
		cfg: cfg,
	}
}

func (t *tokenImpl) GenerateAccess(userId, userIP string) (string, error) {
	t.accessToken = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": userId,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(t.cfg.TtlAccess) * time.Minute).Unix(),
		"ip":  userIP,
	})

	return t.accessToken.SignedString([]byte(t.cfg.Key))
}

func (t *tokenImpl) GenerateRefresh() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func (t *tokenImpl) GetHashRefresh(refresh string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	return string(hash), err
}

func (t *tokenImpl) IsActualRefresh(createdAt time.Time) bool {
	return createdAt.Add(time.Duration(t.cfg.TtlRefresh)*time.Minute).Unix() > time.Now().Unix()
}

func (t *tokenImpl) IsSame(refreshHash, refresh string) bool {
	return bcrypt.CompareHashAndPassword([]byte(refreshHash), []byte(refresh)) == nil
}
