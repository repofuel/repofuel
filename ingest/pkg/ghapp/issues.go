package ghapp

import (
	"bytes"
	"context"
	"regexp"
	"strconv"

	"github.com/repofuel/repofuel/pkg/common"
)

var (
	githubIssuesRegexp = regexp.MustCompile("((gh\\-|GH\\-)|(?:((\\w[\\w-.]+)\\/(\\w[\\w-.]+)|\\B))#)([1-9]\\d*)\\b")
)

type IssueID struct {
	Owner  string
	Repo   string
	Number int
}

func (id IssueID) String() string {
	var buffer bytes.Buffer
	if id.Owner != "" {
		buffer.WriteString(id.Owner)
		buffer.WriteString("/")
		buffer.WriteString(id.Repo)
	}
	buffer.WriteString("#")
	buffer.WriteString(strconv.Itoa(id.Number))
	return buffer.String()
}

func IssueIDsFromText(s string) []IssueID {
	var issues []IssueID

	for _, v := range githubIssuesRegexp.FindAllStringSubmatch(s, -1) {
		num, err := strconv.Atoi(v[6])
		if err != nil {
			continue
		}
		issues = append(issues, IssueID{
			Owner:  v[4],
			Repo:   v[5],
			Number: num,
		})
	}

	return issues
}

func (ghr *Repository) IssuesFromText(ctx context.Context, s string) ([]common.Issue, bool, error) {
	var includeBugIssue bool
	ids := IssueIDsFromText(s)
	results := make([]common.Issue, len(ids))

	for i, id := range ids {
		strId := id.String()
		if id.Owner == "" {
			id.Owner = ghr.owner
			id.Repo = ghr.repo
		}

		issue, _, err := ghr.github.Issues.Get(ctx, id.Owner, id.Repo, id.Number)
		if err != nil {
			if err == context.Canceled {
				return nil, false, err
			}

			results[i] = common.Issue{
				Id:      strId,
				Fetched: false,
			}
			continue
		}

		//todo: we could consider the following to improve the linking:
		// 1) if a commit mentioned in the issue
		// 2) use the closing keywords: https://help.github.com/en/github/managing-your-work-on-github/closing-issues-using-keywords
		// 3) if this is a pull request `issue.IsPullRequest()`

		var bug bool
		for i := range issue.Labels {
			// todo: add more labels
			// todo: it will be good if users can customize the labels
			if issue.Labels[i].GetName() == "bug" {
				bug = true
				includeBugIssue = true
				break
			}
		}

		results[i] = common.Issue{
			Id:        strId,
			Fetched:   true,
			Bug:       bug,
			CreatedAt: issue.GetCreatedAt(),
		}
	}

	return results, includeBugIssue, nil
}
