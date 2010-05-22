The networking covered in this session is about how nodes communicate
with each other, in contrast to applications communicating.

Transport
=========

This is how over protocol stack looks like::

   cake  
  airhook
    udp
    ip

So, the obvious thing is that we are not using the normal http over
tcp. So, what are the parts doing?

Lets look at their webpages.

cake
  http://www.cakem.net
  
  "CAKE is a networking protocol in which all messages are addressed to
  a public key, and are signed by the source public key. Public key
  identifiers are treated like IP addresses. They represent the
  destination or source of any particular message."

airhook
  http://airhook.org/
  
  "Reliable, efficient transmission control for networks that suck."
  
  "Airhook is a reliable data delivery protocol, like TCP. Unlike TCP, 
  Airhook gracefully handles intermittent, unreliable, or delayed
  networks. Other features include session recovery, queue control, 
  and delivery status notification."

udp + ip
  They do the normal stuff - sending packets to the receiver

The reasoning::
  
  - we need high responsiveness - we can't afford to have long
    timeouts or congestion problems that tcp brings along. -> airhook.

  - we need link2link encryption -> cake

  - we are more interested in speaking to a "key" - the entity - than
    to a certain ip-address.

  - we would like to work with/around NAT and Firewalls. UDP makes
    life a bit simpler

Data exchange
=============

Once we can somehow connect to the other node, some negotiation takes
place before we start the 'real talking'.

  - send an initialisation vector (contains a key)
  - add the key to the key message digest (hmac)
  - negotiate the mode of transfer (xml, yaml, binary, plaintext, ...)

The idea here is to pretty easily extendable. Even though we know of
course the best selection of protocols, we only know that now :-).
So we need to be able to adopt to future developments. (See auto update).

XXX What exactly is happening?


Mesh
====

The mesh is something equivalent to our normal phonebook. It describes
at what number you can reach which entity. An entry in the mesh
consists of several parts:

 - method - http, ftp, pigeons ...
 - location - a position that matches the method - ip for http, gps
   for pigeon.
 - ttl - The timespan in which this entry is valid
 - signature - a crypto signature which contains the public key.

The mesh is a distributed hash table (dht)- which means that on each node
(one ore more) buckets live, which contain a part of the mesh - I have
A-B, you have D-E. The dht is more or less transparent to the nodes.
The idea here is to use something like kademlia
(http://en.wikipedia.org/wiki/Kademlia). XXX is that true?

Also note that the mesh does not contain information on what services
an entity provides. 

XXX is a mesh about nodes or entities?

Routing table
=============

Now, in order to speak to a node you look up the right entry in the
mesh. When you then start talking to the node you will get extra
information during the negotation process - and that data you store in
your local "address book". It's a bit like a local cache of mesh
information thats relevant to you. 

You would store in that routing table which nodes have high
responsiveness, where to find them, how to best communicate with them
etc.

Seed
====

If you start up your node you are basically alone. To help you to
connect to the plex you will get a list of some nodes you can reach
along with your distribution. Their might be even some seed servers
which point you to highly alive nodes. There you node will be able to
connect to the mesh.

Autoupdate
==========

In order to spread new version of the software, a form of auto-update
is built in from the beginning. It will receive signed code packages.

The question now is what happens if the mpaa or another mafia
organisation tries to shut us down, using an army of lawyers? They
would usk us to "update" the network to a close. 

So solve this, esp is not going to be the only entity that has to sign
the update. Instead there will be a list of keys of which, let's say
70% have to sign the update. Of course, these people should only sign
when the open source code of the update seems to be ok. 

That means that the army of lawyers now need to have a court order for
70% of the people. Which might be quite a hassle if all keys (and
holders) are spread over the globe.

I think that the method is not 100% error proof, but maybe as good as
you can get at the moment. Suggestions?

