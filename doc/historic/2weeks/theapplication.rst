This is about an application that can be hopefully sold to customers,
which would be organizations. 

Structure
=========

The application is, as you may have guessed, built on top the plex. It
consists basically of these elements:

Entity management
  Take care of entities - create, edit, modify or delete entities,
  which to a large extend equals users.

Knowledge base
  A place to create, edit, keep, share and rate information.

Exchange
  A system to transfer (representations of) resources from one point
  to the other, e.g. funds,people's work time or a ticket system.

Decision making
  A system to facilitate decision making processes. Supports different
  forms of voting.  

File sharing
  Versioned sharing of files. Not a full source control. 

Communication
  Something like instant messaging


Components
==========

Lets have a look at the components required to create such an application.

Plex
----

No, really? 


Shailas
-------

Used in: knowledge base.

"Summaries forked across different perspectives" (tav) . For example
you have a document about the plex. The text will be provided in
different  languages, levels of complexities or styles for certain 
viewpoints. An English investor might want to prefer another document 
as the Dutch application developer.
Which basically means that you have different texts which have
attributes describing the perspectives on it tied together in a node.
Or maybe having a main document that links to other documents saying
'this is the english version'.

Plexedit
--------

Used in: knowledge base.

Collaborative editing component. Moonedit, SubEthaEdit or the Textpad
in groove would be examples. Work together at the same time on the
same document, see the others typing. Not an absolute requirement, but
something that would be really useful.


Annotation
----------

Used in: knowledge base, but also at other places.

Write an annotation, and get the objects that are annotated informed
of the annotation being there (using events).


Exchange system
---------------

Used in: exchange system and decision making.


Something that allows you to shift around resources. The trick is that
you can think of a decision making system as a system for exchanging
votes, or a ticket system a system for exchanging tickets. It will
need:

Resource generator
  Something that generates your resources (based on whatever
  algorithm). May take real world into account - for example not
  generating more money than money put on a certain bank account.

Basic accounting
  Keeping track of which resource is where.

Interface
  Something to actually select resources, where to put them etc.

The idea is that you have pluggable algorithms for this system  - a
voting system maybe generates one vote for each entity and allows you
only to transfer it once.


Chat
----

Used in: communication

Like you normal chat. But it basically means that you send out an
event somewhere, which then is spread to all the participants of the
communication (which can be only one). Your normal plex event system
is going to be used here. 

Pecus
=====

As we would of course like to use the system ourselves (don't we?), it
would great to have a system for using pecus in. Which works like
that: depending on the amount and quality of work you have done in the
last time period (tp), the reputation of you and other factors you get 
the right to issue pecus. You can accumulate this right, but the
amount will slowly decrease over time. You can then assign pecus to
tasks, and the people doing these tasks will get the pecus transfered
to their earnings account. The pecus can't leave the earnings account.
Now depending on you share of issued pecus in that tp the share of the
profit that esp is making is determined. (Basic needs are covered anyhow).

Target ui
=========

As we don't have the ueberlein interface ready a the moment, the idea
would be to use irc as the sample interface. Should everyone force to
make the interface as small and simple as possible. :-)

