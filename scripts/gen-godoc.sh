#!/bin/bash

if [ ! $# -eq 2 ]; then
	echo "Usage: gen-godoc.sh pkg-name output-directory" 1>&2
	exit 1
fi

PKG="$1"
OUT="$2"

set -e

godoc -http=:8081 &
PID=$!

URL=http://localhost:8081/pkg/${PKG}/

cd "$OUT"
set +e
wget -r -m -k -E -p -erobots=off --include-directories="/pkg,/lib" --exclude-directories="*" "$URL"
mv localhost:8081/* .
rmdir localhost:8081

kill $PID

## generate a simple redirect
cat<<EOF > index.html
<html>
<head>
<meta http-equiv="refresh" content="0; url=pkg/${PKG}/index.html" />
</head>
<body>
</body>
EOF
