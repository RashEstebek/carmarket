package main

import (
	"bytes"
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/jsonlog"
	"carMarket.dreamteam.kz/internal/mailer"
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

var getAppInstance application

func init() {
	getAppInstance = GetApp()
}

func GetApp() application {
	var cfg config
	cfg.port = 4000
	cfg.env = "development"
	cfg.db.dsn = "postgres://postgres:12345@localhost/carmarket?sslmode=disable"

	cfg.db.maxOpenConns = 25
	cfg.db.maxIdleConns = 25
	cfg.db.maxIdleTime = "15m"

	cfg.smtp.host = "sandbox.smtp.mailtrap.io"
	cfg.smtp.port = 2525
	cfg.smtp.username = "211428@astanait.edu.kz"
	cfg.smtp.password = "Rik_Aitu2024*"
	cfg.smtp.sender = "211428@astanait.edu.kz"

	logger := jsonlog.NewToActivate(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.NewToActivate(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	return *app
}

func TestListCarsHandlerIntegration(t *testing.T) {
	req1, err := http.NewRequest("GET", "/v1/cars", nil)
	req1.Header.Add("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	ht1 := httptest.NewRecorder()
	handler1 := http.HandlerFunc(getAppInstance.listCarsHandler)
	handler1.ServeHTTP(ht1, req1)

	if status := ht1.Code; status != http.StatusOK {
		t.Errorf("Handler function returned unexpected status code: we got %v, but want %v",
			status, http.StatusOK)
	}

	var arr1 []data.Car
	res1 := ht1.Body.String()
	err = json.Unmarshal([]byte(res1), &arr1)

	if err != nil {
		t.Errorf("Unmarshall can't marshall response to car type")
	}
}

func TestShowCarHandlerIntegration(t *testing.T) {
	// Create a new request to the show car endpoint with an ID that doesn't exist in the database.
	id := int64(12345)
	url := fmt.Sprintf("/v1/cars/%d", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the params to the request context.
	params := httprouter.Params{
		{Key: "id", Value: strconv.FormatInt(id, 10)},
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)

	// Create a new recorder and HTTP handler for the request.
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(getAppInstance.showCarHandler)
	handler.ServeHTTP(recorder, req)

	// Verify that the response status code is HTTP 404 Not Found.
	if status := recorder.Code; status != http.StatusNotFound {
		t.Errorf("Handler function returned unexpected status code: got %v, want %v",
			status, http.StatusNotFound)
	}
}

func TestDeleteCarHandlerIntegration(t *testing.T) {
	// Use a nonexistent ID to ensure the handler returns a 404 response.
	id := int64(12345)

	// Create a new request to the delete car endpoint.
	url := fmt.Sprintf("/v1/cars/%d", id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the params to the request context.
	params := httprouter.Params{
		{Key: "id", Value: strconv.FormatInt(id, 10)},
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)

	// Create a new recorder and HTTP handler for the request.
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(getAppInstance.deleteCarHandler)
	handler.ServeHTTP(recorder, req)

	// Verify that the response status code is HTTP 404 Not Found.
	if status := recorder.Code; status != http.StatusNotFound {
		t.Errorf("Handler function returned unexpected status code: got %v, want %v",
			status, http.StatusNotFound)
	}
}

func TestUpdateCarHandlerIntegration(t *testing.T) {
	id := int64(12345)
	car := `{
		"model":     "T51",
		"year":    2001,
		"price":    1250000,
		"marka":    "audi",
		"color":    "black",
		"type":      "sedan",
		"image":     "audi.png",
		"description": "very good car"
	}`

	byteArr := []byte(car)
	bodyReader := bytes.NewReader(byteArr)

	req, err := http.NewRequest(http.MethodPut, "/v1/cars", bodyReader)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getAppInstance.updateCarHandler)

	params := httprouter.Params{
		{Key: "id", Value: strconv.FormatInt(id, 10)},
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler function returned unexpected status code: got %v, want %v",
			status, http.StatusNotFound)
	}
}

func TestCreateMarkaHandlerIntegration1(t *testing.T) {
	marka := `{
		"name": "Toyota", 
		"producer": "Japan", 
		"logo": "https://example.com/toyota.png"
	}`

	byteArr := []byte(marka)

	bodyReader := bytes.NewReader(byteArr)

	req, err := http.NewRequest(http.MethodPost, "/v1/cars", bodyReader)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getAppInstance.createMarkaHandler)

	handler.ServeHTTP(rr, req)

	type responseID struct {
		ID int64 `json:"id"`
	}

	data := responseID{}
	err = json.Unmarshal([]byte(rr.Body.String()), &data)
	if err != nil {
		t.Errorf("Unmarshall can't marshall response to car type")
	}

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func TestAuthorizationIntegration(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)
	req.Header.Set("Authorization", "Bearer test")

	w := httptest.NewRecorder()

	handler := getAppInstance.authenticate(http.HandlerFunc(getAppInstance.listCarsHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("We expected the status code %v, but received %v", http.StatusUnauthorized, w.Code)
	}

	token, err := getAppInstance.models.Tokens.New(1, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		t.Errorf("%v", err)
	}

	bearerToken := "Bearer " + token.Plaintext
	req.Header.Set("Authorization", bearerToken)

	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req)

	if w1.Code != http.StatusOK {
		t.Errorf("We expected the status code %v, but received %v", http.StatusOK, w1.Code)
	}

}

func TestRequireAdminRoleIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)
	req.Header.Set("Authorization", "Bearer test")

	user := &data.User{
		ID:        1,
		Activated: true,
	}

	req = getAppInstance.contextSetUser(req, user)
	w := httptest.NewRecorder()

	handler := getAppInstance.requireAdminRole(getAppInstance.listCarsHandler)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("We expected the status code %v, but received %v", http.StatusOK, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/cars", nil)
	req.Header.Set("Authorization", "Bearer test")

	user = &data.User{
		ID:        2,
		Activated: false,
	}

	req = getAppInstance.contextSetUser(req, user)

	w1 := httptest.NewRecorder()

	handler = getAppInstance.requireAdminRole(getAppInstance.listMarkasHandler)
	handler.ServeHTTP(w1, req)

	if w1.Code != http.StatusForbidden {
		t.Errorf("We expected the status code %v, but received %v", http.StatusForbidden, w1.Code)
	}
}
