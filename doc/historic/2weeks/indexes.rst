Basic Layout
------------

Indexes are used to store all sorts of information, basically
everything besides binary data. Every entity has its own set of
indexes.

The basic structure looks like that::

 indexid property:value
 ----------------------
 Item
 Item
 ...

Indexid
  The id of the index. Has to be unit within the entity. In the
  examples they will just plain ints, 10, 20, 30.

Item
  The actual entry in the index. RDF would consider this to the
  subject. Consists of:
   
  id
    An unique id within the index. Can be a number of an int. 

  date
    The date this item was created.

  type
    The type of the value. String, int, whatever.

  value
    The actual value of the entry. Of special importance is the 'unit'
    type '()', see below.

property
  What RDF would call the predicate. 'is_a','status_of','is_shining'
  would be all valid examples.

value
  What RDF would call the object. Can be just about everything that
  one can refer to in the Plex.  Can be a reference or a basic type.
  'joerg', '10/4' as a reference to entry number 4 in index number 10. 

The unit '()'
-------------

The unit is a special value in an item. It indicates that the item in
which it is the value is the final destination of the resolving
process. No more searching. Truth is reached. This is it. So, the unit
defines a point of reference, something to talk about. With it you can
have somewhat the same functionality you have in RDF with anonymous
nodes - something you can talk about.


An example
----------

Let's say I would like to store a remarkable observation in my
indexes: 'the sun is shining' (if you know Berlin in winter, you know
what I mean).::


  10 is_a:metatype            # We define an index of metatypes.
  ----------------   
  1   ....   ()               # Item 1 is a unit, something to talk about.
  

  20 alias_for:10/1           # Let's name item 1 of index 10 (think alias). 
  -----------------          
  1   ....  'heavenly body'   # Yeah, 10/1 that is what 10/1 should be
                              # known as


  30 is_a:10/1                # Index of 'things' that are of metatype 10/1
  ------------
  1   ...   ()                # This item is again a unit. 



  40 alias_for:30/1           # We need to name 30/1 of course
  -----------------
  1   ...  'sun'              # Now we know what 30/1 is called



  50 is:'shining'             # Stuff that is shining
  ---------------
  1    ...   30/1             # 30/1 is rather bright thing.


If this whole stuff would have been in our beloved programming
language, it looked like this::

  class Foo:                  # equals 10
     pass
  
  HeavenlyBody = Foo          # equals 20
  
  sun = Foo()                 # 30 and 40 combined

  sun.status='shinning'       # 50
  
  XXX or should that be 

  shining_things.append(sun)

  
Storage of Indexes
------------------

Indexes are stored in the Plex Object Database, which is roughly
something like the ZODB for Zope, but better suited for heavy writing. 

Distributed Indexes
-------------------

Indexes can be distributed among more that one machine. A bit like a
btree, that splits up, and has its different parts being stored on
different machines. 

Versioning
----------

It is possible to choose from three ways of versioning:

- full - copies of the index are kept around
- change - the diffs are kept
- none - don't version the index

Index Query
-----------

Now this is the thing I am really looking forward to. How the whole
magic works, the speedy, shiny information retrieval that will
enlighten our lives....