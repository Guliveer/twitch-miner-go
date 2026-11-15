// Package feedback builds the GitHub issue form links shown by the web
// interfaces and the system tray, and fills the placeholders embedded in the
// web UI shells.
package feedback

import (
	"html"
	"net/url"
	"strings"

	"github.com/Guliveer/twitch-miner-go/internal/version"
)

const issuesNewURL = "https://github.com/Guliveer/twitch-miner-go/issues/new"

// Issue form templates live in .github/ISSUE_TEMPLATE.
const (
	bugReportTemplate      = "bug_report.yml"
	featureRequestTemplate = "feature_request.yml"
)

// Placeholders used by the embedded web UI shells. They are replaced once at
// startup so the pages stay static files on disk.
const (
	versionPlaceholder        = "{{APP_VERSION}}"
	bugReportPlaceholder      = "{{BUG_REPORT_URL}}"
	featureRequestPlaceholder = "{{FEATURE_REQUEST_URL}}"
)

// BugReportURL returns the bug report form prefilled with the running version.
func BugReportURL() string {
	return issueURL(bugReportTemplate)
}

// FeatureRequestURL returns the feature request form prefilled with the
// running version.
func FeatureRequestURL() string {
	return issueURL(featureRequestTemplate)
}

// Expand replaces the version and issue-link placeholders in an embedded page.
// The substituted values are HTML-escaped because the version comes from the
// build environment.
func Expand(page string) string {
	replacer := strings.NewReplacer(
		versionPlaceholder, html.EscapeString(version.Label()),
		bugReportPlaceholder, html.EscapeString(BugReportURL()),
		featureRequestPlaceholder, html.EscapeString(FeatureRequestURL()),
	)
	return replacer.Replace(page)
}

// issueURL builds a new-issue URL. The version parameter targets the "version"
// field of the issue form, which GitHub prefills from the query string.
func issueURL(template string) string {
	query := url.Values{}
	query.Set("template", template)
	if version.Number != "" {
		query.Set("version", version.Number)
	}
	return issuesNewURL + "?" + query.Encode()
}
