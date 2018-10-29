package artifactory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type SecurityService service

type UserDetails struct {
	Name  *string `json:"name,omitempty"`
	Uri   *string `json:"uri,omitempty"`
	Realm *string `json:"realm,omitempty"`
}

func (r UserDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the users list
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ListUsers(ctx context.Context) (*[]UserDetails, *http.Response, error) {
	path := "/api/security/users"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeUsers)

	users := new([]UserDetails)
	resp, err := s.client.Do(ctx, req, users)
	return users, resp, err
}

// application/vnd.org.jfrog.artifactory.security.User+json
type User struct {
	Name                     *string   `json:"name,omitempty"`                     // Optional element in create/replace queries
	Email                    *string   `json:"email,omitempty"`                    // Mandatory element in create/replace queries, optional in "update" queries
	Password                 *string   `json:"password,omitempty"`                 // Mandatory element in create/replace queries, optional in "update" queries
	Admin                    *bool     `json:"admin,omitempty"`                    // Optional element in create/replace queries; Default: false
	ProfileUpdatable         *bool     `json:"profileUpdatable,omitempty"`         // Optional element in create/replace queries; Default: true
	DisableUIAccess          *bool     `json:"disableUIAccess,omitempty"`          // Optional element in create/replace queries; Default: false
	InternalPasswordDisabled *bool     `json:"internalPasswordDisabled,omitempty"` // Optional element in create/replace queries; Default: false
	LastLoggedIn             *string   `json:"lastLoggedIn,omitempty"`             // Read-only element
	Realm                    *string   `json:"realm,omitempty"`                    // Read-only element
	Groups                   *[]string `json:"groups,omitempty"`                   // Optional element in create/replace queries
}

func (r User) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the details of an Artifactory user
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) GetUser(ctx context.Context, username string) (*User, *http.Response, error) {
	path := fmt.Sprintf("/api/security/users/%s", username)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeUser)

	user := new(User)
	resp, err := s.client.Do(ctx, req, user)
	return user, resp, err
}

// Get the encrypted password of the authenticated requestor
// Since: 3.3.0
// Security: Requires a privileged user
func (s *SecurityService) GetEncryptedPassword(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/security/encryptedPassword"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	return String(buf.String()), resp, err
}

