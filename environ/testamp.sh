#! /usr/bin/env bash

# Public Domain (-) 2012 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

AMPIFY_ROOT=$(dirname $(dirname $0))

if [[ "x$AMPIFY_ROOT" == "x" ]]; then
        echo "ERROR: Sorry, couldn't detect the Ampify Root Directory."
fi

. $AMPIFY_ROOT/environ/ampenv.sh

_meta_status=0

# Test the Go packages.
cd $AMPIFY_ROOT/src/amp
for _package in `find . -mindepth 1 -maxdepth 1 -not -empty -type d | sed 's%./%amp/%'`; do
	if [ "x$_package" != "xamp/cmd" ]; then
		go test -v $_package
		_cur_status=$?
		if [ $_cur_status -ne 0 ]; then
	        _meta_status=$_cur_status
		fi
	fi
done

exit $_meta_status
