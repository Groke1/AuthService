package handler

import (
	"AuthService/mocks"
	"AuthService/pkg"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var serviceError = errors.New("service error")

type testCase struct {
	name         string
	user         pkg.User
	refreshToken string
	returnToken  pkg.Token
	returnError  error
}

var testCases = []testCase{
	{
		name: "Ok",
		user: pkg.User{
			UserId: "1",
			IP:     "192.1.1.1",
		},
		refreshToken: "token",
		returnToken: pkg.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
		returnError: nil,
	},
	{
		name: "Error",
		user: pkg.User{
			UserId: "1",
			IP:     "192.1.1.1",
		},
		refreshToken: "token",
		returnToken: pkg.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
		returnError: serviceError,
	},
}

func testInit(t *testing.T) (*mocks.MockService, *mux.Router) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockService := mocks.NewMockService(mockCtrl)

	handler := New(mockService)
	router := mux.NewRouter()
	handler.InitRoutes(router)
	return mockService, router
}

func testHelper(t *testing.T, router *mux.Router, tc *testCase,
	target string, body io.Reader) {
	req := httptest.NewRequest(http.MethodPost, target, body)
	req.RemoteAddr = tc.user.IP
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if tc.returnError == nil {
		require.Equal(t, http.StatusOK, res.StatusCode)
		var tkn pkg.Token
		require.NoError(t, json.NewDecoder(res.Body).Decode(&tkn))
		require.Equal(t, tc.returnToken, tkn)
	} else {
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}

func TestHandler_Authorization(t *testing.T) {
	mockService, router := testInit(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.EXPECT().GetToken(gomock.Any(), tc.user).
				Return(tc.returnToken, tc.returnError)
			testHelper(t, router, &tc, "/auth/"+tc.user.UserId, nil)
		})
	}
}

func TestHandler_UpdateToken(t *testing.T) {
	mockService, router := testInit(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.EXPECT().UpdateToken(gomock.Any(), tc.refreshToken, tc.user).
				Return(tc.returnToken, tc.returnError)
			testHelper(t, router, &tc, "/auth/refresh/"+tc.user.UserId,
				bytes.NewBufferString(`{"refresh_token": "`+tc.refreshToken+`"}`))
		})
	}
}
