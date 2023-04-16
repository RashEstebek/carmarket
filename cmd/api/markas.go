package main

import (
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createMarkaHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Producer string `json:"producer"`
		Logo     string `json:"logo,omitempty"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	marka := &data.Marka{
		Name:     input.Name,
		Producer: input.Producer,
		Logo:     input.Logo,
	}
	v := validator.NewToActivate()

	if data.ValidateMarka(v, marka); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Markas.Insert(marka)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/markas/%d", marka.ID))
	err = app.writeJSON(w, http.StatusCreated, marka, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showMarkaHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	marka, err := app.models.Markas.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, marka, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMarkaHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	marka, err := app.models.Markas.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Name     *string `json:"name"`
		Producer *string `json:"producer"`
		Logo     *string `json:"logo,omitempty"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		marka.Name = *input.Name
	}
	if input.Producer != nil {
		marka.Producer = *input.Producer
	}
	if input.Logo != nil {
		marka.Logo = *input.Logo
	}

	v := validator.NewToActivate()
	if data.ValidateMarka(v, marka); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Markas.Update(marka)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, marka, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMarkaHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Markas.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "marka successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMarkasHandler(w http.ResponseWriter, r *http.Request) {
	markas, err := app.models.Markas.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, markas, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
