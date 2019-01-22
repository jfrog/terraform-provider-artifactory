## 1.4.3 (January 22, 2019)
BUG FIXES:
* Fixed setting of passwords in replications and remote repository

NOTES:
* Added integration tests
* Bumped to terraform v0.11.11
* Cleaned up lint checks and build

## 1.4.2 (December 4, 2018)
FEATURES:
* Added docs website [#14]

BUG FIXES:
* Added enable_token_authentication to remote repositories [#18]

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
