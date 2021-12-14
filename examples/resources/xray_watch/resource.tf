resource "xray_watch" "all-repos" {
  name        = "all-repos-watch"
  description = "Watch for all repositories, matching the filter"
  active      = true

  watch_resource {
    type = "all-repos"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  assigned_policy {
    name = xray_security_policy.allowed_licenses.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.banned_licenses.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "repository" {
  name        = "repository-watch"
  description = "Watch a single repo or a list of repositories"
  active      = true

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = "your-repository-name"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = "your-other-repository-name"

    filter {
      type  = "package-type"
      value = "Docker"
    }
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "build" {
  name        = "build-watch"
  description = "Watch a single build or a list of builds"
  active      = true

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-build-name"
  }

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-other-build-name"
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}