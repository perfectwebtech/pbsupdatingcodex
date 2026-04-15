package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LiveChatService manages real-time chat for live streams
type LiveChatService struct {
	db        *sql.DB
	rooms     map[string]*ChatRoom
	roomsMu   sync.RWMutex
	bannedKws []string
}

// NewLiveChatService creates a new live chat service
func NewLiveChatService(db *sql.DB) *LiveChatService {
	return &LiveChatService{
		db:    db,
		rooms: make(map[string]*ChatRoom),
		bannedKws: []string{
			"spam", "scam", "click here", "free money",
			// Add more banned keywords
		},
	}
}

// =====================================================
// ERRORS
// =====================================================

var (
	ErrChatRoomNotFound = errors.New("chat room not found")
	ErrUserBanned       = errors.New("user is banned from chat")
	ErrMessageTooLong   = errors.New("message exceeds maximum length")
	ErrRateLimitExceeded = errors.New("message rate limit exceeded")
	ErrInappropriateContent = errors.New("message contains inappropriate content")
)

// =====================================================
// MODELS
// =====================================================

// ChatRoom represents a live chat room for a stream
type ChatRoom struct {
	ID            string                 `json:"id"`
	StreamID      string                 `json:"stream_id"`
	Name          string                 `json:"name"`
	IsActive      bool                   `json:"is_active"`
	IsModerated   bool                   `json:"is_moderated"`
	SlowMode      int                    `json:"slow_mode_seconds"`
	MaxMessageLen int                    `json:"max_message_length"`
	Members       map[string]*ChatMember `json:"-"`
	MessageCount  int64                  `json:"message_count"`
	ActiveUsers   int                    `json:"active_users"`
	CreatedAt     time.Time              `json:"created_at"`
	mu            sync.RWMutex           `json:"-"`
}

// ChatMember represents a user in a chat room
type ChatMember struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	AvatarURL    string    `json:"avatar_url"`
	Role         string    `json:"role"` // viewer, moderator, vip, owner
	JoinedAt     time.Time `json:"joined_at"`
	LastMessageAt time.Time `json:"last_message_at"`
	IsMuted      bool      `json:"is_muted"`
	MuteUntil    time.Time `json:"mute_until"`
	BadgeIcons   []string  `json:"badge_icons"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID         string    `json:"id"`
	RoomID     string    `json:"room_id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	AvatarURL  string    `json:"avatar_url"`
	Message    string    `json:"message"`
	Type       string    `json:"type"` // text, emoji, gif, super_chat, system
	Color      string    `json:"color,omitempty"`
	BadgeIcons []string  `json:"badge_icons,omitempty"`
	SuperChatAmount float64 `json:"super_chat_amount,omitempty"`
	Mentions   []string  `json:"mentions,omitempty"`
	Reactions  map[string]int `json:"reactions,omitempty"`
	IsDeleted  bool      `json:"is_deleted"`
	CreatedAt  time.Time `json:"created_at"`
}

// Reaction represents an emoji reaction during a stream
type Reaction struct {
	ID        string    `json:"id"`
	StreamID  string    `json:"stream_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"` // 👍 ❤️ 🔥 😂 😮 etc.
	Timestamp int64     `json:"timestamp"` // Stream position in ms
	CreatedAt time.Time `json:"created_at"`
}

// Poll represents a live poll during a stream
type Poll struct {
	ID         string       `json:"id"`
	StreamID   string       `json:"stream_id"`
	Question   string       `json:"question"`
	Options    []PollOption `json:"options"`
	Duration   int          `json:"duration_seconds"`
	IsActive   bool         `json:"is_active"`
	TotalVotes int          `json:"total_votes"`
	StartedAt  time.Time    `json:"started_at"`
	EndsAt     time.Time    `json:"ends_at"`
}

// PollOption represents a poll choice
type PollOption struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Votes int    `json:"votes"`
}

// =====================================================
// CHAT ROOM MANAGEMENT
// =====================================================

