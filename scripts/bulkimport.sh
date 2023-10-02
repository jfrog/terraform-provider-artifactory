#!/usr/bin/env bash
${DEBUG:+set -x}
set -e
set -o errexit -o pipefail -o noclobber -o nounset
shopt -s expand_aliases

# shellcheck disable=SC2139
alias curl="curl -snL${DEBUG:+v}f"

HOME=${HOME:-"~"}

command -v curl >/dev/null || (echo "You must install curl to proceed" >&2 && exit 1)
command -v jq >/dev/null || (echo "You must install jq to proceed" >&2 && exit 1)
command -v terraform >/dev/null || (echo "You must install terraform to proceed" >&2 && exit 1)

function usage {
	echo "${0} [{--users |-u} {--repos|-r} | {--groups|-g} | {--all | -a}] -h|--host https://\${host}
	duplicate resource declarations are de duped (see below)

	example:
	${0} -u --repos --users -h https://myartifactory.com > import.tf
	terraform plan -no-color -generate-config-out generated.tf -out ${RANDOM}-out -parallelism=10
  terraform apply -no-color -parallelism=10

  You may enable debug with: DEBUG=1 ${0} ..." >&2
	exit 1
}

resources=()
while getopts urdgah:-: OPT; do
	# shellcheck disable=SC2154
	if [ "$OPT" = "-" ]; then # long option: reformulate OPT and OPTARG
		OPT="${OPTARG%%=*}"      # extract long option name
		OPTARG="${OPTARG#$OPT}"  # extract long option argument (may be empty)
		OPTARG="${OPTARG#=}"     # if long option argument, remove assigning `=`
	fi
	# shellcheck disable=SC2214
	case "${OPT}" in
	u | users)
		resources+=(users)
		;;
	r | repos)
		resources+=(repos)
		;;
	g | groups)
		resources+=(groups)
		;;
	h | host)
		host="${OPTARG}"
		;;
	a | all)
		resources=(users repos groups)
		;;
	??*)
		usage
		;;
	*)
		usage
		;;
	esac
done

function netrc_location {
		case "$(uname)" in
  		CYGWIN* | MSYS* | MINGW*)
  			echo "$HOME/_netrc"
  			;;
  		Darwin* | Linux*)
  			echo "$HOME/.netrc"
  			;;
  		*)
  			echo "unsupported OS" >&2
  			return 255
  			;;
  	esac
}
function assert_netrc {
	local location
	location="$(netrc_location)"
	test -f "${location}" || touch "${location}"
	echo "${location}"
}

function toHost {
	local host="${1:?You must supply a host}"
	host="${host/https:\/\//}"
	echo "${host/http:\/\//}"
}
function hasNetRcEntry {
	local h="${1:?No host supplied}"
  grep -qE "machine[ ]+${h}" "$(toHost "${h}")"
}

function write_netrc {
	local host="${1:?You must supply a host name}"
	local netrc
	netrc=$(assert_netrc)
	host=$(toHost "${host}")
	read -r -p "Please enter the username for ${host}: " username
	read -rs -p "Enter the api token (will not be echoed): " token
	# append only to both files
	cat <<-EOF  >> "${netrc}"
		machine ${host}
		login ${username}
		password ${token}
	EOF
	echo "${host}"
}

# if they have no netrc file at all, create the file and add an entry
if ! hasNetRcEntry "${host}" ; then
	cat <<-EOF >&2

	added entry
	to $(netrc_location)
	for $(write_netrc "${host}" )"
	EOF
fi

function repos {
	local host="${1:?You must supply the artifactory host}"
	# literally, usage of jq is 6x faster than bash/read/etc
	curl  "${host}/artifactory/api/repositories" |
		jq -re 'map({
					 		key,
					 		type: (.type | ascii_downcase),
					 		packageType: (.packageType | ascii_downcase)
					 }) |
					 group_by(.packageType == "docker" and .type == "local") |
						(.[0] | map({
								type: "artifactory_\(.type)_\(.packageType)_repository.\(.key)",
								name: .key
							})
						) +
						(.[1] | map({
								type: "artifactory_local_docker_v2_repository.\(.key)",
								name: .key
							})
						) | .[] |
