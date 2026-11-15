package feedback

import (
	"net/url"
	"strings"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/version"
)

func setVersion(t *testing.T, number string) {
	t.Helper()
	original := version.Number
	t.Cleanup(func() { version.Number = original })
	version.Number = number
}

func queryOf(t *testing.T, raw string) url.Values {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	if parsed.Host != "github.com" || parsed.Path != "/Guliveer/twitch-miner-go/issues/new" {
		t.Fatalf("unexpected target: %s", raw)
	}
	return parsed.Query()
}

func TestIssueURLsCarryTemplateAndVersion(t *testing.T) {
	setVersion(t, "1.26.6")

	tests := map[string]struct {
		url      string
		template string
	}{
		"bug report":      {url: BugReportURL(), template: "bug_report.yml"},
		"feature request": {url: FeatureRequestURL(), template: "feature_request.yml"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			query := queryOf(t, tc.url)
			if got := query.Get("template"); got != tc.template {
				t.Errorf("template = %q, want %q", got, tc.template)
			}
			if got := query.Get("version"); got != "1.26.6" {
				t.Errorf("version = %q, want %q", got, "1.26.6")
			}
		})
	}
}

func TestIssueURLsOmitEmptyVersion(t *testing.T) {
	setVersion(t, "")

	query := queryOf(t, BugReportURL())
	if _, ok := query["version"]; ok {
		t.Errorf("version parameter should be omitted when the build has none: %v", query)
	}
	if got := query.Get("template"); got != "bug_report.yml" {
		t.Errorf("template = %q, want %q", got, "bug_report.yml")
	}
}

func TestExpandReplacesPlaceholders(t *testing.T) {
	setVersion(t, "1.26.6")

	page := `<span>{{APP_VERSION}}</span><a href="{{BUG_REPORT_URL}}">bug</a><a href="{{FEATURE_REQUEST_URL}}">feature</a>`
	got := Expand(page)

	if strings.Contains(got, "{{") {
		t.Fatalf("placeholder left unexpanded: %s", got)
	}
	if !strings.Contains(got, ">v1.26.6<") {
		t.Errorf("version label missing: %s", got)
	}
	if !strings.Contains(got, "template=bug_report.yml") || !strings.Contains(got, "template=feature_request.yml") {
		t.Errorf("issue links missing: %s", got)
	}
}
