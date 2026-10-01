package comms

import (
	"errors"
	"testing"
)

func TestCombineInboxCloseError_PreservesBothErrors(t *testing.T) {
	errScan := errors.New("scan sentinel")
	errClose := errors.New("close sentinel")
	err := combineInboxCloseError(errScan, errClose)
	if !errors.Is(err, errScan) || !errors.Is(err, errClose) {
		t.Fatalf("combined error %v does not preserve both sentinels", err)
	}
}

func TestSendAndReadMessage(t *testing.T) {
	store := testStore(t)

	err := Send(store, "coordinator", SendInput{
		To:      []string{"worker-1"},
		Subject: "Start task",
		Body:    "Please begin implementing feature X",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Read the message.
	msg, err := ReadMessage(store, 1)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if msg.FromAgent != "coordinator" {
		t.Errorf("from_agent = %q, want %q", msg.FromAgent, "coordinator")
	}
	if msg.Subject != "Start task" {
		t.Errorf("subject = %q, want %q", msg.Subject, "Start task")
	}
	if msg.Body != "Please begin implementing feature X" {
		t.Errorf("body = %q, want %q", msg.Body, "Please begin implementing feature X")
	}
	if msg.Importance != "normal" {
		t.Errorf("importance = %q, want %q", msg.Importance, "normal")
	}
}

func TestSend_WithImportance(t *testing.T) {
	store := testStore(t)

	err := Send(store, "coordinator", SendInput{
		To:         []string{"worker-1"},
		Subject:    "Urgent",
		Body:       "Stop everything",
		Importance: "urgent",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	msg, err := ReadMessage(store, 1)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if msg.Importance != "urgent" {
		t.Errorf("importance = %q, want %q", msg.Importance, "urgent")
	}
}

func TestInbox(t *testing.T) {
	store := testStore(t)

	// Send 3 messages to worker-1.
	for i := 0; i < 3; i++ {
		if err := Send(store, "coordinator", SendInput{
			To:      []string{"worker-1"},
			Subject: "Message",
			Body:    "body",
		}); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}

	summaries, err := Inbox(store, "worker-1", 5, false)
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}
	if len(summaries) != 3 {
		t.Errorf("expected 3 messages, got %d", len(summaries))
	}
	// Summaries should not contain body (it's not in the struct).
}

func TestInbox_MaxFive(t *testing.T) {
	store := testStore(t)

	// Send 7 messages.
	for i := 0; i < 7; i++ {
		if err := Send(store, "coordinator", SendInput{
			To:      []string{"worker-1"},
			Subject: "Message",
			Body:    "body",
		}); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}

	summaries, err := Inbox(store, "worker-1", 0, false)
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}
	if len(summaries) != 5 {
		t.Errorf("expected max 5 messages, got %d", len(summaries))
	}
}

func TestInbox_UrgentOnly(t *testing.T) {
	store := testStore(t)

	if err := Send(store, "coordinator", SendInput{
		To:         []string{"worker-1"},
		Subject:    "Normal",
		Body:       "normal body",
		Importance: "normal",
	}); err != nil {
		t.Fatalf("Send normal: %v", err)
	}
	if err := Send(store, "coordinator", SendInput{
		To:         []string{"worker-1"},
		Subject:    "Urgent",
		Body:       "urgent body",
		Importance: "urgent",
	}); err != nil {
		t.Fatalf("Send urgent: %v", err)
	}

	summaries, err := Inbox(store, "worker-1", 5, true)
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 urgent message, got %d", len(summaries))
	}
	if summaries[0].Subject != "Urgent" {
		t.Errorf("subject = %q, want %q", summaries[0].Subject, "Urgent")
	}
}

func TestInbox_Empty(t *testing.T) {
	store := testStore(t)

	summaries, err := Inbox(store, "worker-1", 5, false)
	if err != nil {
		t.Fatalf("Inbox: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 messages, got %d", len(summaries))
	}
}

func TestAck(t *testing.T) {
	store := testStore(t)

	if err := Send(store, "coordinator", SendInput{
		To:          []string{"worker-1"},
		Subject:     "Ack me",
		Body:        "body",
		AckRequired: true,
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Verify ack_required is set.
	msg, _ := ReadMessage(store, 1)
	if !msg.AckRequired {
		t.Error("ack_required should be true")
	}
	if msg.Acknowledged {
		t.Error("acknowledged should be false initially")
	}

	// Ack the message.
	if err := Ack(store, 1); err != nil {
		t.Fatalf("Ack: %v", err)
	}

	// Verify acknowledged.
	msg, _ = ReadMessage(store, 1)
	if !msg.Acknowledged {
		t.Error("acknowledged should be true after ack")
	}
}

func TestAck_NotFound(t *testing.T) {
	store := testStore(t)

	err := Ack(store, 999)
	if err == nil {
		t.Error("expected error for nonexistent message")
	}
}

func TestReadMessage_NotFound(t *testing.T) {
	store := testStore(t)

	_, err := ReadMessage(store, 999)
	if err == nil {
		t.Error("expected error for nonexistent message")
	}
}
