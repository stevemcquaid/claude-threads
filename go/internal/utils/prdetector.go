package utils

import (
	"regexp"
	"strconv"
)

// PullRequestInfo contains information about a detected pull/merge request.
type PullRequestInfo struct {
	URL      string
	Platform string
	Number   int
	Repo     string
}

var prPatterns = []struct {
	re       *regexp.Regexp
	platform string
	numGroup int
}{
	{
		re:       regexp.MustCompile(`https?://github\.com/[\w./-]+/pull/(\d+)`),
		platform: "github",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://(?:[\w.-]+\.)?gitlab\.com/[\w./-]+/-/merge_requests/(\d+)`),
		platform: "gitlab",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://bitbucket\.org/[\w./-]+/pull-requests/(\d+)`),
		platform: "bitbucket",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://[\w.-]+\.visualstudio\.com/[\w./-]+/_git/[\w.-]+/pullrequest/(\d+)`),
		platform: "azuredevops",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://dev\.azure\.com/[\w./-]+/_git/[\w.-]+/pullrequest/(\d+)`),
		platform: "azuredevops",
		numGroup: 1,
	},
}

// DetectPullRequests finds all PR/MR URLs in text. Deduplicates by URL.
func DetectPullRequests(text string) []PullRequestInfo {
	seen := map[string]bool{}
	var results []PullRequestInfo

	for _, p := range prPatterns {
		matches := p.re.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			url := m[0]
			if seen[url] {
				continue
			}
			seen[url] = true
			num, _ := strconv.Atoi(m[p.numGroup])
			results = append(results, PullRequestInfo{
				URL:      url,
				Platform: p.platform,
				Number:   num,
			})
		}
	}
	return results
}

// ExtractPullRequestURL returns the first PR URL found in text, or nil.
func ExtractPullRequestURL(text string) *string {
	prs := DetectPullRequests(text)
	if len(prs) == 0 {
		return nil
	}
	url := prs[0].URL
	return &url
}
