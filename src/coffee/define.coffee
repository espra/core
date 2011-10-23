# Public Domain (-) 2011 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

# The `define` function provides a utility wrapper to namespace code blocks. If
# an explicit target isn't passed in, it defaults to using `window` or `exports`
# as the top-level object.
#
# This functionality has been adapted from the `namespace` function defined in
# the [CoffeeScript FAQ](https://github.com/jashkenas/coffee-script/wiki/FAQ).
define = (root, name, constructor) ->
  [root, name, constructor] = [(if typeof exports isnt 'undefined' then exports else window), arguments...] if arguments.length < 3
  target = root
  target = target[item] or= {} for item in name.split '.'
  constructor target, root
  return