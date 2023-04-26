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
	"time"

	"github.com/TypicalAM/gopoker/config"
	"github.com/TypicalAM/gopoker/models"
	"github.com/TypicalAM/gopoker/routes"
	"github.com/TypicalAM/gopoker/services/game"
	"github.com/TypicalAM/gopoker/services/upload"
	"github.com/TypicalAM/gopoker/texas"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var tdb *gorm.DB
var trouter *gin.Engine

// A time offset to make sure that the connection is established
// and the hub creates the game/clients
var wsConnectTime = 500 * time.Millisecond

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

	router, err := routes.New(db, cfg, uploader)
	if err != nil {
		return err
	}

	trouter = router
	return nil
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatal("error setting up tests:", err)
	}

	code := m.Run()

	if err := teardown(); err != nil {
		log.Fatal("error tearing down tests:", err)
	}

	os.Exit(code)
}

type userWS struct {
	username string
	conn     *websocket.Conn
}

type queueResponse struct {
	UUID string `json:"uuid"`
}

// createConnect creates three users for testing, logs them in, and connects them
// to the game server
func createConnect(t *testing.T) ([]userWS, *httptest.Server) {
	t.Helper()

	users := make([]userWS, 3)
	server := httptest.NewServer(trouter)
	rawURL, _ := url.ParseRequestURI(server.URL)

	for i := 0; i < 3; i++ {
		userpass, _ := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("testpass%d", i)), bcrypt.DefaultCost)
		user := models.User{
			Username: fmt.Sprintf("user%d", i),
			Password: string(userpass),
		}

		if res := tdb.Save(&user); res.Error != nil {
			t.Fatalf("error creating user: %s", res.Error)
		}

		body := fmt.Sprintf(`{"username": "user%d", "password": "testpass%d"}`, i, i)
		req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(body)))
		if err != nil {
			t.Fatalf("error creating login request: %s", err)
		}

		rr := httptest.NewRecorder()
		trouter.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("error logging in user: %s", rr.Body.String())
		}

		req, err = http.NewRequest("POST", "/api/game/queue", nil)
		if err != nil {
			t.Fatalf("error creating queue request: %s", err)
		}
		req.AddCookie(rr.Result().Cookies()[0])

		rr = httptest.NewRecorder()
		trouter.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("error queuing user: %s", rr.Body.String())
		}

		var queueRes queueResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &queueRes); err != nil {
			t.Fatalf("error unmarshalling queue response: %s", err)
		}

		wsURL := "ws" + server.URL[4:] + "/api/game/id/" + queueRes.UUID
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(rawURL, []*http.Cookie{rr.Result().Cookies()[0]})
		dialer := websocket.DefaultDialer
		dialer.Jar = jar
		ws, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Errorf("error dialing websocket: %s", err)
		}

		users[i] = userWS{
			username: user.Username,
			conn:     ws,
		}
	}

	time.Sleep(wsConnectTime)
	return users, server
}

// sendMessage sends a message to the websocket
func sendMessage(t *testing.T, user userWS, msg game.GameMessage) {
	t.Helper()

	m, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("error marshalling message: %s", err)
	}

	if err := user.conn.WriteMessage(websocket.TextMessage, m); err != nil {
		t.Fatalf("error writing message: %s", err)
	}
}

// readMessage reads a message from the websocket
func readMessage(t *testing.T, user userWS) game.GameMessage {
	t.Helper()

	_, m, err := user.conn.ReadMessage()
	if err != nil {
		t.Fatalf("error reading message: %s", err)
	}

	var msg game.GameMessage
	if err := json.Unmarshal(m, &msg); err != nil {
		t.Fatalf("error unmarshalling message: %s", err)
	}

	return msg
}

// Make sure that the users can connect to the game server and that the game starts afterwards
func TestGameConnect(t *testing.T) {
	users, server := createConnect(t)
	defer func() {
		server.Close()
		for _, user := range users {
			user.conn.Close()
		}
	}()

	var user models.User
	if res := tdb.First(&user, "username = ?", users[0].username); res.Error != nil {
		t.Fatalf("error finding user: %s", res.Error)
	}

	var gameModel models.Game
	if res := tdb.First(&gameModel, "id = ?", user.GameID); res.Error != nil {
		t.Fatalf("error finding game: %s", res.Error)
	}

	if gameModel.Playing != true {
		t.Fatalf("game not playing")
	}
}

// TestBroadcast tests that the game server broadcasts messages to all users
func TestBroadcast(t *testing.T) {
	users, server := createConnect(t)
	defer func() {
		server.Close()
		for _, user := range users {
			user.conn.Close()
		}
	}()

	for _, user := range users {
		msg := readMessage(t, user)
		if msg.Type != game.MsgState {
			t.Fatalf("expected state message, got %s", msg.Type)
		}

		var state texas.TexasHoldEm
		if err := json.Unmarshal([]byte(msg.Data), &state); err != nil {
			t.Fatalf("error unmarshalling state: %s", err)
		}
	}
}

// TestExampleErrors tests that the game server broadcasts messages to all users
func TestExampleErrors(t *testing.T) {
	users, server := createConnect(t)
	defer func() {
		server.Close()
		for _, user := range users {
			user.conn.Close()
		}
	}()

	tt := []struct {
		name      string
		userIndex int
		msg       game.GameMessage
	}{
		{
			name:      "invalid action",
			userIndex: 0,
			msg: game.GameMessage{
				Type: game.MsgAction,
				Data: "Invalid",
			},
		},
		{
			name:      "wrong type",
			userIndex: 0,
			msg: game.GameMessage{
				Type: "typetype!!!",
				Data: texas.Fold,
			},
		},
	}

	for _, tc := range tt {
		sendMessage(t, users[tc.userIndex], tc.msg)

		var msg game.GameMessage
		drop := true
		for drop {
			msg = readMessage(t, users[tc.userIndex])
			if msg.Type != game.MsgState {
				drop = false
			}
		}

		if msg.Type != game.MsgError {
			t.Fatalf("expected error message, got %s", msg.Type)
		}
	}
}

// teardown deletes the test users
func teardown() error {
	return tdb.Delete(&models.User{}, "username LIKE ?", "user%").Error
}
	