// CreateChatRoom creates a new chat room for a stream
func (s *LiveChatService) CreateChatRoom(ctx context.Context, streamID, name string) (*ChatRoom, error) {
	room := &ChatRoom{
		ID:            uuid.New().String(),
		StreamID:      streamID,
		Name:          name,
		IsActive:      true,
		IsModerated:   true,
		SlowMode:      0,
		MaxMessageLen: 500,
		Members:       make(map[string]*ChatMember),
		CreatedAt:     time.Now(),
	}

	query := `
		INSERT INTO chat_rooms (id, stream_id, name, is_active, is_moderated, slow_mode, max_message_length, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		room.ID, room.StreamID, room.Name, room.IsActive, room.IsModerated,
		room.SlowMode, room.MaxMessageLen,
	)
	if err != nil {
		return nil, err
	}

	s.roomsMu.Lock()
	s.rooms[room.ID] = room
	s.roomsMu.Unlock()

	return room, nil
}

// JoinChatRoom adds a user to a chat room
func (s *LiveChatService) JoinChatRoom(ctx context.Context, roomID, userID, username, avatarURL, role string) error {
	s.roomsMu.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMu.RUnlock()

	if !exists {
		// Try loading from DB
		var err error
		room, err = s.loadChatRoom(ctx, roomID)
		if err != nil {
			return ErrChatRoomNotFound
		}
		s.roomsMu.Lock()
		s.rooms[roomID] = room
		s.roomsMu.Unlock()
	}

	// Check if user is banned
	if banned, _ := s.isUserBanned(ctx, roomID, userID); banned {
		return ErrUserBanned
	}

	member := &ChatMember{
		UserID:    userID,
		Username:  username,
		AvatarURL: avatarURL,
		Role:      role,
		JoinedAt:  time.Now(),
		BadgeIcons: s.getUserBadges(ctx, userID),
	}

	room.mu.Lock()
	room.Members[userID] = member
	room.ActiveUsers = len(room.Members)
	room.mu.Unlock()

	return nil
}

// LeaveChatRoom removes a user from a chat room
func (s *LiveChatService) LeaveChatRoom(roomID, userID string) {
	s.roomsMu.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.Members, userID)
	room.ActiveUsers = len(room.Members)
	room.mu.Unlock()
}

// =====================================================
// MESSAGING
// =====================================================

// SendMessage sends a message to a chat room
func (s *LiveChatService) SendMessage(ctx context.Context, roomID, userID, message, msgType string) (*ChatMessage, error) {
	s.roomsMu.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMu.RUnlock()

	if !exists {
		return nil, ErrChatRoomNotFound
	}

	// Validate message length
	if len(message) > room.MaxMessageLen {
		return nil, ErrMessageTooLong
	}

	// Check if user is in room
	room.mu.RLock()
	member, ok := room.Members[userID]
	room.mu.RUnlock()

	if !ok {
		return nil, errors.New("user not in chat room")
	}

	// Check if user is muted
	if member.IsMuted && time.Now().Before(member.MuteUntil) {
		return nil, errors.New("user is muted")
	}

	// Rate limiting (slow mode)
	if room.SlowMode > 0 {
		if time.Since(member.LastMessageAt).Seconds() < float64(room.SlowMode) {
			return nil, ErrRateLimitExceeded
		}
	}

	// Content moderation
	if s.isInappropriate(message) {
		return nil, ErrInappropriateContent
	}

	// Extract mentions
	mentions := s.extractMentions(message)

	chatMsg := &ChatMessage{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		UserID:     userID,
		Username:   member.Username,
		AvatarURL:  member.AvatarURL,
		Message:    message,
		Type:       msgType,
		BadgeIcons: member.BadgeIcons,
		Mentions:   mentions,
		Reactions:  make(map[string]int),
		CreatedAt:  time.Now(),
	}

	// Save to database
	query := `
		INSERT INTO chat_messages (id, room_id, user_id, message, type, mentions, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	mentionsJSON, _ := json.Marshal(mentions)
	_, err := s.db.ExecContext(ctx, query,
		chatMsg.ID, roomID, userID, message, msgType, string(mentionsJSON),
	)
	if err != nil {
		return nil, err
	}

	// Update last message time
	room.mu.Lock()
	member.LastMessageAt = time.Now()
	room.MessageCount++
	room.mu.Unlock()

	return chatMsg, nil
}

// SendSuperChat sends a paid highlighted message
func (s *LiveChatService) SendSuperChat(ctx context.Context, roomID, userID, message string, amount float64) (*ChatMessage, error) {
	chatMsg, err := s.SendMessage(ctx, roomID, userID, message, "super_chat")
	if err != nil {
		return nil, err
	}

	chatMsg.SuperChatAmount = amount
	chatMsg.Color = s.getSuperChatColor(amount)

	// Update database with super chat info
	query := `UPDATE chat_messages SET super_chat_amount = ?, color = ? WHERE id = ?`
	s.db.ExecContext(ctx, query, amount, chatMsg.Color, chatMsg.ID)

	return chatMsg, nil
}

// GetRecentMessages retrieves recent chat messages
func (s *LiveChatService) GetRecentMessages(ctx context.Context, roomID string, limit int) ([]*ChatMessage, error) {
	query := `
		SELECT m.id, m.room_id, m.user_id, u.username, u.avatar_url,
			   m.message, m.type, COALESCE(m.color, ''), COALESCE(m.super_chat_amount, 0),
			   m.is_deleted, m.created_at
		FROM chat_messages m
		LEFT JOIN users u ON u.id = m.user_id
		WHERE m.room_id = ? AND m.is_deleted = 0
		ORDER BY m.created_at DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		msg := &ChatMessage{}
		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.AvatarURL,
			&msg.Message, &msg.Type, &msg.Color, &msg.SuperChatAmount,
			&msg.IsDeleted, &msg.CreatedAt,
		)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// =====================================================
// REACTIONS (Live emoji reactions on stream)
// =====================================================

// AddReaction adds a real-time reaction during a stream
func (s *LiveChatService) AddReaction(ctx context.Context, streamID, userID, emoji string, timestamp int64) (*Reaction, error) {
	reaction := &Reaction{
		ID:        uuid.New().String(),
		StreamID:  streamID,
		UserID:    userID,
		Emoji:     emoji,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO stream_reactions (id, stream_id, user_id, emoji, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		reaction.ID, streamID, userID, emoji, timestamp,
	)
	return reaction, err
}

// GetStreamReactions returns aggregated reactions for a stream
func (s *LiveChatService) GetStreamReactions(ctx context.Context, streamID string, lastSeconds int) (map[string]int, error) {
	query := `
		SELECT emoji, COUNT(*) as cnt
		FROM stream_reactions
		WHERE stream_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL ? SECOND)
		GROUP BY emoji
	`
	rows, err := s.db.QueryContext(ctx, query, streamID, lastSeconds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make(map[string]int)
	for rows.Next() {
		var emoji string
		var count int
		if err := rows.Scan(&emoji, &count); err != nil {
			continue
		}
		reactions[emoji] = count
	}
	return reactions, nil
}

// =====================================================
// LIVE POLLS
// =====================================================

// CreatePoll creates a new live poll during a stream
func (s *LiveChatService) CreatePoll(ctx context.Context, streamID, question string, options []string, durationSec int) (*Poll, error) {
	poll := &Poll{
		ID:        uuid.New().String(),
		StreamID:  streamID,
		Question:  question,
		Duration:  durationSec,
		IsActive:  true,
		StartedAt: time.Now(),
		EndsAt:    time.Now().Add(time.Duration(durationSec) * time.Second),
		Options:   make([]PollOption, len(options)),
	}

	for i, opt := range options {
		poll.Options[i] = PollOption{
			ID:   uuid.New().String(),
			Text: opt,
		}
	}

	optionsJSON, _ := json.Marshal(poll.Options)
	query := `
		INSERT INTO live_polls (id, stream_id, question, options, duration, is_active, started_at, ends_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		poll.ID, streamID, question, string(optionsJSON), durationSec,
		true, poll.StartedAt, poll.EndsAt,
	)
	return poll, err
}

