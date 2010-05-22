Entities
========

Entities are used to represent things from reality(tm) in the Plex.
Most things you use a noun for can be represented using an entity.
This especially includes people. So you and I are represented by
entities. Entities live on nodes (or a cluster of nodes).

Central part of an entity is a pair of keys, the public and the
secret key. The secret key can be stored at different places,
depending of usage pattern. I for example would like to store it on a
usb drive. The public key becomes more or less the entity, and is
going to be pretty accessible. The keys are created for you when you
ask a node to create an entity for you.

Entity handling
---------------

With a node (or a node cluster) you can manage entities. Besides the
usual 'dealing with' stuff (see below - Using an entity) you can also
ask a node to make an entity active. That means that if the entity is
not known to the node, it will copy all necessary (or selected) data
over from the old node. The 'old' node gets marked as being no longer
the active node. Then the new active entity will sync it's
stuff with the 'old' node.

Another feature is that along with the creation of the entity all
necessary ingredients for deleting the entity are generated. As the
entity is to a large extend a key, it means actually revoking the
entity, of which the node than automatically takes care as far as possible.

Attributes
----------

name
  Each entity has exactly one name, but that can be changed over time.
  This can be considered as the nick the entity is giving itself.

indexstore
  An entity has lots of indexes, which are all located in the
  indexstore.

plexresolvers
  A hierarchy of plexresolvers. 

eventstore
  The place where all the events are stored.


(Shouldn't these all be stored in the indexstore?)

trustmatrix
  The trustmatrix, the view on the world so to speak is stored here

capabilitystore
  The rights management stores its information here.

misc
  Lots of other stuff can be stored in the entity


Using an entity
---------------

There are many things you can do with an entity, and the PAL is
defining what exactly. But as a general rule you ask a node to give
you a connection to an entity, and using the connection you then can
do your stuff. Think HTTP. 

What can be done?

Shortcut: whenever I write something about 'dealing with' it means the 
usual create, modify, query and delete operations.

dealing with indexes
  works on the data in the indexstore

dealing with events
  -> eventstore

dealing with capabilities
  -> capabilitystore

dealing with data
  This is partly an interface to the storage layer. E.g. you ask an
  entity about a 'picture1' (which happens to be stored at
  storage/file123). Depending on the question and the entity
  it might return the actual contents of storage/file123 or give you
  the reference to storage/file123. And of course capabilities will
  come into play. 

typefactory, type adaptor
  Call the type or adaptor factory to get - tada - a type or an
  adaptor dealing with types.

use function
  entities can provide special functions (think the world famous add
  function to add two numbers.) Different entities, different set of
  functions. XXX Hey, this just asks for xmlrpc


Subentities
-----------

An entity can have subentities - think Jekyll and Hide. A subentity
has it's own key, capabilitystore and plexresolvers, but shares the other
stuff with the main entity. There is a proof of relationship between
main entity and subentity, but that's only accessible to the main entity.

Identities
----------

An Identity is a combination of entity and the id of a cluster of
nodes (which can be of size 1): entity@cluster. The id of the cluster
is again a key. An entity can join more than one cluster. 


Ad-Hoc 
======
An idea was to have 5-fold tuples as the protocol within PAL for speaking to
entities: (node,entity,activity,[{paramters}],hash) The hash is, ah, a
hash, which can be used for doing auth etc. No idea what it really is
for, and why its not just another parameter. 

