package manage

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/repofuel/repofuel/ingest/internal/entity"
)

type FailureSummary struct{}

type PushSummary struct {
	maxRisk *float32
	Commits []*entity.Commit
}

type PullRequestSummaryItem struct {
	Pull    *entity.PullRequest
	Commits []*entity.Commit
}

type PullRequestSummary struct {
	maxRisk *float32
	items   []*PullRequestSummaryItem
}

func (s *PullRequestSummary) Title() string {
	return title(s.MaxRisk())
}

func (s *PullRequestSummary) Summary() string {
	return summary(s.MaxRisk())
}

func (s *PullRequestSummary) DetailsText(providerName, providerUrl, ownerName, repoName string) string {
	const repofuelDomain = "http://dev.repofuel.com"
	var d = templateData{
		RepofuelDomain: repofuelDomain,
		OwnerName:      ownerName,
		RepoName:       repoName,
		ProviderName:   providerName,
		ProviderURL:    providerUrl,
	}

	var sb strings.Builder
	for _, item := range s.items {
		d.writPullRequestHeader(&sb, item.Pull)
		d.writeCommitsTable(&sb, item.Commits)
	}

	return sb.String()
}

func NewFailureSummary() *FailureSummary {
	return &FailureSummary{}
}

func (e FailureSummary) Title() string {
	return "Failed"
}

func (e FailureSummary) Summary() string {
	return "An error happened in processing the check"
}

func (e FailureSummary) DetailsText(providerName, providerUrl, ownerName, repoName string) string {
	return e.Summary()
}

func NewPushSummery(commits []*entity.Commit) *PushSummary {
	return &PushSummary{
		Commits: commits,
	}
}

func (c *PushSummary) MaxRisk() float32 {
	if c.maxRisk != nil {
		return *c.maxRisk
	}

	var maxRisk float32

	for _, c := range c.Commits {
		if c.Analysis.BugPotential > maxRisk {
			maxRisk = c.Analysis.BugPotential
		}
	}

	c.maxRisk = &maxRisk
	return maxRisk
}

func (c *PushSummary) Title() string {
	return title(c.MaxRisk())
}

func title(risk float32) string {
	switch {
	case risk > .80:
		return "High bug potential"
	case risk > .50:
		return "Moderate bug potential"
	default:
		return "Low bug potential"
	}
}

func (c *PushSummary) Summary() string {
	return summary(c.MaxRisk())
}

func summary(risk float32) string {
	switch {
	case risk > .80:
		return "The likelihood of this modification introducing a bug is high.\n>*For more details click on the score of the individual commits below.*"
	case risk > .50:
		return "The likelihood of this modification introducing a bug is moderate.\n>*For more details click on the score of the individual commits below.*"
	default:
		return "The likelihood of this modification introducing a bug is low.\n>*For more details click on the score of the individual commits below.*"
	}
}

func (c *PushSummary) DetailsText(providerName, providerUrl, ownerName, repoName string) string {
	const repofuelDomain = "http://dev.repofuel.com"
	var data = templateData{
		RepofuelDomain: repofuelDomain,
		OwnerName:      ownerName,
		RepoName:       repoName,
		ProviderName:   providerName,
		ProviderURL:    providerUrl,
	}

	var sb strings.Builder
	data.writeCommitsTable(&sb, c.Commits)
	return sb.String()
}

type templateData struct {
	RepofuelDomain, OwnerName, RepoName, ProviderName, ProviderURL string
}

func (d *templateData) writPullRequestHeader(w io.Writer, pull *entity.PullRequest) {
	_, _ = fmt.Fprintf(w, "## Pull request [#%d](%s%s/%s/pull/%d)\n>%s\n",
		pull.Source.Number,
		d.ProviderURL,
		d.OwnerName,
		d.RepoName,
		pull.Source.Number,
		pull.Source.Title,
	)
}

func (d *templateData) writCommitHashLink(w io.Writer, commit *entity.Commit) {
	_, _ = fmt.Fprintf(w, "[![%.f%%](%s/svg/scores/%.f.svg)](%s/repos/%s/%s/%s/Commits/%s)",
		commit.Analysis.BugPotential*100,
		d.RepofuelDomain,
		math.Ceil(float64(commit.Analysis.BugPotential)*10),
		d.RepofuelDomain,
		d.ProviderName,
		d.OwnerName,
		d.RepoName,
		commit.ID.CommitHash.Hex(),
	)
}

func (d *templateData) writeCommitsTable(sb *strings.Builder, commits []*entity.Commit) {
	sb.WriteString(`
| Commit | Score |
|--------|-------|
`)

	for _, commit := range commits {
		sb.WriteString("| ")
		d.writeCommitScore(sb, commit)
		sb.WriteString("  ")
		sb.WriteString(commit.Message)
		sb.WriteString(" | ")
		d.writCommitHashLink(sb, commit)
		sb.WriteString(" |\n")
	}
}

func (d *templateData) writeCommitScore(w io.Writer, commit *entity.Commit) {
	_, _ = fmt.Fprintf(w, "[`%s`](%s%s/%s/commit/%s)",
		commit.ID.CommitHash.ShortHash(),
		d.ProviderURL,
		d.OwnerName,
		d.RepoName,
		commit.ID.CommitHash.Hex(),
	)
}

func NewPullRequestSummary(count int) *PullRequestSummary {
	return &PullRequestSummary{
		items: make([]*PullRequestSummaryItem, 0, count),
	}
}

func (s *PullRequestSummary) AddPullRequest(pull *entity.PullRequest, commits []*entity.Commit) {
	s.items = append(s.items, &PullRequestSummaryItem{
		Pull:    pull,
		Commits: commits,
	})
}

func (c *PullRequestSummary) MaxRisk() float32 {
	if c.maxRisk != nil {
		return *c.maxRisk
	}

	var maxRisk float32

	for _, item := range c.items {
		for _, c := range item.Commits {
			if c.Analysis.BugPotential > maxRisk {
				maxRisk = c.Analysis.BugPotential
			}
		}
	}

	c.maxRisk = &maxRisk
	return maxRisk
}
