The goal
========

The idea is to create a p2p platform, on which services can provided
that easily interoperate. 

The idea
========

Every use can run a node on which one more entities live. These nodes
are peers in a peer to peer network. The entities and nodes provide
services to other members of the network. Entities can move between nodes.

The whole communication between entities, services but also within 
services themselves is done by sending of events. You get informed
about events when they are sent directly to you or when a sensor that 
you placed somewhere picked them up. Events are used everywhere -
wherever you have communication you will find events. The different
parts within one program use them. Discoworkers are the key here.

There is a more or less unified way of storing information - indexes.
It is about metadata, not binary data - binary data (files) are used
in a (also p2p) storage layer. Information stored by one application
can easily be reused by another application. Associate RDF here, and
immediately forget it again - the approach writing everything down the
same way is similar, the implementation is completely different. This
association is a secret - if tav learns about it, he might kill me.

Things are named using plexnames. A plexname is a name for something
that can be dynamically resolved - think dns for content naming. The
resolving is done by resolvers, which might be friends or entities you
trust. If something is not known, you ask your friends (or their
resolvers) first.

Therefore trust plays an important role. You can write down which
entities you trust on which topic in a trustmatrix.
You are looking for a doctor? Ask your friends for recommendations -
they will tell you their opinion or what they heard from friends. That
is how reputation is built. 
If you don't want just reputation, but 'hard' rights to allow or
disallow, you will use capabilities, that entity has to do something -
or hasn't. 


The actors
==========

Service
  Services provide functionality.

  The equivalent of a web service or xml-rpc from the world of www.
  Something that does something, or asks other things for getting
  tasks done. Most of the functionality within the plex is found in
  the services provided by the entities. There are some basic services
  found everywhere, but most entities will provide custom funky
  services. The exchange of information can be done using different
  data formats - yaml, xml, plain text, etc.


Plexname
  Dynamically resolved name.

  Whenever you refer to something, you can do so by using a plexname.
  A plexname can directly point to (or be) a definition of something,
  or it can be dynamically resolved to something. E.g.::
   
   jbh/xnet -> t/xnet -> tav/xnet. 

  This dynamic lookup is done using resolvers, which each entity can
  choose and arrange for itself.

Index
  Indexes handle meta data.

  Indexes store all kind of data. A simplified structure of an index
  holding the artist appearing in a movie looks like this::

   artists_for:lotr
   ----------------------------
   Elrond    matrix/Agent Smith # plexname here - dynamically resolved 
   Gollum                    () # This is the actual definition - no
                                  more resolving is done
   ...

  Indexes are used to store just about every information, with one
  notable exception being binary data - the movie itself. That would
  be stored in the storage layer. There are tons, thousands, million
  of indexes - indexes holding the actual information, indexes
  pointing to information in other indexes, indexes of other indexes.
  Indexes are also used as a way to exchange information.

Index Store
  An index store contains many indexes.

  The container for indexes is the index store. This is the place
  where indexes are stored for making them permanent. 

Entity
  An entity contains indexes and provides services.

  This is the place where you would actually find an index store, and
  is also the thing that provides services. There are lots of
  entities. Plex entities can represent real world objects - you and
  I would have an entity, an organisation has an entity, but reason
  for assigning an entity is arbitrary. It's something that knows
  (indexes) and does (services). The central aspect of an entity is a
  cryptographic key (or a pair thereof). That is the id of an entity.

Node
  The place where entities live.

  Entities live on nodes, but also move between nodes. A node is run
  by a computer, which can also run multiple nodes. A node itself
  behaves a bit like an entity - it has an id, has indexes etc. A node
  talks to other nodes. If you want access a service, you actually
  talk to the node asking for a service from an entity. 
  
  A node can also be a cluster of nodes - a nodecluster or
  peercluster.   

Event
  The form of communication within the plex.

  Whenever something changes or you want something to change or to be
  done an event is sent. An event has an id, a type, a ttl and a
  payload. An event can get passed on. An event does not excpect an
  answer - but it might trigger one. Events get transported using
  different means of transportation - may it be pigeons, http or smtp.
  
  Events also can also hold an address where exceptions that are
  caused by this event are sent to (thrown to).

  Events can be used to allow transactions.

Sensor and hooks
  Sensors are placed on hooks where they listen for events that they
  pass on to services.

  You are a service that wants to be informed about changes, e.g. when
  an index gets a new entry? You will place a sensor on the on_change
  hook at the that index. Now whenever a change event is created, your
  sensor will cause the event to be transported by the transport
  mechanism to you, the sensor. 

Discoworkers
  Discoworkers allow DIStributed COncurrent programming, partly by
  using events and sensors.

  Discoworkers are a way to program. They allow the programmer to code
  functions within a python script that run more or less parallel.
  They communicate using the event system described above. This also
  allows that you can do outsourcing at no extra costs - you can move
  a resource hungry function to another node, and the events between
  the functions just have to travel a bit further. 


Trustmatrix
  The place where trust relations are stored.

  If entity foo trusts entity bar on domain spam, this information is
  stored in the trust matrix (which might be implemented using
  indexes). This information can then be used by a trust metric to
  compute the relative value of other entities (a friend of a friend
  is a friend). The information stored in the trust matrix as well as
  the output from the trust metric by a reputation metric - to find
  out the reputation of an object. What doctor does your peer group
  recommed to you?

Capability
  The stored right to do something.

  A capability is used by services, nodes or entity to determine if
  someone is allowed to access a resource. It contains the permission
  - a unique identifier of a right -, constraints and a cryptographic
  signature of the issuer. It can be kept on the issuing server, than
  the receiver of the capability just gets a reference to it, or can
  be actually given out. E.g. if an entity wants to use a service, the
  service will check either a stored capability or the capability the
  entity hands over.


UIC
  The ueberinterface controller

  Interface logic is decoupled from application logic - pretty much
  using the Model-View-Controller idea. The goal is to allow different
  interfaces for the same application, and also a generic interface to
  many applications - e.g. a command line. 

  The UIC uses the same ideas as the discoworkers - you use events to
  pass on requests, changes and information. You usually have the UIC
  on the node side, which handles flow of events, and a helper running
  on the client being responsible for arrangement of display
  elements and handling of user input. 


