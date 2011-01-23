# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

"""Utility functions for dealing with the operating environment."""

import sys
import subprocess

from os import getcwd

def read_file(filename, mode='rU', content=None):
    """Read and return the contents of the given filename."""

    f = open(filename, mode)
    try:
        content = f.read()
    finally:
        f.close()
    return content

def exit(message, error_code=1):
    """Write an error message to stderr and exit with the given error_code."""
    sys.stderr.write(message + '\n')
    sys.exit(error_code)

class CommandNotFound(Exception):
    """Exception raised when a command line app could not be found."""

def run_command(
    args, retcode=False, reterror=False, exit_on_error=False, error_message="",
    log=None, redirect_stdout=True, redirect_stderr=True, cwd=None,
    shell=sys.platform.startswith('win'), env=None, universal_newlines=True
    ):
    """Execute the command with the given options."""

    log_message = "%s cwd=%s" % (' '.join(args), cwd or getcwd())
    if log:
        if hasattr(log, '__call__'):
            log(log_message)
        else:
            sys.stderr.write("Running command: " + log_message + '\n')

    if redirect_stdout:
        stdout = subprocess.PIPE
    else:
        stdout = None

    if redirect_stderr:
        stderr = subprocess.PIPE
    else:
        stderr = None

    try:
        process = subprocess.Popen(
            args, stdout=stdout, stderr=stderr, shell=shell, cwd=cwd, env=env,
            universal_newlines=universal_newlines
            )
        out, err = process.communicate()
    except OSError:
        error = sys.exc_info()[1]
        if error.errno == 2:
            if exit_on_error:
                exit("Couldn't find the %r command!" % args[0])
            raise CommandNotFound(args[0])
        if exit_on_error:
            exit("Error running: %s\n\n%s" % (log_message, error_message))
        raise

    if process.returncode and exit_on_error:
        if stderr:
            exit_extra = error_message or err
        else:
            exit_extra = error_message or out
        if exit_extra:
            exit("Error running: %s\n\n%s" % (log_message, exit_extra))
        else:
            exit("Error running: %s" % log_message)

    if retcode:
        if reterror:
            return out, err, process.returncode
        return out, process.returncode

    if reterror:
        return out, err
    return out
