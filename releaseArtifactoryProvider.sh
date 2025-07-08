#!/bin/bash
# A script to fetch the latest stable versions and then create a new Git release branch and tag for specific Terraform providers.

# Exit on error, unset var usage, and pipeline errors
set -euo pipefail

# honor non-interactive mode
ASSUME_YES=${ASSUME_YES:-0}
if [[ "${1:-}" == "-y" ]]; then
  ASSUME_YES=1
  shift || true
fi

# --- Function to get the latest stable version from a Git repository ---
get_latest_version() {
  local repo_url="$1"
  # Fetch all tags, sort them by version, and get the latest stable version (not pre-release).
  # We use grep to filter for tags that match the vX.Y.Z pattern, excluding any with hyphens (e.g., v1.2.3-beta).
  local latest_version=$(git ls-remote --tags --refs --sort='-v:refname' "$repo_url" | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*$' | head -n 1)
  
  if [ -z "$latest_version" ]; then
    echo "Version not found"
  else
    # Remove the 'v' prefix for cleaner output
    echo "${latest_version:1}"
  fi
}

# Small helper to confirm an action
confirm() {
  local prompt="$1"
  if [[ "$ASSUME_YES" == "1" ]]; then
    echo "$prompt (auto-yes)"
    return 0
  fi
  echo ""
  read -p "$prompt (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 0
  fi
}

# Determine the default branch name of the current repo
detect_default_branch() {
  git remote show origin 2>/dev/null | sed -n '/HEAD branch/s/.*: //p'
}

# Ensure working tree is clean
ensure_clean_worktree() {
  if ! git diff-index --quiet HEAD --; then
    echo "Your working tree has uncommitted changes."
    confirm "Proceed anyway?"
  fi
}

# Validate version input as SemVer (with optional leading 'v') and normalize to 'vX.Y.Z'
normalize_version() {
  local input="$1"
  if [[ ! "$input" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be SemVer (e.g., 1.2.3 or v1.2.3)." >&2
    exit 1
  fi
  if [[ "$input" =~ ^v ]]; then
    echo "$input"
  else
    echo "v$input"
  fi
}

# Check if a tag already exists locally or remotely
tag_exists() {
  local tag="$1"
  git fetch --tags >/dev/null 2>&1 || true
  if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
    return 0
  fi
  if git ls-remote --tags origin | grep -q "refs/tags/$tag$"; then
    return 0
  fi
  return 1
}

# --- Fetch and Display Latest Stable Versions ---
echo "--- Fetching Latest Stable Provider Versions ---"

# Define the GitHub repositories for each provider.
REPOSITORIES=(
  "jfrog/terraform-provider-artifactory"
)

# Loop through each repository, fetch its latest version, and display it.
for repo in "${REPOSITORIES[@]}"; do
  provider_name=$(basename "$repo")
  repo_url="https://github.com/${repo}"
  latest=$(get_latest_version "$repo_url")
  echo "Latest version for ${provider_name}: v$latest"
done

echo "-------------------------------------"
echo ""

# --- Inputs ---
PROVIDER_NAME="terraform-provider-artifactory"
echo "Using provider: ${PROVIDER_NAME}"

# Read version unless provided via NEW_VERSION env
if [[ -z "${NEW_VERSION:-}" ]]; then
  read -p "Please enter the new version number (e.g., 1.2.3): " NEW_VERSION
fi
NEW_VERSION=$(normalize_version "$NEW_VERSION")

# --- Determine the correct branch to use ---
BRANCH_TO_CHECKOUT=""
case "$PROVIDER_NAME" in
  "terraform-provider-artifactory")
    # Auto-detect default branch; fallback to master
    BRANCH_TO_CHECKOUT="$(detect_default_branch)"
    [[ -z "$BRANCH_TO_CHECKOUT" ]] && BRANCH_TO_CHECKOUT="master"
    ;;
  *)
    echo "Error: Unknown provider name '$PROVIDER_NAME'."
    echo "Known providers are: terraform-provider-artifactory."
    exit 1
    ;;
esac

# Safety checks
ensure_clean_worktree
if tag_exists "$NEW_VERSION"; then
  echo "Error: Tag $NEW_VERSION already exists locally or on origin." >&2
  exit 1
fi

echo "--- Starting release process for provider '${PROVIDER_NAME}' and version ${NEW_VERSION} ---"

# --- Git Workflow ---
# 1. Checkout the correct base branch.
echo "About to checkout branch '${BRANCH_TO_CHECKOUT}'..."
confirm "Proceed to checkout '${BRANCH_TO_CHECKOUT}'?"
git checkout "${BRANCH_TO_CHECKOUT}"

# 2. Pull the latest code.
echo "About to pull latest code from '${BRANCH_TO_CHECKOUT}'..."
confirm "Proceed to pull from '${BRANCH_TO_CHECKOUT}'?"
git pull --ff-only

# 3. Checkout a new branch for the release.
echo "About to create and checkout new release branch: ${NEW_VERSION}..."
confirm "Proceed to create branch '${NEW_VERSION}'?"
git checkout -b "${NEW_VERSION}"

# 4. Push the new branch to the remote repository.
echo "About to push new branch to origin: ${NEW_VERSION}..."
confirm "Proceed to push branch '${NEW_VERSION}' to origin?"
git push -u origin "${NEW_VERSION}"

# 5. Create a new tag from the new branch.
echo "About to create new tag: ${NEW_VERSION}..."
confirm "Proceed to create tag '${NEW_VERSION}'?"
git tag "${NEW_VERSION}"

# 6. Push the new tag to the remote repository.
echo "About to push new tag to origin: ${NEW_VERSION}..."
confirm "Proceed to push tag '${NEW_VERSION}' to origin?"
git push origin tag "${NEW_VERSION}"

echo ""
echo "--- Release process completed successfully for ${PROVIDER_NAME}! ---"

