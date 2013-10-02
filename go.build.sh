#!/bin/bash

export GOPATH="$PWD"/.gopath/

function gothing {
	package="$1"; shift
	go build -race -o "bin/${package%.go}" "main/$package" "$@"
}

for x in `ls main`; do
	echo "building $x ..."
	echo "=================="
	gothing "$x"
	echo
done
echo "done!"
