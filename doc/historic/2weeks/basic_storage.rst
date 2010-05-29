Purpose
=======

What Indexes in the entities' indexstore is for metadata, the storage
layer is for binary data. That means: no binary data in indexes. No
movies, no pictures, no email texts ...

The storage layer is for storing 'dumb' data, and sharing to other
people. The last part becomes especially interesting if you have large
stuff of high interest - being slashdotted is not that great.

So, what you do is you store in your indexes the location of the
binary data in the storage layer. And the indexes are not stored in
the storage layer, but in the POD (the plex object database), which
has nothing to do with the storage layer.

Multiple versions are done by having different links to different
versions. So we are not really storing diffs, but the full versions.
Our storage layer has huge capacity.

Storage Item
============

XXX is it really called 'storage item'?

So, what does an Item stored in the storage layer actually consists
of, of what is stored besides the actual binary data?

Primary hash
  Identifies the file and ensures that it is not altered. SHA-2 will
  be the algorithm of choice, but that can be changed

Secondary hash
  Pretty much the same as the primary hash, but tiger tree in the
  start. The idea is that if one hash algorithm is cracked, you have
  another one to fall back to, so that you then can replace the one
  that is compromised.

List of mime types
  The mime types that the binary file has.

Compression algorithm
  Is the data already compressed, and if so, using what algorithm? -
  of course we don't want to compress stuff that' already compressed.

Encoding
  What's the encoding of the binary data?

Encryption
  Is the data encrypted?

TTL
  Time to live - how long is this item valid? If set to eternity, it
  also means that the storage layer should make sure that this item is
  going to be stored persistently, as in real persistent.

Capability
  The capability that should be checked for when accessing the file
  later.

binary data
  The actual binary data


Alternative
-----------

We had some discussion whether it makes any sense to store some
metadata of a file in the storage layer instead of the indexes. It was
a bit undecided. Reason for storing it is that it might be easier to
use your normal editors etc. to access the file if you have direct
access to the storage. Reason against it is that really want to have
one place to look for metadata, and not many, and that it might be
good accept your storage layer, and accept it as a black box instead
of fiddling around under the hood.

In the case of storing the metadata in indexes you would need:

- Primary hash
- Secondary hash
- Capability
- Binary data


Storage Methods
===============

The big and extensive of the storage layer consists of:

put
  Put data

get
  Get some data by hash

has 
  Do we have the file

alternative hash for
  Get the alternative for a hash

stream
  Do not get it, stream it to me

delete
  Try to delete the file

make persistent
  Might also be changing the ttl to eternity. Raises the question
  whether there should be a modify properties method.


Searching
=========

Now, if you want to have a file, you ask for it at three different
locations:

local storage
  Maybe you already have the file, or some in you nodecluster has it.

Entity providing the link 
  If someone pointed you to a file, there's a good chance that they
  have it.

Distributed storage
  If everything fails, you hook up to your big p2p storage. The very
  same method people get their movies these days. No, we are not
  talking about the nature of the movies, at least not in public.


Different Storages
==================

Storages might be implemented using different ways. Filesystem,
databases, memory, http, whatever. There are going to be two default
storages - filesystem and plex p2p

filesystem
----------

Now, how might this only work? Storing binary data in the file system?
Mmmh, big magic needs to work here. Or maybe just a structure of
folders which are a bit optimized for access?::
 
 /a
   /a  #aa...
   /b  #ab...
 /b
   /a  #ba...
   /b  #bb...


Distributed storage
-------------------

For use with the distributed storage we need to hook up with a p2p
system.  For which we have some requirements:

arbitrary hash
  we want to have our own hashing algorithms, and have them
  exchangeable.

stable
  Should work.

performing
  Should scale.

easy to deploy
  Not to much of a hassle to install.

cross platform
  We don't want to be limited to platform foo.

torrenting
  Files we get we can automatically share to others.

dht
  Should support a distributed hash table, instead of having central
  servers telling where to find files

encryption optional
  Encryption should be possible, but not required (XXX why so?)

persistent storage
  Should be possible to differentiate between files that are meant to
  stay around, and files that are ok to delete after using them

By this carefully handcrafted set of requirements basically most if
not all existing p2p systems are ruled out. Praise to tav. Instead we
implement our own system that is not shitty :-)

Plex p2p
--------

One key element of the p2p system is a dht, which will the same
technology as used for the mesh (see entities, or was it plexnames?). 

This dht, which is a black box for now, puts out a list of names (or
maybe just one possible name) of a node you can ask for the file. So
it puts out 'foo'. You then ask foo to give you the file. Foo will
either give you the actual file, or point you to another node, which
you might ask, or it tells you where to look for the file at what
time, basically meaning that it is going to fetch it for you and leave it
at the described place. 
