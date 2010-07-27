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
	echo "You might want to add the above line to your .bashrc or equivalent."
	echo
	exit
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
# extend the PATH
# ------------------------------------------------------------------------------

export PATH=$AMPIFY_ROOT/environ:$AMPIFY_ROOT/src/codereview:$PATH

# ------------------------------------------------------------------------------
# try to figure out if we are inside an interactive shell or not
# ------------------------------------------------------------------------------

test "$PS1" && _INTERACTIVE_SHELL=true;

# ------------------------------------------------------------------------------
# the auto-completer for optcomplete used by the amp runner
# ------------------------------------------------------------------------------

_amp_completion() {
	COMPREPLY=( $( \
	COMP_LINE=$COMP_LINE  COMP_POINT=$COMP_POINT \
	COMP_WORDS="${COMP_WORDS[*]}"  COMP_CWORD=$COMP_CWORD \
	OPTPARSE_AUTO_COMPLETE=1 $1 ) )
}

# ------------------------------------------------------------------------------
# set us up the bash completion!
# ------------------------------------------------------------------------------

if [ "x$_INTERACTIVE_SHELL" == "xtrue" ]; then

	# first, turn on the extended globbing and programmable completion
	shopt -s extglob progcomp

	# register completers
	complete -o default -F _amp_completion amp
	complete -o default -F _amp_completion optcomplete-commands

	# and finally, register files with specific commands
	complete -f -X '!*.go' 5g 6g 8g
	complete -f -X '!*.5' 5l
	complete -f -X '!*.6' 6l
	complete -f -X '!*.8' 8l

fi
