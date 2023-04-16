package main

import (
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/validator"
	"errors"
	"net/http"
	"strings"
)

func (app *application) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	car, err := app.models.Cars.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	authorizationHeader := r.Header.Get("Authorization")

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	token := headerParts[1]
	v := validator.NewToActivate()
	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidAuthenticationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.models.Users.UpdateCash(user, car.Price)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Carts.AddToCart(user.ID, *car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "car successfully added to the cart"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
