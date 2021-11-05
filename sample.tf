# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    xray = {
      source  = "registry.terraform.io/jfrog/xray"
      version = "0.0.1"
    }
  }
}

provider "xray" {
  //  supply ARTIFACTORY_URL (or JFROG_URL) and ARTIFACTORY_ACCESS_TOKEN (or XRAY_ACCESS_TOKEN) as env vars
}

resource "random_id" "randid" {
  byte_length = 2
}

resource "xray_security_policy" "security1" {
  name        = "test-security-policy-severity-${random_id.randid.dec}"
  description = "Security policy description"
  type        = "security"
  rules {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution = true
      fail_build = true
      notify_watch_recipients = true
      notify_deployer = true
      create_ticket_enabled = false           // set to true only if Jira integration is enabled
      build_failure_grace_period_in_days = 5  // use only if fail_build is enabled
    }
  }
}

resource "xray_security_policy" "security2" {
  name        = "test-security-policy-cvss-${random_id.randid.dec}"
  description = "Security policy description"
  type        = "security"
  rules {
    name     = "rule-name-cvss"
    priority = 1
    criteria {
      cvss_range {
        from = 1.5
        to = 5.3
      }
    }
    actions {
      webhooks = []
      mails = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution = true
      fail_build = true
      notify_watch_recipients = true
      notify_deployer = true
      create_ticket_enabled = false           // set to true only if Jira integration is enabled
      build_failure_grace_period_in_days = 5  // use only if fail_build is enabled
    }
  }
}

resource "xray_license_policy" "license1" {
	name = "test-license-policy-allowed-${random_id.randid.dec}"
	description = "License policy, allow certain licenses"
	type = "license"
	rules {
		name = "License_rule"
		priority = 1
		criteria {
          allowed_licenses = ["Apache-1.0","Apache-2.0"]
          allow_unknown = false
          multi_license_permissive = true
        }
		actions {
          webhooks = []
          mails = ["test@email.com"]
          block_download {
				unscanned = true
				active = true
          }
          block_release_bundle_distribution = false
          fail_build = true
          notify_watch_recipients = true
          notify_deployer = true
          create_ticket_enabled = false           // set to true only if Jira integration is enabled
          custom_severity = "High"
          build_failure_grace_period_in_days = 5  // use only if fail_build is enabled

		}
	}
}

resource "xray_license_policy" "license2" {
  name = "test-license-policy-banned-${random_id.randid.dec}"
  description = "License policy, block certain licenses"
  type = "license"
  rules {
    name = "License_rule"
    priority = 1
    criteria {
      banned_licenses = ["Apache-3.0","Apache-4.0"]
      allow_unknown = false
      multi_license_permissive = false
    }
    actions {
      webhooks = []
      mails = ["test@email.com"]
      block_download {
        unscanned = true
        active = true
      }
      block_release_bundle_distribution = false
      fail_build = true
      notify_watch_recipients = true
      notify_deployer = true
      create_ticket_enabled = false           // set to true only if Jira integration is enabled
      custom_severity = "Medium"
      build_failure_grace_period_in_days = 5  // use only if fail_build is enabled

    }
  }
}