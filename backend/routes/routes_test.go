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
	"golang.org/x/crypto/bcrypt"
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

func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

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

func TestLogin(t *testing.T) {
	if err := createTestUsers(); err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name string
		body string
		code int
	}{
		{
			name: "normal login 1",
			body: `{"username":"user1","password":"testpass1"}`,
			code: http.StatusOK,
		},
		{
			name: "normal login 2",
			body: `{"username":"user2","password":"testpass2"}`,
			code: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(tc.body)))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.code)
			}

			cookies := rr.Result().Cookies()
			if len(cookies) != 1 {
				t.Errorf("handler returned wrong number of cookies: got %v want %v", len(cookies), 1)
			}

			found := false
			for _, cookie := range cookies {
				if cookie.Name == "gopoker_session" {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("handler did not return the correct cookie")
			}
		})
	}
}

// createTestUsers creates test users
func createTestUsers() error {
	user1pass, _ := bcrypt.GenerateFromPassword([]byte("testpass1"), bcrypt.DefaultCost)
	user1 := models.User{
		GameID:   1,
		Username: "user1",
		Password: string(user1pass),
	}

	if res := testDB.Save(&user1); res.Error != nil {
		return res.Error
	}

	user2pass, _ := bcrypt.GenerateFromPassword([]byte("testpass2"), bcrypt.DefaultCost)
	user2 := models.User{
		GameID:   1,
		Username: "user2",
		Password: string(user2pass),
	}

	if res := testDB.Save(&user2); res.Error != nil {
		return res.Error
	}

	return nil
}

// teardown clears the tables of the database permanently
func teardown() {
	ses := &gorm.Session{AllowGlobalUpdate: true}
	testDB.Session(ses).Unscoped().Delete(&models.Session{})
	testDB.Session(ses).Unscoped().Delete(&models.User{})
	testDB.Session(ses).Unscoped().Delete(&models.Game{})
}
