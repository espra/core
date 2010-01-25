#! /usr/bin/env ruby

# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

require 'pathname'

MAIN_ROOT = Pathname.new(__FILE__).realpath.dirname.dirname.dirname
RUBYLIBS_ROOT = MAIN_ROOT.join('third_party', 'rubylibs')

Dir.glob(RUBYLIBS_ROOT.join('*', 'lib')).each do |lib_path|
    if /.*jspec.*/.match(lib_path)
        $:.unshift File.join(File.dirname(lib_path))
    else
        $:.unshift lib_path
    end
end