"import {
  to = \(.type)
  id = \"\(.name)\"
}"'

} && export -f repos

function accessTokens {
	local host="${1:?You must supply the artifactory host}"
	return 1
	curl "${host}/artifactory/api/repositories/artifactory/api/security/token"
}

function ldapGroups {
	local host="${1:?You must supply the artifactory host}"
	return 1
	curl "${host}/access/api/v1/ldap/groups"
}

function apiKeys {
	local host="${1:?You must supply the artifactory host}"
	return 1
	curl "${host}/artifactory/api/security/apiKey"

}

function groups {
	local host="${1:?You must supply the artifactory host}"
	curl "${host}/artifactory/api/security/groups" |
		jq -re '.[].name |
"import {
  to = artifactory_group.\(. | ascii_downcase)
  id = \"\(.)\"
}"'
}

function certificates {
	local host="${1:?You must supply the artifactory host}"
	curl "${host}/artifactory/api/system/security/certificates/" |
		jq -re '.[] |
"import {
	to = artifactory_certificate.\(.certificateAlias)
	id = \"(.certificateAlias)\"
}"'
}

function distributionPublicKeys {
	local host="${1:?You must supply the artifactory host}"
	# untested
	return 1
	curl "${host}/artifactory/api/security/keys/trusted" |
	jq -re '.keys[] | "
import {
  to = artifactory_distribution_public_key.\(.alias)
  id = \"\(.kid)\"
}"'
}

function permissions {
	#	these names have spaces in them
	local host="${1:?You must supply the artifactory host}"
	return 1 # untested
	curl "${host}/artifactory/api/v2/security/permissions/" |
		jq -re '.[] | select(.name | startswith("INTERNAL") | not) | "
import {
	to = artifactory_permission_target.\(.name)
	id = \"\(.name)\"
}"'
}
function keyPairs {
	local host="${1:?You must supply the artifactory host}"
	return 1 # untested
	curl "${host}/artifactory/api/security/keypair/" |
		jq -re '.[] | "
import {
  to = artifactory_keypair.\(.pairName)
  id = \"\(.pairName)\"
}"'
}
function users {
	#	.name has values in it that artifactory will never accept, like email@. Not sure if in that case it should just be user-$RANDOM
	local host="${1:?You must supply the artifactory host}"
	curl "${host}/artifactory/api/security/users" | jq -re '.[] |
		{
			user: .name | capture("(?<user>\\w+)@(?<domain>\\w+)").user,
			name
		}|
	"import {
  to = artifactory_user.\(.user)
  id = \"\(.name)\"
}"'
}


function output {
	local host="${1:?You must supply artifactory host name}"
# don't touch this heredoc if you want proper output format
	cat <<-EOF
		terraform {
		  required_providers {
		    artifactory = {
		      source  = "registry.terraform.io/jfrog/artifactory"
		      version = ">= 9.1.0"
		    }
		  }
		}
		provider "artifactory" {
		  url = "${host}"
		}

		$(for f in "${@:2}"; do
			eval "${f} ${host}"
		done)
	EOF
}
if ! grep -qE 'http(s)?://.*' <<<"${host}"; then
	echo "malformed url name: ${host}. must be of the form http(s)?://.*"
	exit 1
fi

# shellcheck disable=SC2046
output "${host}" $(echo "${resources[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' ')
out="$RANDOM-out"
# everytime I try to automate this, I get 'This character is not used within the language.' - it's generating something funky
echo "please run :

terraform plan -no-color -generate-config-out generated.tf -out ${out} -parallelism=10
terraform apply -no-color -parallelism=10
" >&2
