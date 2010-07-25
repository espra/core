#! /bin/sh

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# NOTE: This script has only been tested in the context of a modern Bash Shell
# on Ubuntu Linux and OS X. Any patches to make it work under alternative Unix
# shells, versions and platforms are very welcome!

if [[ "x$BASH_SOURCE" == "x" ]]; then
	echo "Sorry, this only works under Bash shells atm. Patches welcome... =)"
	exit
fi

_OS_NAME=$(uname -s | tr [[:upper:]] [[:lower:]])
_OS_ARCH=$(uname -m)

$((echo $_OS_ARCH | grep "64") > /dev/null) && _OS_ARCH_64=true
$((echo $_OS_ARCH | grep "i386") > /dev/null) && _OS_ARCH_386=true

# ------------------------------------------------------------------------------
# exit if we're not sourced and echo usage example if possible
# ------------------------------------------------------------------------------

if [ "x$0" == "x$BASH_SOURCE" ]; then
	LSOF=$(lsof -p $$ 2> /dev/null | grep -E "/"$(basename $0)"$")
	case $_OS_NAME in
		darwin)
			__FILE=$(echo $LSOF | sed -E s/'^([^\/]+)\/'/'\/'/1 2>/dev/null);;
		linux)
			__FILE=$(echo $LSOF | sed -r s/'^([^\/]+)\/'/'\/'/1 2>/dev/null);;
		freebsd)
			__FILE=$(echo $LSOF | sed -E s/'^([^\/]+)\/'/'\/'/1 2>/dev/null);;
		*)
			echo "ERROR: You need to source this script and not run it directly!";
			exit
	esac
	echo
	echo "Usage:"
	echo
	echo "    source $__FILE"
	echo
	exit
fi

# ------------------------------------------------------------------------------
# work out if we are running within an appropriate version of bash, i.e. v3.0+
# ------------------------------------------------------------------------------

_BASH_VERSION=${BASH_VERSION%.*} # $BASH_VERSION normally looks something like:
                                 # 3.2b.17(1)-release

_BASH_MAJOR_VERSION=${_BASH_VERSION%.*}
_BASH_MINOR_VERSION=${_BASH_VERSION#*.}

if [ $_BASH_MAJOR_VERSION -le 2 ]; then
	echo "ERROR: You need to be running Bash 3.0+"
	return 1
fi

# ------------------------------------------------------------------------------
# try to determine the absolute path of the enclosing startup + root directory
# ------------------------------------------------------------------------------

cd "$(dirname $BASH_SOURCE)" || return $?

export AMPIFY_STARTUP_DIRECTORY=`pwd -P 2> /dev/null` || return $?

cd $OLDPWD || return $?

export AMPIFY_ROOT=$(dirname $AMPIFY_STARTUP_DIRECTORY)

# ------------------------------------------------------------------------------
# exit if $AMPIFY_ROOT is not set
# ------------------------------------------------------------------------------

if [ "x$AMPIFY_ROOT" == "x" ]; then
	echo "ERROR: Sorry, couldn't detect the Ampify Root Directory."
	return
fi

# ------------------------------------------------------------------------------
# utility funktions
# ------------------------------------------------------------------------------

function _have () {
	unset -v _have
	type $1 &> /dev/null && _have="yes"
}

# ------------------------------------------------------------------------------
# set/extend some core variables
# ------------------------------------------------------------------------------

export AMPIFY_LOCAL=$AMPIFY_ROOT/environ/local

if [ "x$PATH" != "x" ]; then
	export PATH=$AMPIFY_ROOT/environ:$AMPIFY_LOCAL/bin:$AMPIFY_ROOT/src/codereview:$PATH
else
	export PATH=$AMPIFY_ROOT/environ:$AMPIFY_LOCAL/bin:$AMPIFY_ROOT/src/codereview
fi

case $_OS_NAME in
	darwin)
		export PATH=$AMPIFY_ROOT/environ/osx:$PATH;
		export DYLD_FALLBACK_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$AMPIFY_LOCAL/freeswitch/lib:$DYLD_LIBRARY_PATH:$HOME/lib:/usr/local/lib:/lib:/usr/lib;;
	linux)
		export PATH=$AMPIFY_ROOT/environ/linux:$PATH;
		export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$LD_LIBRARY_PATH;;
	freebsd)
		export PATH=$AMPIFY_ROOT/environ/freebsd:$PATH;
		export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$LD_LIBRARY_PATH;;
	*) echo "ERROR: Unknown system operating system: ${_OS_NAME}"
esac
	
if [ "x$PYTHONPATH" != "x" ]; then
	export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/src/zero:$AMPIFY_ROOT/third_party/pylibs:$PYTHONPATH
else
	export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/src/zero:$AMPIFY_ROOT/third_party/pylibs
fi

if [ "x$MANPATH" != "x" ]; then
	export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/man:$MANPATH
else
	export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/man
fi

# ------------------------------------------------------------------------------
# go related variables
# ------------------------------------------------------------------------------

export GOROOT=$AMPIFY_ROOT/third_party/go
export GOBIN=$AMPIFY_LOCAL/bin

if [ "x$_OS_ARCH_64" != "x" ]; then
	export GOARCH="amd64"
else
	if [ "x$_OS_ARCH_386" != "x" ]; then
		export GOARCH="386"
	else
		case _OS_ARCH in
			arm) export GOARCH="arm";;
			*) echo "ERROR: Unknown system architecture: ${_OS_ARCH}"
		esac
	fi
fi

case $_OS_NAME in
	darwin) export GOOS="darwin";;
	freebsd) export GOOS="freebsd";;
	linux) export GOOS="linux";;
	*) echo "ERROR: Unknown system operating system: ${_OS_NAME}"
esac

# ------------------------------------------------------------------------------
# nativeclient related variables
# ------------------------------------------------------------------------------

export NACL_ROOT=$AMPIFY_LOCAL/third_party/nativeclient

# ------------------------------------------------------------------------------
# clean up after ourselves
# ------------------------------------------------------------------------------

unset _OS_NAME _OS_ARCH _OS_ARCH_64 _OS_ARCH_386
unset _BASH_VERSION _BASH_MAJOR_VERSION _BASH_MINOR_VERSION
unset _have