// Creates a new user in Artifactory or replaces an existing user
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Missing values will be set to the default values as defined by the consumed type.
// Security: Requires an admin user
func (s *SecurityService) CreateOrReplaceUser(ctx context.Context, username string, user *User) (*http.Response, error) {
	path := fmt.Sprintf("/api/security/users/%s", username)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, user)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Updates an exiting user in Artifactory with the provided user details.
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Missing values will be set to the default values as defined by the consumed type
// Security: Requires an admin user
func (s *SecurityService) UpdateUser(ctx context.Context, username string, user *User) (*http.Response, error) {
	path := fmt.Sprintf("/api/security/users/%s", username)
	req, err := s.client.NewJSONEncodedRequest("POST", path, user)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Removes an Artifactory user.
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) DeleteUser(ctx context.Context, username string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/users/%v", username)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Expires a user's password
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ExpireUserPassword(ctx context.Context, username string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/users/authorization/expirePassword/%s", username)
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Expires password for a list of users
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ExpireMultipleUsersPassword(ctx context.Context, usernames []string) (*http.Response, error) {
	path := "/api/security/users/authorization/expirePasswords"
	req, err := s.client.NewJSONEncodedRequest("POST", path, usernames)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Expires password for all users
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ExpireAllUsersPassword(ctx context.Context) (*http.Response, error) {
	path := "/api/security/users/authorization/expirePasswordForAllUsers"
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Unexpires a user's password
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) UnexpireUserPassword(ctx context.Context, username string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/users/authorization/unexpirePassword/%s", username)
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type PasswordChangeOptions struct {
	Username     *string `json:"username,omitempty"`
	OldPassword  *string `json:"oldPassword,omitempty"`
	NewPassword1 *string `json:"newPassword1,omitempty"`
	NewPassword2 *string `json:"newPassword2,omitempty"`
}

// Changes a user's password
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Admin can apply this method to all users, and each (non-anonymous) user can use this method to change his own password.
func (s *SecurityService) ChangePassword(ctx context.Context, opts *PasswordChangeOptions) (*string, *http.Response, error) {
	path := "/api/security/users/authorization/changePassword"
	req, err := s.client.NewJSONEncodedRequest("POST", path, opts)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type PasswordExpirationPolicy struct {
	Enabled        *bool `json:"enabled,omitempty"`
	PasswordMaxAge *int  `json:"passwordMaxAge,omitempty"`
	NotifyByEmail  *bool `json:"notifyByEmail,omitempty"`
}

// Retrieves the password expiration policy
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) GetPasswordExpirationPolicy(ctx context.Context) (*PasswordExpirationPolicy, *http.Response, error) {
	path := "/api/security/configuration/passwordExpirationPolicy"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(PasswordExpirationPolicy)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Sets the password expiration policy
// Since: 4.4.2
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) SetPasswordExpirationPolicy(ctx context.Context, policy *PasswordExpirationPolicy) (*PasswordExpirationPolicy, *http.Response, error) {
	path := "/api/security/configuration/passwordExpirationPolicy"
	req, err := s.client.NewJSONEncodedRequest("PUT", path, policy)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(PasswordExpirationPolicy)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type UserLockPolicy struct {
	Enabled       *bool `json:"enabled,omitempty"`
	LoginAttempts *int  `json:"loginAttempts,omitempty"`
}

// Retrieves the currently configured user lock policy.
// Since: 4.4
// Security: Requires a valid admin user
func (s *SecurityService) GetUserLockPolicy(ctx context.Context) (*UserLockPolicy, *http.Response, error) {
	path := "/api/security/userLockPolicy"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", mediaTypeJson)

	v := new(UserLockPolicy)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Configures the user lock policy that locks users out of their account if the number of repeated incorrect login attempts exceeds the configured maximum allowed.
// Since: 4.4
// Security: Requires a valid admin user
func (s *SecurityService) SetUserLockPolicy(ctx context.Context, policy *PasswordExpirationPolicy) (*string, *http.Response, error) {
	path := "/api/security/userLockPolicy"
	req, err := s.client.NewJSONEncodedRequest("PUT", path, policy)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// If locking out users is enabled, lists all users that were locked out due to recurrent incorrect login attempts.
// Since: 4.4
// Security: Requires a valid admin user
func (s *SecurityService) GetLockedOutUsers(ctx context.Context) ([]string, *http.Response, error) {
	path := "/api/security/lockedUsers"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	var v []string
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Unlocks a single user that was locked out due to recurrent incorrect login attempts.
// Since: 4.4
// Security:  Requires a valid admin user
func (s *SecurityService) UnlockUser(ctx context.Context, username string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/unlockUsers/%s", username)
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Unlocks a list of users that were locked out due to recurrent incorrect login attempts.
// Since: 4.4
// Security:  Requires a valid admin user
func (s *SecurityService) UnlockMultipleUsers(ctx context.Context, usernames []string) (*string, *http.Response, error) {
	path := "/api/security/unlockUsers"
	req, err := s.client.NewJSONEncodedRequest("POST", path, usernames)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Unlocks all users that were locked out due to recurrent incorrect login attempts.
// Since: 4.4
// Security:  Requires a valid admin user
func (s *SecurityService) UnlockedAllUsers(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/security/unlockAllUsers"
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type ApiKey struct {
	ApiKey *string `json:"apiKey,omitempty"`
}

// Create an API key for the current user. Returns an error if API key already exists - use regenerate API key instead.
// Since: 4.3.0
func (s *SecurityService) CreateApiKey(ctx context.Context) (*ApiKey, *http.Response, error) {
	path := "/api/security/apiKey"
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(ApiKey)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Regenerate an API key for the current user
// Since: 4.3.0
func (s *SecurityService) RegenerateApiKey(ctx context.Context) (*ApiKey, *http.Response, error) {
	path := "/api/security/apiKey"
	req, err := s.client.NewRequest("PUT", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(ApiKey)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Get the current user's own API key
// Since: 4.3.0
func (s *SecurityService) GetApiKey(ctx context.Context) (*ApiKey, *http.Response, error) {
	path := "/api/security/apiKey"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(ApiKey)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Revokes the current user's API key
// Since: 4.3.0
func (s *SecurityService) RevokeApiKey(ctx context.Context) (*map[string]interface{}, *http.Response, error) {
	path := "/api/security/apiKey"
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(map[string]interface{})
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Revokes the API key of another user
// Since: 4.3.0
// Security: Requires a privileged user (Admin only)
func (s *SecurityService) RevokeUserApiKey(ctx context.Context, username string) (*map[string]interface{}, *http.Response, error) {
	path := fmt.Sprintf("/api/security/apiKey/%s", username)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(map[string]interface{})
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Revokes all API keys currently defined in the system
// Since: 4.3.0
// Security: Requires a privileged user (Admin only)
func (s *SecurityService) RevokeAllApiKeys(ctx context.Context) (*map[string]interface{}, *http.Response, error) {
	opt := struct {
		DeleteAll int `json:"deleteAll"`
	}{1}
	path, err := addOptions("/api/security/apiKey", opt)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(map[string]interface{})
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// application/vnd.org.jfrog.artifactory.security.Groups+json
type GroupDetails struct {
	Name *string `json:"name,omitempty"`
	Uri  *string `json:"uri,omitempty"`
}

func (r GroupDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the groups list
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ListGroups(ctx context.Context) (*[]GroupDetails, *http.Response, error) {
	path := "/api/security/groups"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeGroups)

	groups := new([]GroupDetails)
	resp, err := s.client.Do(ctx, req, groups)
	return groups, resp, err
}

// application/vnd.org.jfrog.artifactory.security.Group+json
type Group struct {
	Name            *string `json:"name,omitempty"`            // Optional element in create/replace queries
	Description     *string `json:"description,omitempty"`     // Optional element in create/replace queries
	AutoJoin        *bool   `json:"autoJoin,omitempty"`        // Optional element in create/replace queries; default: false (must be false if adminPrivileges is true)
	AdminPrivileges *bool   `json:"adminPrivileges,omitempty"` // Optional element in create/replace queries; default: false
	Realm           *string `json:"realm,omitempty"`           // Optional element in create/replace queries
	RealmAttributes *string `json:"realmAttributes,omitempty"` // Optional element in create/replace queries
}

func (r Group) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the details of an Artifactory Group
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) GetGroup(ctx context.Context, groupName string) (*Group, *http.Response, error) {
	path := fmt.Sprintf("/api/security/groups/%s", groupName)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeGroup)

	group := new(Group)
	resp, err := s.client.Do(ctx, req, group)
	return group, resp, err
}

// Creates a new group in Artifactory or replaces an existing group
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Missing values will be set to the default values as defined by the consumed type.
// Security: Requires an admin user
func (s *SecurityService) CreateOrReplaceGroup(ctx context.Context, groupName string, group *Group) (*http.Response, error) {
	url := fmt.Sprintf("/api/security/groups/%s", groupName)
	req, err := s.client.NewJSONEncodedRequest("PUT", url, group)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Updates an exiting group in Artifactory with the provided group details.
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) UpdateGroup(ctx context.Context, groupName string, group *Group) (*http.Response, error) {
	path := fmt.Sprintf("/api/security/groups/%s", groupName)
	req, err := s.client.NewJSONEncodedRequest("POST", path, group)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Removes an Artifactory group.
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) DeleteGroup(ctx context.Context, groupName string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/groups/%v", groupName)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// application/vnd.org.jfrog.artifactory.security.PermissionTargets+json
type PermissionTargetsDetails struct {
	Name *string `json:"name,omitempty"`
	Uri  *string `json:"uri,omitempty"`
}

func (r PermissionTargetsDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the permission targets list
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) ListPermissionTargets(ctx context.Context) ([]*PermissionTargetsDetails, *http.Response, error) {
	path := "/api/security/permissions"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePermissionTargets)

	var permissionTargets []*PermissionTargetsDetails
	resp, err := s.client.Do(ctx, req, &permissionTargets)
	return permissionTargets, resp, err
}

type Principals struct {
	Users  *map[string][]string `json:"users,omitempty"`
	Groups *map[string][]string `json:"groups,omitempty"`
}

// application/vnd.org.jfrog.artifactory.security.PermissionTarget+json
// Permissions are set/returned according to the following conventions:
//     m=admin; d=delete; w=deploy; n=annotate; r=read
type PermissionTargets struct {
	Name            *string     `json:"name,omitempty"`            // Optional element in create/replace queries
	IncludesPattern *string     `json:"includesPattern,omitempty"` // Optional element in create/replace queries
	ExcludesPattern *string     `json:"excludesPattern,omitempty"` // Optional element in create/replace queries
	Repositories    *[]string   `json:"repositories,omitempty"`    // Mandatory element in create/replace queries, optional in "update" queries
	Principals      *Principals `json:"principals,omitempty"`      // Optional element in create/replace queries
}

func (r PermissionTargets) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Get the details of an Artifactory Permission Target
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) GetPermissionTargets(ctx context.Context, permissionName string) (*PermissionTargets, *http.Response, error) {
	path := fmt.Sprintf("/api/security/permissions/%s", permissionName)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePermissionTarget)

	permission := new(PermissionTargets)
	resp, err := s.client.Do(ctx, req, permission)
	return permission, resp, err
}

// Creates a new permission target in Artifactory or replaces an existing permission target
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Missing values will be set to the default values as defined by the consumed type.
// Security: Requires an admin user
func (s *SecurityService) CreateOrReplacePermissionTargets(ctx context.Context, permissionName string, permissionTargets *PermissionTargets) (*http.Response, error) {
	path := fmt.Sprintf("/api/security/permissions/%s", permissionName)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, permissionTargets)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Deletes an Artifactory permission target.
// Since: 2.4.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *SecurityService) DeletePermissionTargets(ctx context.Context, permissionName string) (*string, *http.Response, error) {
	path := fmt.Sprintf("/api/security/permissions/%v", permissionName)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Permissions are returned according to the following conventions:
// m=admin; d=delete; w=deploy; n=annotate; r=read
type ItemPermissions struct {
	Uri        *string     `json:"uri,omitempty"`
	Principals *Principals `json:"principals,omitempty"`
}

func (r ItemPermissions) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Returns a list of effective permissions for the specified item (file or folder).
// Only users and groups with some permissions on the item are returned. Supported by local and local-cached repositories.
// Since: 2.3.4
// Notes: Requires Artifactory Pro
// Security: Requires a valid admin or local admin user.
func (s *SecurityService) GetEffectiveItemPermissions(ctx context.Context, repoName string, itemPath string) (*ItemPermissions, *http.Response, error) {
	if !strings.HasPrefix(itemPath, "/") {
		itemPath = itemPath[1:]
	}
	path := fmt.Sprintf("/api/storage/%s/%s?permissions", repoName, itemPath)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeItemPermissions)

	itemPermissions := new(ItemPermissions)
	resp, err := s.client.Do(ctx, req, itemPermissions)
	return itemPermissions, resp, err
}

// Retrieve the security configuration (security.xml).
// Since: 2.2.0
// Notes: This is an advanced feature - make sure the new configuration is really what you wanted before saving.
// Security: Requires a valid admin us
func (s *SecurityService) GetSecurityConfiguration(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/security"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeXml)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Creates a new Artifactory encryption key and activates Artifactory key encryption.
// Since: 3.2.2
// Notes: This is an advanced feature intended for administrators
// Security: Requires a valid admin user
func (s *SecurityService) ActivateArtifactoryKeyEncryption(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/encrypt"
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Removes the current Artifactory encryption key and deactivates Artifactory key encryption.
// Since: 3.2.2
// Notes: This is an advanced feature intended for administrators
// Security: Requires a valid admin user
func (s *SecurityService) DeactivateArtifactoryKeyEncryption(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/decrypt"
	req, err := s.client.NewRequest("POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Gets the public key that Artifactory provides to Debian and Opkg clients to verify packages
// Security: Requires an authenticated user, or anonymous (if "Anonymous Access" is globally enabled)
func (s *SecurityService) GetGPGPublicKey(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/gpg/key/public"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Sets the public key that Artifactory provides to Debian and Opkg clients to verify packages
// Security: Requires a valid admin user
func (s *SecurityService) SetGPGPublicKey(ctx context.Context, gpgKey string) (*string, *http.Response, error) {
	path := "/api/gpg/key/public"
	req, err := s.client.NewRequest("PUT", path, strings.NewReader(gpgKey))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-type", mediaTypePlain)
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Sets the private key that Artifactory will use to sign Debian and ipk packages
// Security: Requires a valid admin user
func (s *SecurityService) SetGPGPrivateKey(ctx context.Context, gpgKey string) (*string, *http.Response, error) {
	path := "/api/gpg/key/private"
	req, err := s.client.NewRequest("PUT", path, strings.NewReader(gpgKey))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-type", mediaTypePlain)
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Sets the pass phrase required signing Debian and ipk packages using the private key
// Security: Requires a valid admin user
func (s *SecurityService) SetGPGPassPhrase(ctx context.Context, passphrase string) (*string, *http.Response, error) {
	path := "/api/gpg/key/passphrase"
	req, err := s.client.NewRequest("PUT", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-GPG-PASSPHRASE", passphrase)
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type AccessTokenOptions struct {
	// The grant type used to authenticate the request. In this case, the only value supported is "client_credentials" which is also the default value if this parameter is not specified.
	GrantType *string `url:"grant_type,omitempty"` // [Optional, default: "client_credentials"]
	// The user name for which this token is created. If the user does not exist, a transient user is created. Non-admin users can only create tokens for themselves so they must specify their own username.
	// If the user does not exist, the member-of-groups scope token must be provided (e.g. member-of-groups: g1, g2, g3...)
	Username *string `url:"username,omitempty"`
	// The scope to assign to the token provided as a space-separated list of scope tokens. Currently there are three possible scope tokens:
	//     - "api:*" - indicates that the token grants access to REST API calls. This is always granted by default whether specified in the call or not.
	//     - member-of-groups:[<group-name>] - indicates the groups that the token is associated with (e.g. member-of-groups: g1, g2, g3...). The token grants access according to the permission targets specified for the groups listed.
	//       Specify "*" for group-name to indicate that the token should provide the same access privileges that are given to the group of which the logged in user is a member.
	//       A non-admin user can only provide a scope that is a subset of the groups to which he belongs
	//     - "jfrt@<instance-id>:admin" - provides admin privileges on the specified Artifactory instance. This is only available for administrators.
	// If omitted and the username specified exists, the token is granted the scope of that user.
	Scope *string `url:"scope,omitempty"` // [Optional if the user specified in username exists]
	// The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin can only set a value that is equal to or less than the default 3600.
	ExpiresIn *int `url:"expires_in,omitempty"` // [Optional, default: 3600]
	// If true, this token is refreshable and the refresh token can be used to replace it with a new token once it expires.
	Refreshable *string `url:"refreshable,omitempty"` // [Optional, default: false]
	// A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs as obtained from the Get Service ID endpoint.
	// In case you want the token to be accepted by all Artifactory instances you may use the following audience parameter "audience=jfrt@*".
	Audience *string `url:"audience,omitempty"` // [Optional, default: Only the service ID of the Artifactory instance that created the token]
}

type AccessToken struct {
	AccessToken  *string `json:"access_token,omitempty"`
	ExpiresIn    *int    `json:"expires_in,omitempty"`
	Scope        *string `json:"scope,omitempty"`
	TokenType    *string `json:"token_type,omitempty"`
	RefreshToken *string `json:"refresh_token,omitempty"`
}

func (r AccessToken) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Creates an access token
// Since: 5.0.0
// Security: Requires a valid user
func (s *SecurityService) CreateToken(ctx context.Context, opts *AccessTokenOptions) (*AccessToken, *http.Response, error) {
	path := "/api/security/token"
	req, err := s.client.NewURLEncodedRequest("POST", path, opts)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	token := new(AccessToken)
	resp, err := s.client.Do(ctx, req, token)
	return token, resp, err
}

type AccessTokenRefreshOptions struct {
	// Should be set to refresh_token.
	GrantType *string `url:"grant_type,omitempty"`
	// The refresh token of the access token that needs to be refreshed.
	RefreshToken *string `url:"refresh_token,omitempty"`
	// The access token to refresh.
	AccessToken *string `url:"access_token,omitempty"`
	// The user name for which this token is created. If the user does not exist, a transient user is created. Non-admin users can only create tokens for themselves so they must specify their own username.
	// If the user does not exist, the member-of-groups scope token must be provided (e.g. member-of-groups: g1, g2, g3...)
	// Note: access_token and username are mutually exclusive, so only one of these parameters should be specified.
	Username *string `url:"username,omitempty"`
	// The scope to assign to the token provided as a space-separated list of scope tokens. Currently there are three possible scope tokens:
	//     - "api:*" - indicates that the token grants access to REST API calls. This is always granted by default whether specified in the call or not.
	//     - member-of-groups:[<group-name>] - indicates the groups that the token is associated with (e.g. member-of-groups: g1, g2, g3...). The token grants access according to the permission targets specified for the groups listed.
	//       Specify "*" for group-name to indicate that the token should provide the same access privileges that are given to the group of which the logged in user is a member.
	//       A non-admin user can only provide a scope that is a subset of the groups to which he belongs
	//     - "jfrt@<instance-id>:admin" - provides admin privileges on the specified Artifactory instance. This is only available for administrators.
	// If omitted and the username specified exists, the token is granted the scope of that user.
	Scope *string `url:"scope,omitempty"`
	// The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin can only set a value that is equal to or less than the default 3600.
	ExpiresIn *int `url:"expires_in,omitempty"`
	// If true, this token is refreshable and the refresh token can be used to replace it with a new token once it expires.
	Refreshable *string `url:"refreshable,omitempty"`
	// A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs as obtained from the Get Service ID endpoint.
	// In case you want the token to be accepted by all Artifactory instances you may use the following audience parameter "audience=jfrt@*".
	Audience *string `url:"audience,omitempty"`
}

// Refresh an access token to extend its validity. If only the access token and the refresh token are provided (and no other parameters), this pair is used for authentication. If username or any other parameter is provided, then the request must be authenticated by a token that grants admin permissions.
// Since: 5.0.0
// Security: Requires a valid user (unless both access token and refresh token are provided)
func (s *SecurityService) RefreshToken(ctx context.Context, opts *AccessTokenRefreshOptions) (*AccessToken, *http.Response, error) {
	path := "/api/security/token"
	req, err := s.client.NewURLEncodedRequest("POST", path, opts)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	token := new(AccessToken)
	resp, err := s.client.Do(ctx, req, token)
	return token, resp, err
}

type AccessTokenRevokeOptions struct {
	Token string `url:"token,omitempty"`
}

// Revoke an access token
// Since: 5.0.0
// Security: Requires a valid user
func (s *SecurityService) RevokeToken(ctx context.Context, opts AccessTokenRevokeOptions) (*string, *http.Response, error) {
	path := "/api/security/token/revoke"
	req, err := s.client.NewURLEncodedRequest("POST", path, opts)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Provides the service ID of an Artifactory instance or cluster. Up to version 5.5.1, the Artiafctory service ID is formatted jf-artifactory@<id>. From version 5.5.2 the service ID is formatted jfrt@<id>.
// Since: 5.0.0
// Security: Requires an admin user
func (s *SecurityService) GetServiceId(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/service_id"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type CertificateDetails struct {
	CertificateAlias *string `json:"certificateAlias,omitempty"`
	IssuedTo         *string `json:"issuedTo,omitempty"`
	IssuedBy         *string `json:"issuedby,omitempty"`
	IssuedOn         *string `json:"issuedOn,omitempty"`
	ValidUntil       *string `json:"validUntil,omitempty"`
	FingerPrint      *string `json:"fingerPrint,omitempty"`
}

func (r CertificateDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Returns a list of installed SSL certificates.
// Since:5.4.0
// Security: Requires an admin user
func (s *SecurityService) GetCertificates(ctx context.Context) (*[]CertificateDetails, *http.Response, error) {
	path := "/api/system/security/certificates"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	certificates := new([]CertificateDetails)
	resp, err := s.client.Do(ctx, req, certificates)
	return certificates, resp, err
}

// Adds an SSL certificate.
// Since:5.4.0
// Security: Requires an admin user
func (s *SecurityService) AddCertificate(ctx context.Context, alias string, pem *os.File) (*Status, *http.Response, error) {
	path := fmt.Sprintf("/api/system/security/certificates/%s", alias)
	req, err := s.client.NewRequest("POST", path, pem)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-type", mediaTypePlain)
	req.Header.Set("Accept", mediaTypeJson)

	status := new(Status)
	resp, err := s.client.Do(ctx, req, status)
	return status, resp, err
}

// Deletes an SSL certificate.
// Since:5.4.0
// Security: Requires an admin user
func (s *SecurityService) DeleteCertificate(ctx context.Context, alias string) (*Status, *http.Response, error) {
	path := fmt.Sprintf("/api/system/security/certificates/%s", alias)
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	status := new(Status)
	resp, err := s.client.Do(ctx, req, status)
	return status, resp, err
}
