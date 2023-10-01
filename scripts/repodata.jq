group_by(.packageType == "docker" and .rclass == "local") |
(.[0] | map({
		to: "artifactory_\(.rclass)_\(.packageType)_repository.\(.key)",
		key
	})
) +
(.[1] | map({
		to: "artifactory_\(.rclass | ascii_downcase)_\(.packageType | ascii_downcase)_\(.dockerApiVersion | ascii_downcase)_repository.\(.key)",
		key
	})
) | .[] |
"import {
  to = \(.to)
  id = \"\(.key)\"
}"
