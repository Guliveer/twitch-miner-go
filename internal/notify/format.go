package notify

import (
	"fmt"
	"strings"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// betHeadlines replaces the raw log message for prediction events. The log
// itself stays machine-greppable; only the notification body is phrased for
// the operator reading it on a phone.
var betHeadlines = map[model.Event]string{
	model.EventBetStart:   "🎰 Placing bet",
	model.EventBetWin:     "🏆 Bet won",
	model.EventBetLose:    "💸 Bet lost",
	model.EventBetRefund:  "↩️ Bet refunded",
	model.EventBetFilters: "🎰 Bet skipped",
	model.EventBetFailed:  "🎰 Bet failed",
}

// betFieldLabels maps log field keys to the labels used in a bet notification
// body. Keys absent from this map are appended verbatim so a newly added log
// field never disappears silently.
var betFieldLabels = map[string]string{
	"title":      "Question",
	pickField:    "Pick",
	"result":     "Result",
	"balance":    "Balance",
	"minimum":    "Minimum",
	"in_seconds": "Remaining",
}

// betFieldOrder fixes the line order of the body. "pick" is synthetic: it is
// assembled from the choice/outcome and amount fields rather than logged as one.
var betFieldOrder = []string{"title", "pick", "in_seconds", "result", "balance", "minimum"}

// pickField is the synthetic key for the chosen outcome plus stake.
const pickField = "pick"

// betConsumedFields are rendered by dedicated logic and must not be emitted
// again by the trailing "unknown fields" pass.
var betConsumedFields = map[string]bool{
	"streamer": true,
	"category": true,
	"choice":   true,
	"outcome":  true,
	"amount":   true,
}

// isBetEvent reports whether an event belongs to the prediction family.
func isBetEvent(event model.Event) bool {
	switch event {
	case model.EventBetStart, model.EventBetWin, model.EventBetLose,
		model.EventBetRefund, model.EventBetFilters, model.EventBetFailed,
		model.EventBetGeneral:
		return true
	default:
		return false
	}
}

// fieldValue returns the value logged under key, or "" when absent.
func fieldValue(fields []logger.Field, key string) string {
	for _, f := range fields {
		if f.Key == key {
			return f.Value
		}
	}
	return ""
}

// buildBody renders the notification body for an event. Prediction events get
// a labelled multi-line layout that always names the channel; everything else
// keeps the compact single-line form, where the channel comes from the title.
func buildBody(event model.Event, msg string, fields []logger.Field) string {
	if isBetEvent(event) {
		return buildBetBody(event, msg, fields)
	}
	return buildCompactBody(msg, fields)
}

func buildBetBody(event model.Event, msg string, fields []logger.Field) string {
	lines := make([]string, 0, len(fields)+2)

	headline := betHeadlines[event]
	if headline == "" {
		headline = msg
	}
	lines = append(lines, headline)

	if channel := formatChannel(fields); channel != "" {
		lines = append(lines, "Channel: "+channel)
	}

	for _, key := range betFieldOrder {
		var value string
		switch key {
		case pickField:
			value = formatPick(fields)
		case "in_seconds":
			if raw := fieldValue(fields, key); raw != "" {
				value = raw + "s"
			}
		default:
			value = fieldValue(fields, key)
		}

		if value == "" {
			continue
		}
		lines = append(lines, betFieldLabels[key]+": "+value)
	}

	for _, f := range fields {
		if betConsumedFields[f.Key] || betFieldLabels[f.Key] != "" || f.Value == "" {
			continue
		}
		lines = append(lines, f.Key+": "+f.Value)
	}

	return strings.Join(lines, "\n")
}

// formatChannel renders "name (category)", degrading gracefully when either
// half is missing.
func formatChannel(fields []logger.Field) string {
	streamer := fieldValue(fields, "streamer")
	category := fieldValue(fields, "category")

	switch {
	case streamer != "" && category != "":
		return fmt.Sprintf("%s (%s)", streamer, category)
	case streamer != "":
		return streamer
	default:
		return category
	}
}

// formatPick renders the chosen outcome plus the staked amount.
func formatPick(fields []logger.Field) string {
	pick := fieldValue(fields, "choice")
	if pick == "" {
		pick = fieldValue(fields, "outcome")
	}
	amount := fieldValue(fields, "amount")

	switch {
	case pick != "" && amount != "":
		return fmt.Sprintf("%s · %s points", pick, amount)
	case pick != "":
		return pick
	case amount != "":
		return amount + " points"
	default:
		return ""
	}
}

// buildCompactBody keeps the historical one-line format for non-bet events:
// the message followed by its fields, minus streamer and category which the
// notification title already carries.
func buildCompactBody(msg string, fields []logger.Field) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		if f.Key == "streamer" || f.Key == "category" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s={%s}", f.Key, f.Value))
	}

	if len(parts) == 0 {
		return msg
	}
	return fmt.Sprintf("%s (%s)", msg, strings.Join(parts, ", "))
}