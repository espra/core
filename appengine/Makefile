# Released into the Public Domain by tav <tav@espians.com>

#
# A simple Makefile for compiling various static files.
#

# ------------------------------------------------------------------------------
# some konstants
# ------------------------------------------------------------------------------

latest := .latest

# ------------------------------------------------------------------------------
# we declare our phonies so they stop telling us that targets are up-to-date
# ------------------------------------------------------------------------------

.PHONY: update

# ------------------------------------------------------------------------------
# our rules, starting with the default
# ------------------------------------------------------------------------------

latest: $(latest)

$(latest): update
	@touch $(latest)
