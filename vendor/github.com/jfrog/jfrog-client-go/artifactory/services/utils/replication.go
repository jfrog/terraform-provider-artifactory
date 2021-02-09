package utils

type ReplicationBody struct {
	Username               string `json:"username"`
	Password               string `json:"password"`
	URL                    string `json:"url"`
	CronExp                string `json:"cronExp"`
	RepoKey                string `json:"repoKey"`
	EnableEventReplication bool   `json:"enableEventReplication"`
	SocketTimeoutMillis    int    `json:"socketTimeoutMillis"`
	Enabled                bool   `json:"enabled"`
	SyncDeletes            bool   `json:"syncDeletes"`
	SyncProperties         bool   `json:"syncProperties"`
	SyncStatistics         bool   `json:"syncStatistics"`
	PathPrefix             string `json:"pathPrefix"`
}

type ReplicationParams struct {
	Username string
	Password string
	Url      string
	CronExp  string
	// Source replication repository.
	RepoKey                string
	EnableEventReplication bool
	SocketTimeoutMillis    int
	Enabled                bool
	SyncDeletes            bool
	SyncProperties         bool
	SyncStatistics         bool
	PathPrefix             string
}

func CreateReplicationBody(params ReplicationParams) *ReplicationBody {
	return &ReplicationBody{
		Username:               params.Username,
		Password:               params.Password,
		URL:                    params.Url,
		CronExp:                params.CronExp,
		RepoKey:                params.RepoKey,
		EnableEventReplication: params.EnableEventReplication,
		SocketTimeoutMillis:    params.SocketTimeoutMillis,
		Enabled:                params.Enabled,
		SyncDeletes:            params.SyncDeletes,
		SyncProperties:         params.SyncProperties,
		SyncStatistics:         params.SyncStatistics,
		PathPrefix:             params.PathPrefix,
	}
}
