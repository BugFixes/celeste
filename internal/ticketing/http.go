package ticketing

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (t Ticketing) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("x-agent-id")
	if agentID == "" {
		t.Logger.Errorf("ticket fetch system failed:, %+v", r)
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Invalid AgentID",
		}); err != nil {
			t.Logger.Errorf("ticket parse failed json: %v", err)
		}
		return
	}

	var ticket Ticket
	ticket.AgentID = agentID

	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		t.Logger.Errorf("ticket parse failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Body is missing",
		}); err != nil {
			t.Logger.Errorf("ticket parse failed json: %+v", err)
		}
		return
	}

	if err := t.CreateTicket(&ticket); err != nil {
		t.Logger.Errorf("ticket create failed: %v, %+v", err, r)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: fmt.Sprintf("ticket create failed: %+v", err),
		}); err != nil {
			t.Logger.Errorf("ticket create failed json: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
