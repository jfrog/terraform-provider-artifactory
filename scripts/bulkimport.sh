#!/usr/bin/env bash
${DEBUG:+set -x}
set -e
set -o errexit -o pipefail -o noclobber -o nounset

HOME=${HOME:-"~"}

command -v curl >/dev/null || (echo "You must install curl to proceed" >&2 && exit 1)
command -v jq >/dev/null || (echo "You must install jq to proceed" >&2 && exit 1)
command -v terraform >/dev/null || (echo "You must install terraform to proceed" >&2 && exit 1)

function usage {
	echo "${0} [{--users |-u} {--repos|-r} | {--groups|-g}] -h|--host https://\${host}" >&2
	echo "duplicate resource declarations are de duped (see below)"
	echo "example: ${0} -u --repos --users -h https://myartifactory.com > import.tf"
	exit 1
}

resources=()
while getopts urdga:-: OPT; do
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
	a | artifactory-url)
		host="${OPTARG}"
		grep -qE 'http(s)?://.*' <<<"${host}" || (echo "malformed url name: ${host}. must be of the form http(s)?://.*" && exit 1)
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

function hasNetRcEntry {
	local h="${1:?No host supplied}"

	h="${h/https:\/\//}"
  h="${h/http:\/\//}"
  grep -qE "machine[ ]+${h}" "$(assert_netrc)"
}

function write_netrc {
	local host="${1:?You must supply a host name}"
	local netrc
	netrc=$(assert_netrc)
	read -r -p "please enter the username for ${host}: " username
	read -rs -p "enter the api token (will not be echoed): " token
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
	echo "added entry to $(write_netrc "${host}" )  needed for curl" >&2
fi

function repos {
	# jq '.resources |map({type,name})' terraform.tfstate # make sure to not include anything already in state
	# we'll make our internal jq structure match that of the tf state file so we can subtract them easy
	local host="${1:?You must supply the artifactory host}"
	# literally, usage of jq is 6x faster than bash/read/etc
	# GET "${host}/artifactory/api/repositories" returns {key,type,packageType,url} where
	# url) points to the UI for that resource??
	# packageType) is cased ??
	# type) is upcased and in "${host}/artifactory/api/repositories/${key}" it's not AND it's called rclass
	local url="${host}/artifactory/api/repositories"
	curl -snLf "${url}" | jq -re  --arg u "${url}" '.[] | "\($u)/\(.key)"' |
		xargs -P 10 curl -snLf |
			jq -sre '
				group_by(.packageType == "docker" and .rclass == "local") |
				(.[0] | map({
						type: "artifactory_\(.rclass)_\(.packageType)_repository.\(.key)",
						name: key
					})
				) +
				(.[1] | map({
						type: "artifactory_\(.rclass | ascii_downcase)_\(.packageType | ascii_downcase)_\(.dockerApiVersion | ascii_downcase)_repository.\(.key)",
						name: key
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
	curl -snLf "${host}/artifactory/api/repositories/artifactory/api/security/token"
}

function ldapGroups {
	local host="${1:?You must supply the artifactory host}"
	return 1
	curl -snLf "${host}/access/api/v1/ldap/groups"
}

function apiKeys {
	local host="${1:?You must supply the artifactory host}"
	return 1
	curl -snLf "${host}/artifactory/api/security/apiKey"

}

function groups {
	local host="${1:?You must supply the artifactory host}"
	curl -snLf "${host}/artifactory/api/security/groups" |
		jq -re '.[].name |
  "import {
    to = artifactory_group.\(.)
    id = \"\(.)\"
}"'
}

function certificates {
	local host="${1:?You must supply the artifactory host}"
	curl -snLf "${host}/artifactory/api/system/security/certificates/" |
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
	curl -snLf "${host}/artifactory/api/security/keys/trusted" |
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
	curl -snLf "${host}/artifactory/api/v2/security/permissions/" |
		jq -re '.[] | select(.name | startswith("INTERNAL") | not) | "
import {
	to = artifactory_permission_target.\(.name)
	id = \"\(.name)\"
}"'
}
function keyPairs {
	local host="${1:?You must supply the artifactory host}"
	return 1 # untested
	curl -snLf "${host}/artifactory/api/security/keypair/" |
		jq -re '.[] | "
import {
  to = artifactory_keypair.\(.pairName)
  id = \"\(.pairName)\"
}"'
}
function users {
	#	.name has values in it that artifactory will never accept, like email@. Not sure if in that case it should just be user-$RANDOM
	local host="${1:?You must supply the artifactory host}"
	curl -snLf "${host}/artifactory/api/security/users" | jq -re '.[] |
	{user: .name | capture("(?<user>\\w+)@(?<domain>\\w+)").user, name}|
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

# shellcheck disable=SC2046
output "${host}" $(echo "${resources[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' ')
# out="$RANDOM-out"
#terraform plan -generate-config-out generated.tf -out "${out} -parallelism=10
#terraform apply -parallelism=10
