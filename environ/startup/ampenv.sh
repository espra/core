#! /bin/sh

# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# NOTE: This script has only been tested in the context of a modern Bash Shell
# on Ubuntu Linux and OS X. Any patches to make it work under alternative Unix
# shells, versions and platforms are very welcome!

_OS_NAME=$(uname -s | tr [[:upper:]] [[:lower:]])
_OS_ARCH=$(uname -m)

$((echo $_OS_ARCH | grep "64") > /dev/null) && _OS_ARCH_64=true
$((echo $_OS_ARCH | grep "i386") > /dev/null) && _OS_ARCH_386=true

# ------------------------------------------------------------------------------
# exit if we're not sourced and echo usage example if possible
# ------------------------------------------------------------------------------

if [ "$0" == "$BASH_SOURCE" ]; then
	LSOF=$(lsof -p $$ 2> /dev/null | grep -E "/"$(basename $0)"$")
	case $_OS_NAME in
		darwin)
			__FILE=$(echo $LSOF | sed -E s/'^([^\/]+)\/'/'\/'/1 2>/dev/null);;
		linux)
			__FILE=$(echo $LSOF | sed -r s/'^([^\/]+)\/'/'\/'/1 2>/dev/null);;
		*)
			echo "ERROR: You need to source this script and not run it directly!";
			exit
	esac
	echo
	echo "Usage:"
	echo
	echo "    $ source $__FILE"
	echo
	echo "You might want to add it to your login .bash*/profile/etc."
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
# try to figure out if we are inside an interactive shell or not
# ------------------------------------------------------------------------------

test "$PS1" && _INTERACTIVE_SHELL=true;

# ------------------------------------------------------------------------------
# try to determine the absolute path of the enclosing startup + root directory
# ------------------------------------------------------------------------------

cd "$(dirname $BASH_SOURCE)" || return $?

export AMPIFY_STARTUP_DIRECTORY=`pwd -P 2> /dev/null` || return $?

cd $OLDPWD || return $?

export AMPIFY_ROOT=$(dirname $(dirname $AMPIFY_STARTUP_DIRECTORY))

# ------------------------------------------------------------------------------
# exit if $AMPIFY_ROOT is not set
# ------------------------------------------------------------------------------

if [ ! "$AMPIFY_ROOT" ]; then
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

if [ "$PATH" ]; then
	export PATH=$AMPIFY_ROOT/environ/startup:$AMPIFY_LOCAL/bin:$AMPIFY_ROOT/misc/codereview:$PATH
else
	export PATH=$AMPIFY_ROOT/environ/startup:$AMPIFY_LOCAL/bin:$AMPIFY_ROOT/misc/codereview
fi

case $_OS_NAME in
	darwin)
		export PATH=$AMPIFY_ROOT/environ/client/osx:$PATH;
		export DYLD_FALLBACK_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$AMPIFY_LOCAL/freeswitch/lib:$DYLD_LIBRARY_PATH:$HOME/lib:/usr/local/lib:/lib:/usr/lib;;
	linux)
		export PATH=$AMPIFY_ROOT/environ/client/linux:$PATH;
		export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$LD_LIBRARY_PATH;;
	*) echo "ERROR: Unknown system operating system: ${_OS_NAME}"
esac
	
if [ "$PYTHONPATH" ]; then
	export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/third_party/pylibs:$PYTHONPATH
else
	export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/third_party/pylibs
fi

if [ "$MANPATH" ]; then
	export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/man:$MANPATH
else
	export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/man
fi

# ------------------------------------------------------------------------------
# go related variables
# ------------------------------------------------------------------------------

export GOROOT=$AMPIFY_ROOT/third_party/go
export GOBIN=$AMPIFY_LOCAL/bin

if [ "$_OS_ARCH_64" ]; then
	export GOARCH="amd64"
else
	if [ "$_OS_ARCH_386" ]; then
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
# define our bash completion function
# ------------------------------------------------------------------------------

# $1 -- application
# $2 -- current word
# $3 -- previous word
# $COMP_CWORD -- the index of the current word
# $COMP_WORDS -- array of words
# --commands # --sub-commands # context-specific commands

_have ampnode &&
_ampnode_completion() {
	if [ "$2" ]; then
		COMPREPLY=( $( $1 --list-options | grep "^$2" ) )
	else
		COMPREPLY=( $( $1 --list-options ) )
	fi
	return 0
}

# ------------------------------------------------------------------------------
# set us up the bash completion!
# ------------------------------------------------------------------------------

if [ "$_INTERACTIVE_SHELL" == "true" ]; then

	# first, turn on the extended globbing and programmable completion
	shopt -s extglob progcomp

	# register completers
	complete -o default -F _ampnode_completion ampnode
	complete -o default -F _ampnode_completion ampbuild

	# and finally, register files with specific commands
	complete -f -X '!*.nodule' install-nodule
	complete -f -X '!*.go' 5g 6g 8g
	complete -f -X '!*.5' 5l
	complete -f -X '!*.6' 6l
	complete -f -X '!*.8' 8l

	# '!*.@([Pp][Rr][Gg]|[Cc][Ll][Pp])' harbour gharbour hbpp

fi

# ------------------------------------------------------------------------------
# clean up after ourselves
# ------------------------------------------------------------------------------

unset _OS_NAME _OS_ARCH _OS_ARCH_64 _OS_ARCH_386
unset _BASH_VERSION _BASH_MAJOR_VERSION _BASH_MINOR_VERSION
unset _INTERACTIVE_SHELL _have
