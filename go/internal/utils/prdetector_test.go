package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPullRequests_GitHub(t *testing.T) {
	prs := utils.DetectPullRequests("See https://github.com/owner/repo/pull/123 for details")
	require.Len(t, prs, 1)
	assert.Equal(t, "https://github.com/owner/repo/pull/123", prs[0].URL)
	assert.Equal(t, "github", prs[0].Platform)
	assert.Equal(t, 123, prs[0].Number)
}

func TestDetectPullRequests_GitLab(t *testing.T) {
	prs := utils.DetectPullRequests("See https://gitlab.com/owner/repo/-/merge_requests/456")
	require.Len(t, prs, 1)
	assert.Equal(t, "gitlab", prs[0].Platform)
	assert.Equal(t, 456, prs[0].Number)
}

func TestDetectPullRequests_Bitbucket(t *testing.T) {
	prs := utils.DetectPullRequests("https://bitbucket.org/owner/repo/pull-requests/789")
	require.Len(t, prs, 1)
	assert.Equal(t, "bitbucket", prs[0].Platform)
	assert.Equal(t, 789, prs[0].Number)
}

func TestDetectPullRequests_Multiple(t *testing.T) {
	text := "PR1: https://github.com/a/b/pull/1 and PR2: https://github.com/a/b/pull/2"
	prs := utils.DetectPullRequests(text)
	assert.Len(t, prs, 2)
}

func TestDetectPullRequests_Deduplicates(t *testing.T) {
	text := "https://github.com/a/b/pull/1 and https://github.com/a/b/pull/1 again"
	prs := utils.DetectPullRequests(text)
	assert.Len(t, prs, 1)
}

func TestDetectPullRequests_Empty(t *testing.T) {
	prs := utils.DetectPullRequests("no PRs here")
	assert.Empty(t, prs)
}

func TestExtractPullRequestURL(t *testing.T) {
	url := utils.ExtractPullRequestURL("See https://github.com/a/b/pull/1 for details")
	assert.NotNil(t, url)
	assert.Equal(t, "https://github.com/a/b/pull/1", *url)
}

func TestExtractPullRequestURL_None(t *testing.T) {
	url := utils.ExtractPullRequestURL("no PRs here")
	assert.Nil(t, url)
}
