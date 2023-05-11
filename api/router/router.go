package router

import (
	"net/http"

	"github.com/go-chi/chi"

	"spse-role-poc/api/manager"
)

func New() http.Handler {
	r := chi.NewRouter()

	// publicly accessible - to test the api is responding
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello World!"}`))
	})

	// user functions
	r.Post("/create", manager.CreateUserHandler)
	// r.Patch("/addroles", manager.AddRolesHandler)

	return r
}
