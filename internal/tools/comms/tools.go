// Package comms registers MCP tools for comms operations.
package comms

import (
	"encoding/json"
	"fmt"

	"github.com/unbound-force/replicator/internal/comms"
	"github.com/unbound-force/replicator/internal/db"
	"github.com/unbound-force/replicator/internal/tools/registry"
)

// Register adds all comms tools to the registry.
func Register(reg *registry.Registry, store *db.Store) {
	reg.Register(commsInit(store))
	reg.Register(commsSend(store))
	reg.Register(commsInbox(store))
	reg.Register(commsReadMessage(store))
	reg.Register(commsReserve(store))
	reg.Register(commsRelease(store))
	reg.Register(commsReleaseAll(store))
	reg.Register(commsReleaseAgent(store))
	reg.Register(commsAck(store))
	reg.Register(commsHealth(store))
}

func commsInit(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_init",
		Description: "Initialize comms session. Registers the agent.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["project_path"],
			"properties": {
				"agent_name":       {"type": "string"},
				"project_path":     {"type": "string"},
				"task_description": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AgentName       string `json:"agent_name"`
				ProjectPath     string `json:"project_path"`
				TaskDescription string `json:"task_description"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			agentName := input.AgentName
			if agentName == "" {
				agentName = "default"
			}
			agent, err := comms.Init(store, agentName, input.ProjectPath, input.TaskDescription)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(agent, "", "  ")
			return string(out), err
		},
	}
}

func commsSend(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_send",
		Description: "Send a message to other agents via comms.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["to", "subject", "body"],
			"properties": {
				"to":           {"type": "array", "items": {"type": "string"}},
				"subject":      {"type": "string"},
				"body":         {"type": "string"},
				"importance":   {"type": "string", "enum": ["low", "normal", "high", "urgent"]},
				"thread_id":    {"type": "string"},
				"ack_required": {"type": "boolean"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				From string `json:"from"`
				comms.SendInput
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			fromAgent := input.From
			if fromAgent == "" {
				fromAgent = "unknown"
			}
			if err := comms.Send(store, fromAgent, input.SendInput); err != nil {
				return "", err
			}
			return `{"status": "sent"}`, nil
		},
	}
}

func commsInbox(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_inbox",
		Description: "Fetch inbox (max 5 messages, bodies excluded).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"limit":       {"type": "number", "maximum": 5},
				"urgent_only": {"type": "boolean"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AgentName  string `json:"agent_name"`
				Limit      int    `json:"limit"`
				UrgentOnly bool   `json:"urgent_only"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			agentName := input.AgentName
			if agentName == "" {
				agentName = "default"
			}
			summaries, err := comms.Inbox(store, agentName, input.Limit, input.UrgentOnly)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(summaries, "", "  ")
			return string(out), err
		},
	}
}

func commsReadMessage(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_read_message",
		Description: "Fetch one message body by ID.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["message_id"],
			"properties": {
				"message_id": {"type": "number"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				MessageID int `json:"message_id"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			msg, err := comms.ReadMessage(store, input.MessageID)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(msg, "", "  ")
			return string(out), err
		},
	}
}

func commsReserve(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_reserve",
		Description: "Reserve file paths for exclusive editing.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["paths"],
			"properties": {
				"paths":       {"type": "array", "items": {"type": "string"}},
				"exclusive":   {"type": "boolean"},
				"reason":      {"type": "string"},
				"ttl_seconds": {"type": "number"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AgentName  string   `json:"agent_name"`
				Paths      []string `json:"paths"`
				Exclusive  bool     `json:"exclusive"`
				Reason     string   `json:"reason"`
				TTLSeconds int      `json:"ttl_seconds"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			agentName := input.AgentName
			if agentName == "" {
				agentName = "default"
			}
			reservations, err := comms.Reserve(store, agentName, input.Paths, input.Exclusive, input.Reason, input.TTLSeconds)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(reservations, "", "  ")
			return string(out), err
		},
	}
}

func commsRelease(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_release",
		Description: "Release file reservations.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"paths":           {"type": "array", "items": {"type": "string"}},
				"reservation_ids": {"type": "array", "items": {"type": "number"}}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				Paths          []string `json:"paths"`
				ReservationIDs []int    `json:"reservation_ids"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &input); err != nil {
					return "", err
				}
			}
			if err := comms.Release(store, input.Paths, input.ReservationIDs); err != nil {
				return "", err
			}
			return `{"status": "released"}`, nil
		},
	}
}

func commsReleaseAll(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_release_all",
		Description: "Release all file reservations in the project.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			if err := comms.ReleaseAll(store, ""); err != nil {
				return "", err
			}
			return `{"status": "all_released"}`, nil
		},
	}
}

func commsReleaseAgent(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_release_agent",
		Description: "Release all file reservations for a specific agent.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["agent_name"],
			"properties": {
				"agent_name": {"type": "string"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				AgentName string `json:"agent_name"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			if err := comms.ReleaseAgent(store, input.AgentName); err != nil {
				return "", err
			}
			return fmt.Sprintf(`{"status": "released", "agent": %q}`, input.AgentName), nil
		},
	}
}

func commsAck(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_ack",
		Description: "Acknowledge a message.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"required": ["message_id"],
			"properties": {
				"message_id": {"type": "number"}
			}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			var input struct {
				MessageID int `json:"message_id"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return "", err
			}
			if err := comms.Ack(store, input.MessageID); err != nil {
				return "", err
			}
			return `{"status": "acknowledged"}`, nil
		},
	}
}

func commsHealth(store *db.Store) *registry.Tool {
	return &registry.Tool{
		Name:        "comms_health",
		Description: "Check comms database health.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
		Execute: func(args json.RawMessage) (string, error) {
			// Verify the database is accessible by counting key tables.
			var agentCount, msgCount, resCount int
			if err := store.DB.QueryRow("SELECT COUNT(*) FROM agents").Scan(&agentCount); err != nil {
				return "", fmt.Errorf("count agents: %w", err)
			}
			if err := store.DB.QueryRow("SELECT COUNT(*) FROM messages").Scan(&msgCount); err != nil {
				return "", fmt.Errorf("count messages: %w", err)
			}
			if err := store.DB.QueryRow("SELECT COUNT(*) FROM reservations").Scan(&resCount); err != nil {
				return "", fmt.Errorf("count reservations: %w", err)
			}

			result := map[string]any{
				"status":       "healthy",
				"agents":       agentCount,
				"messages":     msgCount,
				"reservations": resCount,
			}
			out, err := json.MarshalIndent(result, "", "  ")
			return string(out), err
		},
	}
}
