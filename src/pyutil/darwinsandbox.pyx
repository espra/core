# Public Domain (-) 2010-2011 The Ampify Authors.
# See the UNLICENSE file for details.

import sys

cdef extern from "sandbox.h":
    ctypedef unsigned long uint64_t
    int sandbox_init(char *profile, uint64_t flags, char **errorbuf)
    int SANDBOX_NAMED
    char kSBXProfileNoWrite[]
    char kSBXProfileNoInternet[]
    char kSBXProfileNoNetwork[]
    char kSBXProfileNoWriteExceptTemporary[]
    char kSBXProfilePureComputation[]

PROFILES = {
    'NoWrite': kSBXProfileNoWrite,
    'NoInternet': kSBXProfileNoInternet,
    'NoNetwork': kSBXProfileNoNetwork,
    'NoWriteExceptTemporary': kSBXProfileNoWriteExceptTemporary,
    'PureComputation': kSBXProfilePureComputation
    }

def sandbox(profile_name):
    profile = PROFILES[profile_name]
    cdef char **errbuf
    retval = sandbox_init(profile, SANDBOX_NAMED, errbuf)
    if retval != 0:
        print "ERROR: Sandboxing failed."
        sys.exit(1)
