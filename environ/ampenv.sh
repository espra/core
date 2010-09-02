#! /bin/bash

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
	echo "You might want to add the above line to your .bashrc or equivalent."
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

export AMPIFY_ENVIRON_DIRECTORY=`pwd -P 2> /dev/null` || return $?

cd $OLDPWD || return $?

export AMPIFY_ROOT=$(dirname $AMPIFY_ENVIRON_DIRECTORY)

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

if [ "x$PRE_AMPENV_PATH" != "x" ]; then
	export PATH=$AMPIFY_ROOT/environ:$AMPIFY_LOCAL/bin:$PRE_AMPENV_PATH
else
	if [ "x$PATH" != "x" ]; then
		export PRE_AMPENV_PATH=$PATH
		export PATH=$AMPIFY_ROOT/environ:$AMPIFY_LOCAL/bin:$PATH
	else
		export PATH=$AMPIFY_ROOT/environ:$AMPIFY_LOCAL/bin
	fi
fi

case $_OS_NAME in
	darwin)
		if [ "x$PRE_AMPENV_DYLD_LIBRARY_PATH" != "x" ]; then
			export DYLD_FALLBACK_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$AMPIFY_LOCAL/freeswitch/lib:$PRE_AMPENV_DYLD_LIBRARY_PATH:/usr/local/lib:/usr/lib
		else
			export PRE_AMPENV_DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH
			if [ "x$DYLD_LIBRARY_PATH" != "x" ]; then
				export DYLD_FALLBACK_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$AMPIFY_LOCAL/freeswitch/lib:$DYLD_LIBRARY_PATH:/usr/local/lib:/usr/lib
			else
				export DYLD_FALLBACK_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$AMPIFY_LOCAL/freeswitch/lib:/usr/local/lib:/usr/lib
			fi
		fi
		export DYLD_LIBRARY_PATH=/this/path/should/not/exist;;
	linux)
		if [ "x$PRE_AMPENV_LD_LIBRARY_PATH" != "x" ]; then
			export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$PRE_AMPENV_LD_LIBRARY_PATH
		else
			if [ "x$LD_LIBRARY_PATH" != "x" ]; then
				export PRE_AMPENV_LD_LIBRARY_PATH=$LD_LIBRARY_PATH
				export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$LD_LIBRARY_PATH
			else
				export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib
			fi
		fi;;
	freebsd)
		if [ "x$PRE_AMPENV_LD_LIBRARY_PATH" != "x" ]; then
			export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$PRE_AMPENV_LD_LIBRARY_PATH
		else
			if [ "x$LD_LIBRARY_PATH" != "x" ]; then
				export PRE_AMPENV_LD_LIBRARY_PATH=$LD_LIBRARY_PATH
				export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib:$LD_LIBRARY_PATH
			else
				export LD_LIBRARY_PATH=$AMPIFY_LOCAL/lib
			fi
		fi;;
	*) echo "ERROR: Unknown system operating system: ${_OS_NAME}"
esac
	
if [ "x$PRE_AMPENV_PYTHONPATH" != "x" ]; then
	export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/third_party/pylibs:$PRE_AMPENV_PYTHONPATH
else
	if [ "x$PYTHONPATH" != "x" ]; then
		export PRE_AMPENV_PYTHONPATH=$PYTHONPATH
		export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/third_party/pylibs:$PYTHONPATH
	else
		export PYTHONPATH=$AMPIFY_ROOT/src:$AMPIFY_ROOT/third_party/pylibs
	fi
fi

_THIRD_PARTY=$AMPIFY_ROOT/third_party
_ENV_VAL=$_THIRD_PARTY/vows/lib:$_THIRD_PARTY/jslibs:$_THIRD_PARTY/coffee-script/lib

if [ "x$PRE_AMPENV_NODE_PATH" != "x" ]; then
	export NODE_PATH=$_ENV_VAL:$PRE_AMPENV_PATH
else
	if [ "x$NODE_PATH" != "x" ]; then
		export PRE_AMPENV_NODE_PATH=$NODE_PATH
		export NODE_PATH=$_ENV_VAL:$NODE_PATH
	else
		export NODE_PATH=$_ENV_VAL
	fi
fi

if [ "x$PRE_AMPENV_MANPATH" != "x" ]; then
	export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/share/man:$PRE_AMPENV_MANPATH
else
	if [ "x$MANPATH" != "x" ]; then
		export PRE_AMPENV_MANPATH=$MANPATH
		export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/share/man:$MANPATH
	else
		export MANPATH=$AMPIFY_ROOT/doc/man:$AMPIFY_LOCAL/share/man:
	fi
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
	complete -o default -F _amp_completion ampnode
	complete -o default -F _amp_completion git-review
	complete -o default -F _amp_completion git-slave
	complete -o default -F _amp_completion hubproxy

	# and finally, register files with specific commands
	complete -f -X '!*.go' 5g 6g 8g
	complete -f -X '!*.5' 5l
	complete -f -X '!*.6' 6l
	complete -f -X '!*.8' 8l

fi

# ------------------------------------------------------------------------------
# clean up after ourselves
# ------------------------------------------------------------------------------

unset _OS_NAME _OS_ARCH _OS_ARCH_64 _OS_ARCH_386
unset _BASH_VERSION _BASH_MAJOR_VERSION _BASH_MINOR_VERSION
unset _have _INTERACTIVE_SHELL
unset _THIRD_PARTY _ENV_VAL