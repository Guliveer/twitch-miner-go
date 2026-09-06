package notify

import (
	"strings"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestBuildBodyBetResultNamesChannel(t *testing.T) {
	body := buildBody(model.EventBetWin, "🏆 Prediction result", []logger.Field{
		{Key: "streamer", Value: "psychohypnotic"},
		{Key: "category", Value: "phasmophobia"},
		{Key: "title", Value: "Will it be Obambo, Thaye or Il Mimo?"},
		{Key: "choice", Value: "Yes (BLUE)"},
		{Key: "amount", Value: "881"},
		{Key: "result", Value: "WIN, Gained: +1762"},
	})

	want := strings.Join([]string{
		"🏆 Bet won",
		"Channel: psychohypnotic (phasmophobia)",
		"Question: Will it be Obambo, Thaye or Il Mimo?",
		"Pick: Yes (BLUE) · 881 points",
		"Result: WIN, Gained: +1762",
	}, "\n")

	if body != want {
		t.Fatalf("bet body mismatch\ngot:\n%s\nwant:\n%s", body, want)
	}
}

func TestBuildBodyBetStartShowsCountdown(t *testing.T) {
	body := buildBody(model.EventBetStart, "🎰 Placing bet", []logger.Field{
		{Key: "streamer", Value: "gronkhtv"},
		{Key: "category", Value: "just-chatting"},
		{Key: "title", Value: "Wer gewinnt?"},
		{Key: "in_seconds", Value: "50"},
	})

	want := strings.Join([]string{
		"🎰 Placing bet",
		"Channel: gronkhtv (just-chatting)",
		"Question: Wer gewinnt?",
		"Remaining: 50s",
	}, "\n")

	if body != want {
		t.Fatalf("bet start body mismatch\ngot:\n%s\nwant:\n%s", body, want)
	}
}

func TestBuildBodyBetKeepsUnknownFields(t *testing.T) {
	body := buildBody(model.EventBetFilters, "🎰 Insufficient points for bet", []logger.Field{
		{Key: "streamer", Value: "iblali"},
		{Key: "balance", Value: "500"},
		{Key: "minimum", Value: "2000"},
		{Key: "some_new_field", Value: "42"},
	})

	for _, want := range []string{
		"🎰 Bet skipped",
		"Channel: iblali",
		"Balance: 500",
		"Minimum: 2000",
		"some_new_field: 42",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\ngot:\n%s", want, body)
		}
	}
}

func TestBuildBodyBetGeneralFallsBackToMessage(t *testing.T) {
	// BET_GENERAL has no fixed headline; the logged message must survive.
	body := buildBody(model.EventBetGeneral, "🎰 Prediction result", []logger.Field{
		{Key: "streamer", Value: "guppa"},
	})

	if !strings.HasPrefix(body, "🎰 Prediction result\n") {
		t.Fatalf("expected logged message as headline, got:\n%s", body)
	}
}

func TestBuildBodyNonBetKeepsCompactForm(t *testing.T) {
	body := buildBody(model.EventDropClaim, "📦 Claiming drop", []logger.Field{
		{Key: "streamer", Value: "thekller"},
		{Key: "category", Value: "escape-from-tarkov"},
		{Key: "drop", Value: "Drop 4"},
	})

	want := "📦 Claiming drop (drop={Drop 4})"
	if body != want {
		t.Fatalf("compact body mismatch\ngot:  %s\nwant: %s", body, want)
	}
}

func TestBuildBodyNonBetWithoutFields(t *testing.T) {
	body := buildBody(model.EventStreamerOnline, "🟢 Streamer is online", nil)
	if body != "🟢 Streamer is online" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestMatrixHTMLEscapesAndEmphasises(t *testing.T) {
	got := matrixHTML("3jakec | 📺 psychohypnotic", "🏆 Bet won\nChannel: psychohypnotic (phasmophobia)")

	want := "<b>3jakec | 📺 psychohypnotic</b>" +
		"<br/><b>🏆 Bet won</b>" +
		"<br/><b>Channel:</b> psychohypnotic (phasmophobia)"

	if got != want {
		t.Fatalf("html mismatch\ngot:  %s\nwant: %s", got, want)
	}
}

func TestMatrixHTMLEscapesHostileTitles(t *testing.T) {
	// Stream titles and prediction questions are attacker-controlled.
	got := matrixHTML("", "🎰 Bet\nQuestion: <img src=x onerror=alert(1)>")

	if strings.Contains(got, "<img") {
		t.Fatalf("raw markup leaked into formatted body: %s", got)
	}
	if !strings.Contains(got, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Fatalf("expected escaped markup, got: %s", got)
	}
}

func TestMatrixHTMLWithoutTitle(t *testing.T) {
	got := matrixHTML("", "🟢 Streamer is online")
	if got != "<b>🟢 Streamer is online</b>" {
		t.Fatalf("unexpected html: %s", got)
	}
}