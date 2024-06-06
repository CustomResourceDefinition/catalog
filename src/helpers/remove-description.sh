yq eval 'del(.. | select(.description? | type == "!!str") | .description)' < "$1" > "$1.tmp"; mv "$1.tmp" "$1"
