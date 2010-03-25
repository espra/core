# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

require 'zerocss'

# Sass::Plugin.options[:style] = :compact

# Sinatra::Application.default_options.merge!(
#   :run => false,
#   :env => :production
# )

run ZeroCSS.new