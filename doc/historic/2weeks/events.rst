Events are about change. They are not about a state something is in,
or should be in, but about what was done or should be done. 

Events are sent over the network, but also used within a singe
application. They are one basic form of communication on the plex
level (which is not the application level).

Attributes
==========

What does an event consist off?

Id
  This is a uuid, which consists of time, nodeid and other stuff. The
  id is universal unique, but constructed in a way that it also acts
  as a logical clock - if you have to ids, you can always tell which
  one came earlier.

Name (aka type, XXX has to be decided)
  A string describing the nature of the event. 

TTL
  Time to live. Is a tuple consisting of:
  
  maximum hops
    The maximum number of hops an event should take (yes, they can be
    passed on)

  counter
    How many hops has it actually passed?

  maximum time
    The time the event is actually valid for. 
Message
  The actual payload. Can also be a pointer to some entry in some
  index. Can have other events attached.

 

Hooks and Sensors
=================

Besides 'manually' creating an event that for example asks for the
creation of an index, there is also a mechanism for automatic event
creation and propagation. 

Hooks
-----

On your indexes you have hooks, to which you can add sensors:

- on_create
- on_query
- on_modify
- on_delete


Sensors
-------

To those hooks you can add a sensor. A sensor could be called
listener, ear, pigeon, etc. It consists of:

entity
  Entity to send the event to

function
  Which function to exactly send to, and also how to reach it - using
  http, smtp, pigeon or whatever. 

string
  A string that should be passed on to the function. Just a place to
  store information to pass on.

capability
  The capability needed for actually talking to the entities'
  function. Otherwise the event might not be accepted by the function

The information in the sensor is then used to pass on the event to the
defined recipient's function using the defined method. 

Note here, that the creation of the event is decoupled from the
transportation of the event. You basically put the event in a prepared
envelop, and sent if of. Someone else is taking care to deliver it.
(It might now be a good time to read discoworkers again - it's exactly
the same principle).

Filtering Sensors
-----------------

By default, when you are anonymous or don't have enough rights you can
only attach a simple sensor, which more or less accepts all events. If
you want to run a filter, to select only specific events, you actually
need to have better rights - if the node executes the filtering code,
is uses resources, and that should not be open to everyone. 


Harvesters
----------

The function that is going to be informed of an event is actually
called 'harvester'. So, from the sensors point of view the recipient
is just a another function, but from the inside its one of those
thingies that one uses to collect events. The ones mentioned for the
sensors are pretty passive ones - they listen for incoming events. But
there are also much more active ones. More like an agent, searching
for stuff, and creating events whenever it encounters something of
interest. 

You could imagine for example an Harvester that watches an rss feed,
passes on the mentioned documents to an indexer, which then creates
events that trigger creation of entries in an index.

See also the services. 


Transactions
============

With the whole event routing mechanism we can nice transactions going.
Which are going to be a bit costly. So don't use them for just our
basic events, which are going to be atomic anyhow. 

The transactions consist of three phases:

send
  You send out a message asking for something to be changed. 

commit
  Applying the required change, but still keeping in a non-valid
  state, so that it can be made undone.

validate
  The change is now becoming real.


The basic idea is that you have an event, that gets passed on to all
the participating players of the transactions, and gets returned (sic)
to the creator at the end of each round. 

Exceptions
----------

The event carries information on where to send exceptions that are
being raised. Typically that would be the creator. By that the creator
knows what went wrong on which state of the transaction.

There is a special exception to think of - the `TemporalParadox`. This
is a player involved in the transaction has to do something while
making the required actions which can't be undone - like undrinking a
cup of tea. As we can't travel back in time (yet), we have a problem.


But now it's time for an

example
-------

Lets say tav wants to transfer 20 EUR from Joerg's account to t's
account. He has all required permissions. 

1. Tav creates an event::
   
    uuid: 123
    name: 'Transfer'
    message: Transfer 20 EUR from bankaccount/joerg to bankaccount/t.
    Deliver exceptions to tav/bankservices

2. Tav puts a sensor to bankaccount/t's on_modify hook.

3. 123 gets copied in the transaction manager, and delivered to
   bankaccount/joerg (wherever that might be, plexname lookup will
   help here). 

4. At bankaccount/joerg an entry in the index of tranfers is created,
   which contains the requested change (deduct 20 EUR) and the event
   (123) request it.

5. The event gets passed on to bankaccount/t, where the equivalent
   happens.

6. The on_change hook gets activated, the sensors senses, and the
   event gets transported back to tav's transaction manager.

7. tav's transaction manager sees that he has got the event returned
   exactly as sent, and knows that he can now send out the commit.

8. The commit, using the same uuid (123)  but having a 'commit'
   payload is sent out, following the same stations. But this time the
   change is actually applied.

9. At the end of that round the commit reaches tav again, who then
   sends out a validate. Everyone is happy. 

If something goes wrong, tav receives the exception, and can then send
out an 'drop transaction 123'. If a timeout is reached (e.g a player recieved the
commit, but not the validate), the player also sends an exception to
tav - who has then to take appropriate measures.

Notes
=====

- The transactions are non-blocking. Even if transaction 1 is still
  waiting for its validate, transaction 2 can take place.
