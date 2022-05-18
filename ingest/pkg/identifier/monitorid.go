package identifier

type MonitorID struct {
	RepoID RepositoryID `json:"repo_id"  bson:"r"`
	UserID UserID       `json:"user_id"  bson:"u"`
}