// VotePoll records a user's vote in a poll
func (s *LiveChatService) VotePoll(ctx context.Context, pollID, optionID, userID string) error {
	// Check if user already voted
	var count int
	checkQuery := `SELECT COUNT(*) FROM poll_votes WHERE poll_id = ? AND user_id = ?`
	s.db.QueryRowContext(ctx, checkQuery, pollID, userID).Scan(&count)
	if count > 0 {
		return errors.New("user already voted")
	}

	query := `
		INSERT INTO poll_votes (id, poll_id, option_id, user_id, voted_at)
		VALUES (?, ?, ?, ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query, uuid.New().String(), pollID, optionID, userID)
	return err
}

// =====================================================
// MODERATION
// =====================================================

// MuteUser mutes a user in a chat room
func (s *LiveChatService) MuteUser(ctx context.Context, roomID, userID string, durationMinutes int) error {
	s.roomsMu.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMu.RUnlock()

	if !exists {
		return ErrChatRoomNotFound
	}

	room.mu.Lock()
	if member, ok := room.Members[userID]; ok {
		member.IsMuted = true
		member.MuteUntil = time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	}
	room.mu.Unlock()

	query := `
		INSERT INTO chat_moderation (id, room_id, user_id, action, expires_at, created_at)
		VALUES (?, ?, ?, 'mute', ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		uuid.New().String(), roomID, userID,
		time.Now().Add(time.Duration(durationMinutes)*time.Minute),
	)
	return err
}

