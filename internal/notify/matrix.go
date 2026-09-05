package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// Matrix sends notifications via the Matrix client-server API.
type Matrix struct {
	baseNotifier
	homeserver  string
	accessToken string
	roomID      string
	httpClient  *http.Client
	txnCounter  atomic.Int64
}

// Send puts a message into the configured Matrix room. The title is rendered
// as the first line—every other provider has a dedicated title slot, Matrix
// does not—and an HTML variant is attached so clients that render formatted
// bodies show the header and field labels in bold.
func (m *Matrix) Send(ctx context.Context, _ model.Event, title, message string) error {
	encodedRoomID := url.PathEscape(m.roomID)
	txnID := fmt.Sprintf("m%d.%d", time.Now().UnixNano(), m.txnCounter.Add(1))

	apiURL := fmt.Sprintf("https://%s/_matrix/client/r0/rooms/%s/send/m.room.message/%s",
		m.homeserver, encodedRoomID, txnID)

	plain := message
	if title != "" {
		plain = title + "\n" + message
	}

	payload := map[string]string{
		"msgtype":        "m.text",
		"body":           plain,
		"format":         "org.matrix.custom.html",
		"formatted_body": matrixHTML(title, message),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("matrix: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("matrix: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.accessToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("matrix: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("matrix: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// matrixHTML renders the notification as Matrix-flavoured HTML: the title and
// the message's first line become headers, and any "Label: value" line has its
// label emphasised. All content is escaped first—streamer names, stream titles
// and prediction questions are attacker-controlled text.
func matrixHTML(title, message string) string {
	var sb strings.Builder

	if title != "" {
		sb.WriteString("<b>" + html.EscapeString(title) + "</b>")
	}

	for i, line := range strings.Split(message, "\n") {
		if sb.Len() > 0 {
			sb.WriteString("<br/>")
		}

		if i == 0 {
			sb.WriteString("<b>" + html.EscapeString(line) + "</b>")
			continue
		}

		if label, value, ok := strings.Cut(line, ": "); ok {
			sb.WriteString("<b>" + html.EscapeString(label) + ":</b> " + html.EscapeString(value))
			continue
		}

		sb.WriteString(html.EscapeString(line))
	}

	return sb.String()
}
