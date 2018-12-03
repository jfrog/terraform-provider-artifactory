## 1.5.0 (Unreleased)

FEATURES:
* Added docs website

## 1.4.1 (November 6, 2018)
BREAKING CHANGES:
* Removed makefile and githooks in favour of default go tool commands.

FEATURES:
* Add support for supplying user password on creation via the environment variable:
  
  ```TF_USER_${artifactory_username_here}_PASSWORD```
    
  This variable may change and is not tracked by terraform

NOTES:
* Removed formatting checks from CI and streamlined build.
* Migrate from dep to go modules. This is transparent for consumers but requires go 1.11+ for development.
