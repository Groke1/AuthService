package handler

import (
	"AuthService/pkg"
	"AuthService/pkg/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	serv service.Service
}

func New(serv service.Service) *Handler {
	return &Handler{
		serv: serv,
	}
}

func (h *Handler) InitRoutes(router *mux.Router) {
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/{user_id}", h.Authorization).Methods(http.MethodPost)
	authRouter.HandleFunc("/refresh/{user_id}", h.UpdateToken).Methods(http.MethodPost)
}

func (h *Handler) jsonResponse(w http.ResponseWriter, getResp func() (any, error)) {
	resp, err := getResp()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Authorization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := pkg.User{UserId: vars["user_id"], IP: h.getIP(r)}
	h.jsonResponse(w, func() (any, error) {
		return h.serv.GetToken(r.Context(), user)
	})
}

func (h *Handler) UpdateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := pkg.User{UserId: vars["user_id"], IP: h.getIP(r)}
	h.jsonResponse(w, func() (any, error) {
		refreshToken, err := h.getRefreshToken(r)
		if err != nil {
			return nil, err
		}
		return h.serv.UpdateToken(r.Context(), refreshToken, user)
	})
}

func (h *Handler) getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

func (h *Handler) getRefreshToken(r *http.Request) (string, error) {
	mp := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&mp)
	return mp["refresh_token"], err
}
