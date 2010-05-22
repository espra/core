We have basic types and metatypes in the plex. Basic types come
predefined, using metatypes you can create you own custom types.

Basic types
===========

- string - contains an encoding. XXX we have not decided about length
  limitations yet.
- int
- complex
- float

- index
- entity
- capability
- storage pointer
- plexname
- datetime
- timedelta
- function
- service
- adaptor
- unit ()
- event
- exception
- sensor

XXX its not decided whether metatypes and properties are types as well

Metatypes
=========

A metatype is an entry in a is_a:metatype index::

  10 is_a:metatype         20 alias_of:10/1
  ----------------         ----------------
  1    ...      ()         1    ...   'foo'

We now have a new metatype 'foo'.

And while looking at the indexes, remember the naming?::

  id property:value
  -----------------
  item
  item
  item

If not, quickly check the document on indexes. Back - ok, where are
possible properties defined? Right, in an index::

  30 is_a:property
  ----------------
  1   ...       ()

Note: the value of the item does not need to be a unit. They can be
strings of functions or whatever. 

That means that a::

  40 liked_by:tav
  ---------------
  item

actually is::
  
  40 30/1:tav
  -----------
  item


and now that we have properties, we can also assign them to
metatypes::

  properties_for:foo
  ------------------
  1   ...       30/1


And while assigning stuff to metatypes, which means I am having
classes a bit in my mind, we should have an index which tells us what
fancy functions we can use with a metatype::

  functions_for:foo
  ------------------
  1 (bar) ....  func

The func here is a function, which is one of our basic types.

XXX There seems to be a problem: does this whole approach with the
is_a:property definition mean that we allow functions as an index
name? And if not - how do we enforce it that it resolves to a unit,
when the value of the item could also be a plexname? It's a bit ad hoc
land here, it seems. 

Type factory
============

Because all that is a bit complicated, you have your beloved type
factory, where you can say "give me a foo", and you just get it. 

