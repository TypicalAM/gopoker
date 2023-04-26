package game

import "sync"

// gameStore is a store of games.
type gameStore struct {
	games map[string]*lobby
	mutex sync.RWMutex
}

// newGameStore creates a new game store.
func newGameStore() gameStore {
	return gameStore{
		games: make(map[string]*lobby),
	}
}

// load gets a game from the store.
func (s *gameStore) load(UUID string) (*lobby, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	game, ok := s.games[UUID]
	return game, ok
}

// save saves a game in the store.
func (s *gameStore) save(l *lobby) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.games[l.uuid] = l
}

// delete deletes a game from the store.
func (s *gameStore) delete(UUID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.games, UUID)
}
