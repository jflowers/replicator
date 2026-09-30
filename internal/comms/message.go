package comms

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/unbound-force/replicator/internal/db"
)

// Message represents a full swarm mail message.
type Message struct {
	ID           int    `json:"id"`
	FromAgent    string `json:"from_agent"`
	ToAgents     string `json:"to_agents"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	Importance   string `json:"importance"`
	ThreadID     string `json:"thread_id,omitempty"`
	AckRequired  bool   `json:"ack_required"`
	Acknowledged bool   `json:"acknowledged"`
	CreatedAt    string `json:"created_at"`
}

// MessageSummary is a message without the body, used for inbox listings.
type MessageSummary struct {
	ID           int    `json:"id"`
	FromAgent    string `json:"from_agent"`
	Subject      string `json:"subject"`
	Importance   string `json:"importance"`
	ThreadID     string `json:"thread_id,omitempty"`
	AckRequired  bool   `json:"ack_required"`
	Acknowledged bool   `json:"acknowledged"`
	CreatedAt    string `json:"created_at"`
}

// SendInput defines the input for sending a message.
type SendInput struct {
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	Importance  string   `json:"importance,omitempty"`
	ThreadID    string   `json:"thread_id,omitempty"`
	AckRequired bool     `json:"ack_required,omitempty"`
}

// Send delivers a message to one or more agents.
func Send(store *db.Store, fromAgent string, input SendInput) error {
	importance := input.Importance
	if importance == "" {
		importance = "normal"
	}

	toJSON, err := json.Marshal(input.To)
	if err != nil {
		return fmt.Errorf("marshal to_agents: %w", err)
	}

	ackRequired := 0
	if input.AckRequired {
		ackRequired = 1
	}

	_, err = store.DB.Exec(`
		INSERT INTO messages (from_agent, to_agents, subject, body, importance, thread_id, ack_required)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fromAgent, string(toJSON), input.Subject, input.Body, importance, input.ThreadID, ackRequired,
	)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	return nil
}

// Inbox returns message summaries (no body) for an agent, max 5 results.
func Inbox(store *db.Store, agentName string, limit int, urgentOnly bool) (summaries []MessageSummary, err error) {
	if limit <= 0 || limit > 5 {
		limit = 5
	}

	// Use json_each() to match the agent name as an exact JSON array element,
	// avoiding partial name collisions (e.g., "worker-1" matching "worker-10").
	query := `
		SELECT id, from_agent, subject, importance, thread_id, ack_required, acknowledged, created_at
		FROM messages
		WHERE EXISTS (SELECT 1 FROM json_each(to_agents) WHERE json_each.value = ?)`
	args := []any{agentName}

	if urgentOnly {
		query += " AND importance = 'urgent'"
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := store.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query inbox: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			summaries = nil
			err = combineInboxCloseError(err, closeErr)
		}
	}()

	for rows.Next() {
		var s MessageSummary
		var threadID *string
		var ackReq, acked int
		err := rows.Scan(&s.ID, &s.FromAgent, &s.Subject, &s.Importance,
			&threadID, &ackReq, &acked, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		if threadID != nil {
			s.ThreadID = *threadID
		}
		s.AckRequired = ackReq == 1
		s.Acknowledged = acked == 1
		summaries = append(summaries, s)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate inbox: %w", rowsErr)
	}

	if summaries == nil {
		summaries = []MessageSummary{}
	}
	return summaries, nil
}

func combineInboxCloseError(primary, closeErr error) error {
	if closeErr == nil {
		return primary
	}
	return errors.Join(primary, fmt.Errorf("close inbox rows: %w", closeErr))
}

// ReadMessage returns a full message by ID.
func ReadMessage(store *db.Store, messageID int) (*Message, error) {
	var m Message
	var threadID *string
	var ackReq, acked int
	err := store.DB.QueryRow(`
		SELECT id, from_agent, to_agents, subject, body, importance, thread_id,
		       ack_required, acknowledged, created_at
		FROM messages WHERE id = ?`, messageID,
	).Scan(&m.ID, &m.FromAgent, &m.ToAgents, &m.Subject, &m.Body, &m.Importance,
		&threadID, &ackReq, &acked, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("read message %d: %w", messageID, err)
	}
	if threadID != nil {
		m.ThreadID = *threadID
	}
	m.AckRequired = ackReq == 1
	m.Acknowledged = acked == 1
	return &m, nil
}

// Ack acknowledges a message.
func Ack(store *db.Store, messageID int) error {
	result, err := store.DB.Exec(
		"UPDATE messages SET acknowledged = 1 WHERE id = ?", messageID,
	)
	if err != nil {
		return fmt.Errorf("ack message: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("message %d not found", messageID)
	}
	return nil
}
