Plexnames
=========

A plexname is for the Plex what a URI is for the Web. They are used in
indexes, events and plexlinks (see below). 

Constructs
----------

There are some basic forms:


plex:movies/alexander
  This is the name for the alexander resource in the movies topic. If
  the resource can not be found, use the resolving mechanism.

plex:~some_entity/movies/alexander
  The same resource, but now as part of some_entity. If the alexander
  resource now directly originates from the entity or not is not defined.

plex:~/movies/alexander
  The same resource, but now within the same entity where that name is
  used. No resolving takes place.

If an entity is asked for such an resource it can either return the
resource itself, e.g the picture, or a pointer to somewhere else, e.g.
the storage layer.

plex:storage/A23SF4333
  This points directly to a file in the storage layer. 

Different Resources
-------------------

It would be possible to have a movie alexander, a picture alexander, a
soundtrack alexander, a textbook alexander and so forth. The question
would be what resource to return, if asked for alexander? The client
has to describe what type it prefers. I would compare it to the
'Accept' header in http.
XXX Can I directly ask for alexander.jpg?


Resolving
---------

If a resource is not found it is searched for using the resolver
hierarchy. An entity defines a hierarchy of resolver which it is going
to use for certain topics - for music ask foo, then bar, for science
better ask bar first, then foo. So it's a bit what dns is for the net,
but this time for resources, not machines. 


Uniqueness
----------

XXX Up to now it's undecided if the same name can be used for more
than one resource, or if it has to be unique. And if so, what exactly
the scope of uniqueness is - within the entity, within a topic, or
something else? tav has to think, I would recommend to have no
requirement for uniqueness - which would mean that an entity always
returns an ordered list of resources when asked for something.

Plexlinks
=========
Plexlinks can for example be used in Documents - comparable to hrefs
in html or WikiLinks in wikpages.

The standard format for writing links to other resources is
[plex:movies/alexander]. Square brackets. 

Functions
---------

Resources can also be services - maybe comparable to an
URI that points to a xml-rpc resource. The question now arises what is
meant when pointing to a resource - the resource itself or the output
of it. The problem is a bit comparable with URLs in HTML - do you want
to link to it, or include it - link to the picture, or display it in a page.

The solution with plexlinks is that you can either point to a resource 
`[plex:joerg/add]` or that you call the function `[#plex:joerg/add]`.

Now, to make it a bit more fancy, you can also combine resources /
functions::

 [#convert from:usd to:gpb ~amazon/alexander.price | #print-localized-currency]

This will call the local convert resource, passes from and to as
parameters and tells it to work on the price of alexander at amazon
(USD 8 -> GBP 6). The price is then passed to the local
print-localized-currency (6,00 GBP in Germany, 6.00 GBP on that
island).

Notice the '#', which triggers the calling, and see the the fancy pipe
'|', which behaves very much like a pipe in a shell. 
