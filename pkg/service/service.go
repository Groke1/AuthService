package service

import (
	"AuthService/pkg"
	r "AuthService/pkg/repository"
	"AuthService/pkg/service/token"
	"context"
	"errors"
	"fmt"
)

var (
	userNotFoundError        = errors.New("user not found")
	refreshTokenInvalidError = errors.New("refresh token is invalid")
)

//go:generate mockgen -source=service.go -destination=../../mocks/mock_service.go -package=mocks

type Service interface {
	GetToken(ctx context.Context, user pkg.User) (pkg.Token, error)
	UpdateToken(ctx context.Context, refreshToken string, user pkg.User) (pkg.Token, error)
}

type serviceImpl struct {
	repo  r.Repository
	token token.Token
}

func New(repo r.Repository, tokenConfig token.Config) *serviceImpl {
	return &serviceImpl{
		repo:  repo,
		token: token.New(tokenConfig),
	}
}

func (s *serviceImpl) getTokenImpl(user pkg.User) (pkg.Token, error) {
	var tkn pkg.Token
	var err error

	if tkn.AccessToken, err = s.token.GenerateAccess(user.UserId, user.IP); err != nil {
		return tkn, err
	}

	if tkn.RefreshToken, err = s.token.GenerateRefresh(); err != nil {
		return tkn, err
	}

	if tkn.RefreshHash, err = s.token.GetHashRefresh(tkn.RefreshToken); err != nil {
		return tkn, err
	}

	return tkn, nil
}

func (s *serviceImpl) GetToken(ctx context.Context, user pkg.User) (pkg.Token, error) {
	var tkn pkg.Token
	isExists, err := s.repo.IsExists(ctx, user.UserId)
	if err != nil {
		return tkn, err
	}
	if !isExists {
		return tkn, userNotFoundError
	}

	tkn, err = s.getTokenImpl(user)
	if err != nil {
		return tkn, err
	}

	if err = s.repo.AddSession(ctx, user, tkn.RefreshHash); err != nil {
		return tkn, err
	}

	return tkn, nil
}

func (s *serviceImpl) UpdateToken(ctx context.Context, refreshToken string, user pkg.User) (pkg.Token, error) {

	var tkn pkg.Token
	sessions, err := s.repo.GetSessions(ctx, user.UserId)
	if err != nil {
		return tkn, err
	}

	var nowSession *pkg.Session
	for _, session := range sessions {
		if s.token.IsSame(session.RefreshHash, refreshToken) {
			nowSession = &session
		}
	}

	if nowSession == nil || !s.token.IsActualRefresh(nowSession.CreatedAt) {
		return tkn, refreshTokenInvalidError
	}

	if user.IP != nowSession.UserIP {
		s.sendMsg(nowSession.UserEmail, "Warning message")
	}

	if tkn, err = s.getTokenImpl(pkg.User{
		UserId: nowSession.UserId,
		IP:     nowSession.UserIP,
	}); err != nil {
		return tkn, err
	}

	if err = s.repo.UpdateSession(ctx, nowSession.RefreshHash, tkn.RefreshHash); err != nil {
		return tkn, err
	}
	return tkn, nil
}

func (s *serviceImpl) sendMsg(email, msg string) {
	fmt.Println(msg, "to ", email)
}
