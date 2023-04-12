package game_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	os.Exit(m.Run())
}

type userCookie struct {
	username string
	cookie   *http.Cookie
}

type queueResponse struct {
	UUID string `json:"uuid"`
}

// createUsers creates three users for testing, logs them in, and returns their
// cookies.
func createUsers() (error, []userCookie, string) {
	uuid := ""
	users := make([]userCookie, 3)

	for i := 0; i < 3; i++ {
		userpass, _ := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("testpass%d", i)), bcrypt.DefaultCost)
		user := models.User{
			GameID:   1,
			Username: fmt.Sprintf("user%d", i),
			Password: string(userpass),
		}

		if res := testDB.Save(&user); res.Error != nil {
			return res.Error, nil, ""
		}

		body := fmt.Sprintf(`{"username": "user%d", "password": "testpass%d"}`, i, i)
		req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(body)))
		if err != nil {
			return err, nil, ""
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			return err, nil, ""
		}

		req, err = http.NewRequest("POST", "/api/game/queue", nil)
		if err != nil {
			return err, nil, ""
		}
		req.AddCookie(rr.Result().Cookies()[0])

		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			return err, nil, ""
		}

		users[i] = userCookie{
			username: fmt.Sprintf("user%d", i),
			cookie:   rr.Result().Cookies()[0],
		}

		var queueRes queueResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &queueRes); err != nil {
			return err, nil, ""
		}
		uuid = queueRes.UUID
	}

	return nil, users, uuid
}

func TestGameConnect(t *testing.T) {
	err, users, uuid := createUsers()
	if err != nil {
		t.Errorf("error creating users: %s", err)
	}

	var game models.Game
	res := testDB.Model(&models.Game{}).Preload("Players").Where("uuid = ?", uuid).First(&game)
	if res.Error != nil {
		t.Error("error finding game")
	}

	sockets := make([]*websocket.Conn, 3)
	s := httptest.NewServer(router)
	defer s.Close()

	rawURL, _ := url.ParseRequestURI(s.URL)
	wsURL := "ws" + s.URL[4:] + "/api/game/id/" + uuid

	for i, user := range users {
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(rawURL, []*http.Cookie{user.cookie})
		dialer := websocket.DefaultDialer
		dialer.Jar = jar
		ws, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Errorf("error dialing websocket: %s", err)
		}
		sockets[i] = ws
		defer ws.Close()
	}
}
