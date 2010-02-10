---
license: Public Domain
layout: page
title: Zero Architecture
---

Zero Architecture
=================

Ampify Zero is building on top of existing technologies like EC2, jQuery, Git,
App Engine, Go, Node.js, Python, Redis, Ruby, S3, Sendgrid, Twilio and V8 as an
initial step towards the decentralised vision of version `1.0`.


Node Structure
--------------

A Node is started up using the `ampnode` executable.

On startup all nodes establish a connection to the Seed node.

<pre class="ascii-art">

       +----------------+
       | Internet Horde |
       +----------------+
             |                                  +-------------+
             |             +-----------+        | Other Nodes |
             ±             | Seed Node |        +-------------+
             |             +-----------+              |
             |              |     \                   |
         +-------------+    |      \                  |
         | Public Port |    |     +----------------------------------+
         +-------------+    |     | Meta Port (Internal Access Only) |
                \           |     +----------------------------------+
                 \          |       /
                  \         |      /
       +===========\========|=====/=================================+
       |            \       |    /                                  |
       |          +----------------------+                          |
       |          | Node: Parent Process |                          |
       |          +----------------------+                          |
       |                   |                                        |
       |                   |                                        |
       |    +-----------------------+-----------------------+       |
       |    |                       |                       |       |
       |  +---------------+         |            +---------------+  |
       |  | Child Process |  +---------------+   | Child Process |  |
       |  +---------------+  | Child Process |   +---------------+  |
       |                     +---------------+                      |
       |                                                            |
       +============================================================+

</pre>