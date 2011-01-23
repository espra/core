# Public Domain (-) 2010-2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

require 'pathname'

MAIN_ROOT = Pathname.new(__FILE__).realpath.dirname.dirname
RUBYLIBS_ROOT = MAIN_ROOT.join('third_party', 'rubylibs')

Dir.glob(RUBYLIBS_ROOT.join('*', 'lib')).each do |lib_path|
    if /.*jspec.*/.match(lib_path)
        $:.unshift File.join(File.dirname(lib_path))
    else
        $:.unshift lib_path
    end
end
