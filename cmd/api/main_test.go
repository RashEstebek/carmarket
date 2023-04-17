package main

import (
	"bytes"
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/jsonlog"
	"carMarket.dreamteam.kz/internal/mailer"
	"carMarket.dreamteam.kz/internal/validator"
	"context"
	"encoding/json"
	"flag"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func GetApp() application {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:12345@localhost/carmarket?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle time")
	// flag.StringVar(&cfg.db.maxLifetime, "db-max-lifetime", "1h", "PostgreSQL max idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp-mail.outlook.com", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	// use your own credentials here as username and password

	flag.StringVar(&cfg.smtp.username, "smtp-username", "211542@astanait.edu.kz", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "Ktl2021!", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "211542@astanait.edu.kz", "SMTP sender")

	flag.Parse()

	// Using new json oriented logger
	logger := jsonlog.NewToActivate(os.Stdout, jsonlog.LevelInfo)

	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil) // calling PrintFatal function if there is an error with db server connection
	}
	// db will be closed before main function is completed.
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil) // printing custom info if db server connection is established

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // data.NewModels() function to initialize a Models struct
		// Initialize a new Mailer instance using the settings from the command line
		// flags, and add it to the application struct.
		mailer: mailer.NewToActivate(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	return *app
}

func TestValidateUser(t *testing.T) {
	// Test invalid user
	user := &data.User{
		Name:      "",
		Email:     "invalidemail",
		Role:      "user",
		Activated: false,
		Cash:      -100000,
	}
	user.Password.Set("short")

	v := validator.NewToActivate()
	data.ValidateUser(v, user)
	if v.Valid() {
		t.Errorf("unexpectedly valid user: %v", user)
	} else {
		t.Logf("validation errors: %v", v.Errors)
	}

	// Test valid user
	user = &data.User{
		Name:      "Valid User",
		Email:     "validuser@example.com",
		Role:      "admin",
		Activated: true,
		Cash:      100000,
	}
	user.Password.Set("validpassword")

	v = validator.NewToActivate()
	data.ValidateUser(v, user)
	if !v.Valid() {
		t.Errorf("unexpected validation errors: %v", v.Errors)
	}
}

func TestValidatePermittedValue(t *testing.T) {
	value1 := "s"
	value2 := "small"

	permittedValues := []string{"s", "m", "l"}

	v := validator.NewToActivate()
	v.Check(validator.PermittedValue(value1, permittedValues...), "size", "invalid size")
	if !v.Valid() {
		t.Errorf("%v", v.Errors)
	}

	v1 := validator.NewToActivate()
	v1.Check(validator.PermittedValue(value2, permittedValues...), "size", "invalid size")
	if v1.Valid() {
		t.Errorf("Have to get error, unexpected result")
	}
}

func TestReadJSON(t *testing.T) {
	app := GetApp()
	var input struct {
		Name string `json:"name"`
	}

	jsonBody := []byte(`{"name": "Rash"}`)
	bodyReader := bytes.NewReader(jsonBody)
	w := new(http.ResponseWriter)
	r, _ := http.NewRequest(http.MethodPost, "http://localhost:4000/v1/cars/1", bodyReader)
	err := app.readJSON(*w, r, &input)
	if err != nil {
		t.Errorf("Got an unexpected error: %v", err)
	}
}

func TestReadCSV(t *testing.T) {
	app := GetApp()
	url1, _ := url.Parse("http://localhost:4000/v1/cars")
	urlValues := url1.Query()
	sizes := app.readCSV(urlValues, "sizes", []string{})

	if len(sizes) != 3 {
		t.Errorf("Expected array with size 3, but got %d", len(sizes))
	}
	for i := 0; i < len(sizes); i++ {
		if sizes[i] != "xs" && sizes[i] != "s" && sizes[i] != "m" {
			t.Errorf("Unexpected value %v", sizes[i])
		}
	}
}

func TestReadString(t *testing.T) {
	app := GetApp()
	url1, err := url.Parse("http://localhost:4000/v1/cars")
	urlValues := url1.Query()
	key := app.readString(urlValues, "key", "")
	if key != "bmw" || err != nil {
		t.Errorf("Expected value is bmw, but got %v", key)
	}
}

func TestReadInt(t *testing.T) {
	app := GetApp()
	r, _ := url.Parse("http://localhost:4000/v1/cars")
	v := validator.NewToActivate()
	qs := r.Query()
	price := app.readInt(qs, "price", 200000, v)

	if price != 35000 {
		t.Errorf("Expected value is 35000, but got %d", price)
	}
}

func TestReadID(t *testing.T) {
	app := GetApp()
	req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)

	params := httprouter.Params{
		{Key: "id", Value: "1"},
	}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	req = req.WithContext(ctx)
	id, err := app.readIDParam(req)

	if id != 1 || err != nil {
		t.Errorf("Expected 1, got %v", id)
	}
}

func TestGetAll(t *testing.T) {
	app := GetApp()
	req, err := http.NewRequest("GET", "/v1/cars", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.listCarsHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var arr []data.Car
	var arr1 []data.Car

	response := rr.Body.String()

	err = json.Unmarshal([]byte(response), &arr)
	if err != nil {
		t.Errorf("Can't marshall response to clothe type")
	}

	car := &data.Car{
		ID:    1,
		Price: 150000,
		Model: "new",
		Color: "test",
		Marka: "toyota",
		Type:  "",
		Image: "",
	}

	app.models.Cars.Insert(car)

	rr1 := httptest.NewRecorder()

	handler.ServeHTTP(rr1, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	secondResponse := rr1.Body.String()
	err = json.Unmarshal([]byte(secondResponse), &arr1)
	if err != nil {
		t.Errorf("Can't marshall response to car type")
	}

	if len(arr1)-len(arr) != 1 {
		t.Errorf("Data length before insertion %v, after %v", len(arr), len(arr1))
	}
}

func (app *application) TestAuthorization(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)
	req.Header.Set("Authorization", "John tests")

	w := httptest.NewRecorder()

	handler := app.authenticate(http.HandlerFunc(app.listCarsHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected to get status code 401, but got %v", w.Code)
	}

}

func (app *application) TestRequireRole(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)
	req.Header.Set("Authorization", "John tests")

	user := &data.User{
		ID:        1,
		Name:      "admin",
		Email:     "test@gmail.com",
		Activated: true,
		Version:   1,
	}

	req = app.contextSetUser(req, user)

	w := httptest.NewRecorder()

	handler := app.requireAdminRole(http.HandlerFunc(app.listMarkasHandler))
	handler.ServeHTTP(w, req)

	println(w.Body.String())
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected to get status code 403, but got %v", w.Code)
	}
}
