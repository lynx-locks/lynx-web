package routes

import (
	"api/controller"
	"github.com/go-chi/chi/v5"
)

func CredentialsRouter() chi.Router {
	r := chi.NewRouter()
	r.Delete("/session", controller.DeleteSessionData)
	return r
}
