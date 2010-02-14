---
license: Public Domain
layout: page
title: Zero Architecture
---

Zero Architecture
=================

Ampify Zero is building on top of existing technologies like Caja, jQuery, Git,
App Engine, Go, Node.js, Python, Redis, S3, Sendgrid, Twilio and V8 as an
initial step towards the decentralised vision of `Ampify 1.0`. It is intended as
a working demo and nothing more.


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
       +===========\========|=====/====================================+
       |            \       |    /                                     |
       |          +----------------------+                             |
       |          | Node: Parent Process |                             |
       |          +----------------------+                             |
       |                   |                                           |
       |                   |                                           |
       |    +-------------------------+-----------------------+        |
       |    |                         |                       |        |
       |  +----------------+          |            +----------------+  |
       |  | Nodule Process |  +----------------+   | Nodule Process |  |
       |  +----------------+  | Nodule Process |   +----------------+  |
       |                      +----------------+                       |
       |                                                               |
       +===============================================================+

</pre>

Argonought
----------

Argonought is a JSON-inspired serialisation format that offers us features not
available in other formats:

* JSON-like simplicty.
* Efficient binary encoding.
* Lexicographically sortable encoding.

As an example, consider how you would sort *really* long numbers on datastores
like Google App Engine or SimpleDB. The latter only allows you to store strings
and the former truncates anything greater than 64bits.

{% highlight irb %}
>> (2 ** 29).class
=> Fixnum

>> (2 ** 30).class
=> Bignum
{% endhighlight %}


The official documentation for SimpleDB
[recommends](http://developer.amazonwebservices.com/connect/entry.jspa?categoryID=152&externalID=1232)

{% highlight pycon %}
>>> '9' > '8'
True

>>> '10' > '9'
False
{% endhighlight %}

Argonought takes care of all of this hassle:


{% highlight pycon %}
>>> argonought.pack_number(8234364)
'\xfe\xa2\xa0'

>>> a.pack_number(8234364) > a.pack_number(-234364)
True

{% endhighlight %}

Message
-------

The message looks like:

{% highlight go %}
type Message struct {
    from         string
    by           string
    to           string
    aspect       string
    content      string
    value_number big.Number
    value_list   []string
    version      int
}
{% endhighlight %}

For example, the following message written by `@tav` and sent to `#espians`:

    /balance ~/account

Would like the following in Ruby after being parsed:

{% highlight ruby %}
{
  :from => "tav",
  :aspect => "/balance",
  :to => "#espians",
  :value_number => 9203180132
}
{% endhighlight %}

Hmz

Asr

{% highlight pycon %}
>>> nodule.init()
[0, 'START']

>>> status == 12
True

>>> def foo():
...     if N_set: return True
[0, 'START']

{% endhighlight %}
