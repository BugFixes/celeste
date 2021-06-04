package ticketing

import (
	"encoding/json"
	"fmt"
	"net/http"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

func (t Ticketing) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("x-agent-id")
	if agentID == "" {
		bugLog.Debugf("ticket fetch system failed:, %+v", r)
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Invalid AgentID",
		}); err != nil {
			bugLog.Debugf("ticket parse failed json: %v", err)
		}
		return
	}

	var ticket Ticket
	ticket.AgentID = agentID

	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		bugLog.Debugf("ticket parse failed: %+v, %+v", err, r)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: "Body is missing",
		}); err != nil {
			bugLog.Debugf("ticket parse failed json: %+v", err)
		}
		return
	}

	if err := t.CreateTicket(&ticket); err != nil {
		bugLog.Debugf("ticket create failed: %v, %+v", err, r)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(struct {
			Error string
		}{
			Error: fmt.Sprintf("ticket create failed: %+v", err),
		}); err != nil {
			bugLog.Debugf("ticket create failed json: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
