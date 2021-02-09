package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory/client"
	"net/http"
)

type SecurityService Service

// read, write, annotate, delete, manage
const (
	PERM_READ     = "read"
	PERM_WRITE    = "write"
	PERM_ANNOTATE = "annotate"
	PERM_DELETE   = "delete"
	PERM_MANAGE   = "manage"

	PERMISSION_SCHEMA = "application/vnd.org.jfrog.artifactory.security.PermissionTargetV2+json"
)

type Entity struct {
	Users  *map[string][]string `json:"users,omitempty"`
	Groups *map[string][]string `json:"groups,omitempty"`
}

type Permission struct {
	IncludePatterns *[]string `json:"include-patterns,omitempty"`
	ExcludePatterns *[]string `json:"exclude-patterns,omitempty"`
	Repositories    *[]string `json:"repositories,omitempty"`
	Actions         *Entity   `json:"actions,omitempty"`
}

type PermissionTarget struct {
	Name  *string     `json:"name,omitempty"` // Optional element in create/replace queries
	Repo  *Permission `json:"repo,omitempty"`
	Build *Permission `json:"build,omitempty"`
}

func (r PermissionTarget) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

func (s *SecurityService) CreatePermissionTarget(ctx context.Context, permissionName string, permissionTargets *PermissionTarget) (*http.Response, error) {
	path := fmt.Sprintf("/api/v2/security/permissions/%s", permissionName)
	req, err := s.client.NewJSONEncodedRequest(http.MethodPost, path, permissionTargets)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

func (s *SecurityService) GetPermissionTarget(ctx context.Context, permissionName string) (*PermissionTarget, *http.Response, error) {
	path := fmt.Sprintf("/api/v2/security/permissions/%s", permissionName)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", client.MediaTypeJson)

	permission := new(PermissionTarget)
	resp, err := s.client.Do(ctx, req, permission)
	return permission, resp, err
}

func (s *SecurityService) HasPermissionTarget(ctx context.Context, permissionName string) (bool, error) {
	path := fmt.Sprintf("/api/v2/security/permissions/%s", permissionName)
	req, err := s.client.NewRequest(http.MethodHead, path, nil)
	if err != nil {
		return false, err
	}

	resp, err := s.client.Do(ctx, req, nil)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// Missing permission target values will be set to the default values as defined by the consumed type.
// The values defined in the request payload will replace what currently exists in the permission target entity.
// In case the request is missing one of the permission target entities (repo/build), the entity will be deleted.
// This means that if an update request is sent to an entity that contains both repo and build, with only repo,
// the build values will be removed from the entity.
func (s *SecurityService) UpdatePermissionTarget(ctx context.Context, permissionName string, permissionTargets *PermissionTarget) (*http.Response, error) {
	path := fmt.Sprintf("/api/v2/security/permissions/%s", permissionName)
	req, err := s.client.NewJSONEncodedRequest(http.MethodPut, path, permissionTargets)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

func (s *SecurityService) DeletePermissionTarget(ctx context.Context, permissionName string) (*http.Response, error) {
	path := fmt.Sprintf("/api/v2/security/permissions/%v", permissionName)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