// BanUser permanently bans a user from a chat room
func (s *LiveChatService) BanUser(ctx context.Context, roomID, userID, reason string) error {
	query := `
		INSERT INTO chat_moderation (id, room_id, user_id, action, reason, created_at)
		VALUES (?, ?, ?, 'ban', ?, NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		uuid.New().String(), roomID, userID, reason,
	)

	// Remove from room
	s.LeaveChatRoom(roomID, userID)
	return err
}

// DeleteMessage removes a message (moderator action)
func (s *LiveChatService) DeleteMessage(ctx context.Context, messageID, moderatorID string) error {
	query := `UPDATE chat_messages SET is_deleted = 1, deleted_by = ?, deleted_at = NOW() WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, moderatorID, messageID)
	return err
}

// =====================================================
// HELPER METHODS
// =====================================================

func (s *LiveChatService) loadChatRoom(ctx context.Context, roomID string) (*ChatRoom, error) {
	room := &ChatRoom{
		Members: make(map[string]*ChatMember),
	}
	query := `
		SELECT id, stream_id, name, is_active, is_moderated, slow_mode, max_message_length, created_at
		FROM chat_rooms WHERE id = ?
	`
	err := s.db.QueryRowContext(ctx, query, roomID).Scan(
		&room.ID, &room.StreamID, &room.Name, &room.IsActive,
		&room.IsModerated, &room.SlowMode, &room.MaxMessageLen, &room.CreatedAt,
	)
	return room, err
}

func (s *LiveChatService) isUserBanned(ctx context.Context, roomID, userID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM chat_moderation
		WHERE room_id = ? AND user_id = ? AND action = 'ban'
	`
	err := s.db.QueryRowContext(ctx, query, roomID, userID).Scan(&count)
	return count > 0, err
}

func (s *LiveChatService) getUserBadges(ctx context.Context, userID string) []string {
	badges := []string{}
	// Subscriber badges, moderator badges, etc.
	query := `
		SELECT badge_type FROM user_badges
		WHERE user_id = ? AND is_active = 1
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return badges
	}
	defer rows.Close()
	for rows.Next() {
		var badge string
		if err := rows.Scan(&badge); err == nil {
			badges = append(badges, badge)
		}
	}
	return badges
}

func (s *LiveChatService) isInappropriate(message string) bool {
	lowerMsg := message
	for _, kw := range s.bannedKws {
		if containsIgnoreCase(lowerMsg, kw) {
			return true
		}
	}
	return false
}

func (s *LiveChatService) extractMentions(message string) []string {
	mentions := []string{}
	// Simple @mention extraction
	for i := 0; i < len(message); i++ {
		if message[i] == '@' {
			j := i + 1
			for j < len(message) && (isAlphanumeric(message[j]) || message[j] == '_') {
				j++
			}
			if j > i+1 {
				mentions = append(mentions, message[i+1:j])
			}
			i = j
		}
	}
	return mentions
}

func (s *LiveChatService) getSuperChatColor(amount float64) string {
	switch {
	case amount >= 100:
		return "#E91E63" // Red - highest tier
	case amount >= 50:
		return "#FF5722" // Orange-red
	case amount >= 20:
		return "#FF9800" // Orange
	case amount >= 10:
		return "#FFEB3B" // Yellow
	case amount >= 5:
		return "#4CAF50" // Green
	default:
		return "#2196F3" // Blue
	}
}

func containsIgnoreCase(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
