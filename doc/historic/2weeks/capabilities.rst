Capabilities are used as the mechanism to grant rights on resources,
e.g. using a service, connecting to a node etc.

Content
=======

A capability is a bit like a certificate. It's a signed paper, but
this time the issuer signs:

Permission
  What purpose is the capability for - 'view picture', 'attach sensor'
  etc. Like rights in zope or more general in acls. In reality of
  course strings are not used, but unique id's

Constraints
  What other things have to be fulfilled before the capability can be
  used? Does it have to be monday, needs some money to be on a certain
  account, is the use of the capability limited to a certain entity?
  Effectively the constraint is a piece of code that will check stuff.

Who it is from
  The signature

Storage
=======

Now this is a main difference to a certificate - while the certificate
is issued to the entity requesting it, a capability can be kept with the
issuing entity, and the entity requesting it only gets a pointer to
the capability. A bit like a normal rights management entry on the
server.  This means that the receiver can not pass on the capability
nor can it show the capability to other entities. 

But if the issuer decides to hand out a capability, it behaves a bit
like a certificate - it might be shown to others, or passed on, and
even be used by other entities - if the constraints allow that. 

Notes
=====

Capabilities are not ACLs, which are thought to be to hard to manage
if every user has to define his own ACLs for all the other users.
Capabilities don't have a role.

They use tons of permissions - which will allow very fine grained
access. 

We are aware that capabilities that are handed out can't be
effectively revoked. 


Links
=====

- erights.org
- e language
- eros - e operating system