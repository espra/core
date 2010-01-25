#! /usr/bin/env python

def configure(conf):
    conf.check_tool('python')
    conf.check_python_version((2,4))
