package routes_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var controller routes.Controller
var router *gin.Engine

// setup sets up the tests
func setup() {
	gin.SetMode(gin.TestMode)
	cfg, err := config.ReadConfig("../")
	if err != nil {
		os.Exit(1)
	}

	db, err := models.ConnectToTestDatabase(cfg)
	if err != nil {
		os.Exit(1)
	}

	err = models.MigrateDatabase(db)
	if err != nil {
		os.Exit(1)
	}

	testDB = db

	controller = routes.New(db, nil, cfg)
	engine, err := routes.SetupRouter(db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	router = engine
}

// setupDatabase sets up the test database connection
func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

// TestRegister tests the Register route
func TestRegister(t *testing.T) {
	tt := []struct {
		name string
		body string
		code int
	}{
		{
			name: "password too short",
			body: `{"username":"test2","password":"test"}`,
			code: http.StatusBadRequest,
		},
		{
			name: "normal register",
			body: `{"username":"test","password":"testtest"}`,
			code: http.StatusCreated,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer([]byte(tc.body)))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.code)
			}
		})
	}
}

// teardown clears the tables of the database permanently
func teardown() {
	ses := &gorm.Session{AllowGlobalUpdate: true}
	testDB.Session(ses).Unscoped().Delete(&models.User{})
	testDB.Session(ses).Unscoped().Delete(&models.Session{})
	testDB.Session(ses).Unscoped().Delete(&models.Game{})
}
