package websockets

import "sync"

// gameStore is a store of games.
type gameStore struct {
	games map[string]*Game
	mutex sync.RWMutex
}

// newGameStore creates a new game store.
func newGameStore() gameStore {
	return gameStore{
		games: make(map[string]*Game),
	}
}

// load gets a game from the store.
func (s *gameStore) load(UUID string) (*Game, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	game, ok := s.games[UUID]
	return game, ok
}

// loadAll gets all the games from the store.
func (s *gameStore) loadAll() []*Game {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	games := make([]*Game, 0, len(s.games))
	for _, game := range s.games {
		games = append(games, game)
	}

	return games
}

// save saves a game in the store.
func (s *gameStore) save(game *Game) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.games[game.UUID] = game
}

// delete deletes a game from the store.
func (s *gameStore) delete(UUID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.games, UUID)
}
