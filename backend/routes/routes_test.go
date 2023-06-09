package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/services/upload"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var tdb *gorm.DB
var trouter *gin.Engine

// setup sets up the tests
func setup() error {
	gin.SetMode(gin.TestMode)
	cfg := config.New()

	db, err := models.New(cfg)
	if err != nil {
		return err
	}

	tdb = db

	if err = models.Migrate(db); err != nil {
		return err
	}

	uploader, err := upload.NewCloudinary(cfg.CloudinaryURL, "test", 5*time.Second)
	if err != nil {
		return err
	}

	router, err := New(db, cfg, uploader)
	if err != nil {
		return err
	}

	trouter = router

	return nil
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatalf("error setting up tests: %s", err)
	}

	code := m.Run()

	if err := teardown(); err != nil {
		log.Fatalf("error tearing down tests: %s", err)
	}

	os.Exit(code)
}

func TestRegister(t *testing.T) {
	tt := []struct {
		name string
		body string
		code int
	}{
		{
			name: "password too short",
			body: `{"username":"test3","password":"test"}`,
			code: http.StatusBadRequest,
		},
		{
			name: "normal register",
			body: `{"username":"test4","password":"testtest"}`,
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
			trouter.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.code)
			}
		})
	}
}

// createTestUsers creates test users
func createTestUsers() error {
	user1pass, _ := bcrypt.GenerateFromPassword([]byte("testpass1"), bcrypt.DefaultCost)
	user1 := models.User{
		Username: "user1",
		Password: string(user1pass),
	}

	if res := tdb.Save(&user1); res.Error != nil {
		return res.Error
	}

	user2pass, _ := bcrypt.GenerateFromPassword([]byte("testpass2"), bcrypt.DefaultCost)
	user2 := models.User{
		Username: "user2",
		Password: string(user2pass),
	}

	if res := tdb.Save(&user2); res.Error != nil {
		return res.Error
	}

	return nil
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
			trouter.ServeHTTP(rr, req)

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

func logInUser(body string) (error, *http.Cookie) {
	if err := createTestUsers(); err != nil {
		return err, nil
	}

	req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err, nil
	}

	rr := httptest.NewRecorder()
	trouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		return fmt.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK), nil
	}

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		return fmt.Errorf("handler returned wrong number of cookies: got %v want %v", len(cookies), 1), nil
	}

	return nil, cookies[0]
}

func TestLogout(t *testing.T) {
	tt := []struct {
		name string
		body string
		code int
	}{
		{
			name: "normal logout",
			body: `{"username":"user1","password":"testpass1"}`,
			code: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err, cookie := logInUser(tc.body)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/api/logout", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.AddCookie(cookie)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			trouter.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.code)
			}
		})
	}
}

var QueueResponse struct {
	UUID string `json:"uuid"`
}

func TestQueue(t *testing.T) {
	err, cookie := logInUser(`{"username":"user1","password":"testpass1"}`)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/game/queue", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	trouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if err := json.NewDecoder(rr.Body).Decode(&QueueResponse); err != nil {
		t.Fatal(err)
	}

	var game models.Game
	tdb.Model(&models.Game{}).Preload("Players").Where("uuid = ?", QueueResponse.UUID).First(&game)
	if len(game.Players) != 1 {
		t.Errorf("handler returned wrong number of players: got %v want %v", len(game.Players), 1)
	}

	if game.Players[0].Username != "user1" {
		t.Errorf("handler returned wrong player: got %v want %v", game.Players[0].Username, "user1")
	}
}

// teardown removes the test users from the database
func teardown() error {
	return tdb.Where("username = ?", "user1").Or("username = ?", "user2").Delete(&models.User{}).Error
}
