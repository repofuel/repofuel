package manage

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	mock_entity "github.com/repofuel/repofuel/ingest/internal/mock/entity"
	mock_providers "github.com/repofuel/repofuel/ingest/internal/mock/providers"
	"github.com/repofuel/repofuel/ingest/pkg/engine"
	"github.com/repofuel/repofuel/ingest/pkg/engine/git2go"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/pkg/common"
)

func TestMarkPullRequestCommits(t *testing.T) {
	t.Skip("skip to avoid git cloning")

	ctx := context.Background()
	repoID := identifier.RepositoryID{}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repo := engine.NewRepository(repoID, "../../repos/tests/fastapi", &engine.RepositoryOpts{
		OriginURL: "https://github.com/emadshihab/fastapi",
		Adapter:   git2go.NewAdapter(),
		Source:    mock_providers.NewMockSourceIntegration(mockCtrl),
	})

	err := repo.Open()
	if err != nil {
		err := repo.Clone(ctx)
		if err != nil {
			t.Error(err)
		}
	}

	mockReposDB := mock_entity.NewMockRepositoryDataSource(mockCtrl)
	mockReposDB.EXPECT().SaveCommitsCount(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockReposDB.EXPECT().Branches(gomock.Any(), gomock.Any())

	mockCommitItr := mock_entity.NewMockCommitIter(mockCtrl)
	mockCommitItr.EXPECT().ForEach(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockCommitsDB := mock_entity.NewMockCommitDataSource(mockCtrl)
	mockCommitsDB.EXPECT().ReTagBranch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockCommitsDB.EXPECT().FindRepoCommits(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockCommitItr, nil)
	mockCommitsDB.EXPECT().ReTagPullRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Len(5)).Return(nil)

	_, err = ingestLocalBranches(ctx, mockReposDB, mockCommitsDB, repo)
	if err != nil {
		t.Error(err)
	}

	err = MarkPullRequestCommits(ctx, mockCommitsDB, &entity.PullRequest{
		Source: common.PullRequest{
			Head: &common.Branch{
				SHA: "2c014659ded2d0b11d2ee0a9d1fb3f3f5a255f6e",
			},
			Base: &common.Branch{
				SHA: "cad4c8cae061507942ee3976253311770061487f",
			},
		},
	}, repo)
	if err != nil {
		t.Error(err)
	}
}
