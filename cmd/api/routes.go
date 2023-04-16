package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/cars", app.listCarsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/cars", app.requireAdminRole(app.createCarHandler))
	router.HandlerFunc(http.MethodGet, "/v1/cars/:id", app.requireActivated(app.showCarHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/cars/:id", app.requireAdminRole(app.updateCarHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/cars/:id", app.requireAdminRole(app.deleteCarHandler))

	router.HandlerFunc(http.MethodGet, "/v1/markas", app.listMarkasHandler)
	router.HandlerFunc(http.MethodPost, "/v1/markas", app.requireAdminRole(app.createMarkaHandler))
	router.HandlerFunc(http.MethodGet, "/v1/markas/:id", app.requireActivated(app.showMarkaHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/markas/:id", app.requireAdminRole(app.updateMarkaHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/markas/:id", app.requireAdminRole(app.deleteMarkaHandler))

	router.HandlerFunc(http.MethodPut, "/v1/tocart/:id", app.requireActivated(app.addToCartHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
