#!/usr/bin/env bash


function importRepos {
  while read -r key type package; do
	  cat <<-EOF
		import {
		  to = artifactory_${type,,}_${package,,}_repository.${key,,}
		  id =  "${key,,}"
		}
		EOF
	done < <(curl -snLf https://partnerenttest.jfrog.io/artifactory/api/repositories | jq -re '.[] | "\(.key) \(.type) \(.packageType)"'	)
} && export -f importRepos

function importUsers {
  for i in {1..10}; do
	  local username="username-${RANDOM}-${i}"
	  cat <<-EOF
		import {
		  to = artifactory_user.${username}
		  id = %s
		}
		EOF
	done
}


resources=(importRepos importUsers importRepos)
for f in $(echo "${resources[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '); do
  eval "${f}"
done


