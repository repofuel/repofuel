package loginwith2

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v30/github"
	"github.com/repofuel/repofuel/pkg/common"
	"github.com/suhaibmujahid/go-bitbucket-server/bitbucket"
)

var defaultBaseURL, _ = url.Parse("https://api.github.com/")

func FetchGithubUser(provider string, baseURL *url.URL) common.FetchAuthUserFunc {
	return func(ctx context.Context, client *http.Client) (*common.User, error) {
		gh := github.NewClient(client)
		if baseURL.Host == "github.com" {
			gh.BaseURL = defaultBaseURL
		} else {
			gh.BaseURL, _ = baseURL.Parse("/api/v3/")
		}

		user, _, err := gh.Users.Get(ctx, "")
		if err != nil {
			return nil, err
		}

		return &common.User{
			Provider:  provider,
			ID:        user.GetNodeID(),
			Username:  user.GetLogin(),
			FullName:  user.GetName(),
			AvatarURL: user.GetAvatarURL(),
			Location:  user.GetLocation(),
			HomePage:  user.GetHTMLURL(),
		}, nil
	}
}

func FetchBitbucketServerUserFunc(provider string, baseURL string) common.FetchAuthUserFunc {
	return func(ctx context.Context, client *http.Client) (*common.User, error) {
		c, err := bitbucket.NewServerClient(baseURL, client)
		if err != nil {
			return nil, err
		}

		user, _, err := c.Users.Myself(ctx)
		if err != nil {
			return nil, err
		}

		return &common.User{
			Provider:  provider,
			ID:        strconv.Itoa(user.Id),
			Username:  user.Slug,
			FullName:  user.DisplayName,
			AvatarURL: "",
			Location:  "",
			HomePage:  user.Links.Self[0].Href,
		}, nil
	}
}

func FetchJiraUserFunc(provider string, baseURL string) common.FetchAuthUserFunc {
	return func(i context.Context, client *http.Client) (*common.User, error) {
		c, err := jira.NewClient(client, baseURL)
		if err != nil {
			return nil, err
		}

		user, _, err := c.User.GetSelf()
		if err != nil {
			return nil, err
		}

		id := user.AccountID
		if id == "" {
			id = user.Key
		}
		return &common.User{
			Provider:  provider,
			ID:        id,
			FullName:  user.DisplayName,
			AvatarURL: user.AvatarUrls.Four8X48,
		}, nil
	}
}
