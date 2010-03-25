# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

require 'rubylibs'
require 'sinatra/async'
require 'sass'

$a = 1

class ZeroCSS < Sinatra::Base
  register Sinatra::Async

  disable :show_exceptions

  aget '/' do
    content_type 'text/css', :charset => 'utf-8'
    foo = params[:id]
    body {
      haml '%h2= foo', :locals => { :foo => $a }
    }

    # "CSS! #{params}"
  end

  aget '/bar' do
    $a += 1
    body "incr"
  end

  aget '/delay/:n' do |n|
    EM.add_timer(n.to_i) { body { "delayed for #{n} seconds" } }
  end

end
