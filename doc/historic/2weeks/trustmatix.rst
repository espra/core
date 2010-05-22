The trust matrix is used to store trust information. Entities have one
as well as nodes. 

XXX The storage of the matrix is undecided

XXX Lots of stuff in here is undecided. But let's try to give an idea
of what it is about. 

The idea
========

For many uses you need to know how well you can trust other entities -
do you want allow them to connect to you, can you transfer your
services there, etc. Also, for getting rating about things you might
not only want to use you own knowledge, but also the knowledge of
other people you trust. 

To store the information about trust, you use a trust matrix. To do
calculations you use a trust metric of your choice.

Basic layout
============

A fancy ascii art of a matrix looks like this::

  
       |    _ the value
       |   /
 Entity|  x
       |
       |___________
          D
          o
          m
          a
          i
          n

(There might be a reason why people usually don't let me do the graphics)

Entity
  The Entity we would like to say something about. XXX could be also an
  arbitrary plexname.

Domain
  The domain in which the trust is expressed - I trust tav in regard
  to nutella taste for example. XXX what exactly is a domain?

The 'x' - the value
  That's the actual value of the trust expression. What it actually
  says is undecided. XXX decide this. There is an understanding that
  the value should have (at least) two properties:

  value itself XXX we need a better word here
    This could either be a percentage, e.g 0.8, or a relative
    expression as in "more than foo". This is quite undecided as well.
    Percentages and relative expression both have their pros and cons

    It is also not decided if you can express distrust here.

  horizon
    This means how far the trust reaches - do you just the entity, or
    also other entities that this entities trust? And if so, over how
    many hops? The friends of friends of friends of a friend?.

    You are also able to express that you don't trust the entity in
    the given domain, but other entities that the entity trusts - Bob
    might have no clue about ping elephant babies, but surely his
    recommendations on this topic are great.


Metric
======

This whole trust matrix comes to use if one uses it to get an
impression of something. You check what you know about this foo, and
you can also see what people you trust might say about foo, or people
that these people trust say.

We could also say that you would like to find out the reputation of
something. For that you need to find people you trust, and their
rating of the thing.::

  
         foo
       /     \          Reputation
      /       \
 ----------------------
  DerT         tav
    |         / |
    | ___ A _/  |       Trust
    |/          |
    B-----------C

Here you can clearly and easily see that Trust is about finding the
connections to the right people (A,B,C,DerT,tav), while reputation is 
about the rating people make on things (foo)

Trust metric
------------

The trust metric is a function that takes a trust matrix as an input,
and returns you a trust value for an entity.

Reputation metric
-----------------

The reputation is similar to a trust matrix, but operates with

1. a reputation matrix - similar to a trust matrix, but for ratings
   and
2. the output from a trust metric

to find out an overall rating for the thing/entity/object you are
asking about. 

Usages
======

Where is this complete and fine described model used?

- entities - trust on entities
- reputation cube - the algorithms use trust
- index query - what indexes to ask
- plexnames - getting default names
- networking - who do I talk to in which way
- services - which services do I use
- s.a.f.e - who's code can I trust
- service migration - where to migrate services to


Alternative model to matrix
===========================

Instead of having a matrix with entity and domain as the axis, I could
also imagine a cube, with entities, aspect and domain as the
axis(es?)::
  

         |      /
         |     / Aspect
 Entity  |    /
         |   /
         |  /    x Value
         | /
         |/      
         ----------------
             Domain


Imagine a document - I could then say that the quality (aspect) in
regards to crypto (domain) of that paper (entity) is good, but when it
comes to the statements about coding (domain), the readability
(aspect) is quite low.

I would also recommend not to differentiate between trust metric and
reputation metric. It's two word for effectively the same thing.