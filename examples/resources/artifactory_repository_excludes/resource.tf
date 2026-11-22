resource "artifactory_repository_excludes" "blocked-images" {
  repo_key = "my-docker-remote"
  patterns = [
    "library/nginx/1.21.0/**",
    "library/alpine/3.14.0/**",
  ]
}
