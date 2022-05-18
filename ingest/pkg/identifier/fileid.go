package identifier

type FileID struct {
	RepoID     RepositoryID `json:"repo_id"   bson:"r"`
	CommitHash Hash         `json:"hash"      bson:"h"`
	FilePath   string       `json:"path"      bson:"p"`
}
