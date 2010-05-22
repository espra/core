A service is the point where the action logic resides. A bit like a
program in the plex programing language. It is something you can
'call', and that gives you some output. Pretty much like a function or
a method or a cgi script.

(call here means actually something you can send an event to, and that
- most likely - returns an event back to you with something close to a result).

Not a function?
===============

Confusion may arise about what the difference between a function and a
service is. It is a bit a grey zone. 

But usually a service:

- ties together multiple functions
- is 'larger' than a function
- has a defined interface, a documented api
- is quite stable
- is registered as a service with a node

My comparison would be a bit the difference between private and public
methods in python classes. You can call them both, and there is no
real difference. But one is supposed to be stable, documented and
being used from the outside.

Essential Services
==================

On a node there will be a number of basic services available:

Dealing with [indexes, events, sensors]
  Creation, modification, deletion of stuff. 

Language daemon
  Execute code in a large variety of languages. Yes. Execute as in
  execute. Are we insane? No, at least not completely. Because we have
  S.A.F.E, our security audited function environment, that the
  language daemon is - it will only run code that is signed by people
  you trust.

Type factory
  Creates types for you, and creates the necessary underlying index
  structure for you.

Plexname resolvers
  Resolve a plexname :-)

Type adaptor
  Give it a foo, and it returns something which behaves like bar with
  the data of foo.

Function store
  This is where you find the functions. Basically a mapping of names
  to code. The code actually resides in the storage layer.

Harvester(s)
  Things picking up events or creating events from input. Like an
  indexer for example, who takes an document, splits it up and creates
  content in the indexes by creating the appropriate events.

Views
  Takes input, and gives you something pack which is useful for
  displaying the data. Like an html snippet. More or less a special
  form of adaptors - makes data behave like a display of data.

Brokers
  You need to use a resource on another node, or migrate some of your
  services there? Ask a broker, it will do the financial (sic) stuff
  for you and find you a good deal. 

Service migration
  Your machine gets visited a bit to well? Running low on bandwith?
  Migrate your service somewhere else - using the service migration.
  It will take the necessary steps.

Reputation service
  Tells you what the reputation of something is. The key point here is
  that this service actually is working the reputation metric
  algorithms. Indexes may contain reputation data about foo, but the reputation
  service will tell you if you can trust foo. Is used by broker and
  service migration



  