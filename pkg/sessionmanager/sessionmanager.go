package sessionmanager

import (
	"github.com/corporateanon/my1562bot/pkg/models"
	"github.com/jinzhu/gorm"
)

type SessionManager struct {
	db *gorm.DB
}

type Session struct {
	chatID       int64
	sessionModel models.Session
	updates      map[string]interface{}
	db           *gorm.DB
}

func NewSessionManager(db *gorm.DB) *SessionManager {
	return &SessionManager{db}
}

func (sm *SessionManager) NewSession(chatID int64) *Session {
	sess := models.Session{ChatID: chatID}
	sm.db.Where(models.Session{ChatID: chatID}).FirstOrCreate(&sess)
	updates := make(map[string]interface{})
	return &Session{
		chatID:       chatID,
		sessionModel: sess,
		updates:      updates,
		db:           sm.db,
	}
}

func (s *Session) GetPhase() models.Phase {
	return s.sessionModel.Phase
}

func (s *Session) GetStreetID() int {
	return s.sessionModel.StreetID
}

func (s *Session) SetPhase(phase models.Phase) {
	s.updates["Phase"] = phase
}

func (s *Session) SetStreetID(StreetID int) {
	s.updates["StreetID"] = StreetID
}

func (s *Session) Save() {
	s.db.Model(s.sessionModel).Updates(s.updates)
}
