package artifactory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ArtifactService exposes the Artifact API endpoints from Artifactory
type ArtifactService service

// SingleReplicationConfig is the model of the Artifactory Replication Config
type SingleReplicationConfig struct {
	RepoKey                *string `json:"repoKey,omitempty"`
	URL                    *string `json:"url,omitempty"`
	SocketTimeoutMillis    *int    `json:"socketTimeoutMillis,omitempty"`
	Username               *string `json:"username,omitempty"`
	Password               *string `json:"password,omitempty"`
	Enabled                *bool   `json:"enabled,omitempty"`
	SyncDeletes            *bool   `json:"syncDeletes,omitempty"`
	SyncProperties         *bool   `json:"syncProperties,omitempty"`
	SyncStatistics         *bool   `json:"syncStatistics,omitempty"`
	PathPrefix             *string `json:"pathPrefix,omitempty"`
	CronExp                *string `json:"cronExp,omitempty"` // Only required when getting list of repositories as C*UD operations will be done through a repConfig obj
	EnableEventReplication *bool   `json:"enableEventReplication,omitempty"`
}

// ReplicationConfig is the model for the multi replication config API endpoints. Its usage is preferred over
// SingleReplicationConfig as it is a more direct mapping of the replicationConfig in the UI
type ReplicationConfig struct {
	RepoKey                *string                    `json:"-"`
	CronExp                *string                    `json:"cronExp,omitempty"`
	EnableEventReplication *bool                      `json:"enableEventReplication,omitempty"`
	Replications           *[]SingleReplicationConfig `json:"replications,omitempty"`
}

func (r ReplicationConfig) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Add or replace replication configuration for given repository key. Supported by local and remote repositories. Accepts the JSON payload returned from Get Repository Replication Configuration for a single and an array of configurations. If the payload is an array of replication configurations, then values for cronExp and enableEventReplication in the first element in the array will determine the corresponding values when setting the repository replication configuration.
// Notes: Requires Artifactory Pro
// Security: Requires a privileged user
func (s *ArtifactService) SetRepositoryReplicationConfig(ctx context.Context, repoKey string, config *ReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/multiple/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Returns the replication configuration for the given repository key, if found. Supported by local and remote repositories. Note: The 'enableEventReplication' parameter refers to both push and pull replication.
// Notes: Requires Artifactory Pro
// Security: Requires a privileged user
func (s *ArtifactService) GetRepositoryReplicationConfig(ctx context.Context, repoKey string) (*ReplicationConfig, *http.Response, error) {
	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeReplicationConfig)

	replications := make([]SingleReplicationConfig, 0)
	resp, err := s.client.Do(ctx, req, &replications)

	if err != nil {
		return nil, resp, err
	}

	replicationConfig := new(ReplicationConfig)

	if len(replications) > 0 {
		replicationConfig.Replications = new([]SingleReplicationConfig)
	}

	for _, replication := range replications {
		replicationConfig.RepoKey = replication.RepoKey
		replicationConfig.CronExp = replication.CronExp
		replicationConfig.EnableEventReplication = replication.EnableEventReplication

		*replicationConfig.Replications = append(*replicationConfig.Replications, replication)
	}

	return replicationConfig, resp, nil
}

// Update existing replication configuration for given repository key, if found. Supported by local and remote repositories.
// Notes: Requires Artifactory Pro
// Security: Requires a privileged user
func (s *ArtifactService) UpdateRepositoryReplicationConfig(ctx context.Context, repoKey string, config *ReplicationConfig) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/multiple/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("POST", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *ArtifactService) DeleteRepositoryReplicationConfig(ctx context.Context, repoKey string) (*http.Response, error) {
	path := fmt.Sprintf("/api/replications/%s", repoKey)
	req, err := s.client.NewJSONEncodedRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
