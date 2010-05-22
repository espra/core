The problem
===========

Imagine to functions - 'foo' and 'bar'. If foo should work on the
output of bar, it would somethink like `foo(bar())`. Pretty easy, and
no problem if the two are on the same machine. Now, what happens if
you have to seperate those two to different machines? You have to
implement some form of mechanism so that they can talk to each other.
Messages are going to be sound around, using a mechansim of choice. 

Wouldn't it be wise to use this message passing mechanism from the
beginning, so that later, when you seperate foo and bar, you don't
need to rewrite the code? 

The idea
========

The idea is to use message passing from the very beginning - the
services communicate using messages, instead of using the normal
calling mechanism of you language. 

Cool, now we have messages. A bit like sending letters between people.
The problem now is that you basically just have one mailman in your
program - the python interpreter that is running from top to bottom.
Delivers a message to bar, wait till that has written it's answer, and
then runs of with this new message to foo. 

The idea now is to make the work in foo, bar and the message delivery
concurrent. While bar works on the reply to the message, the postman
is delivering new messages, and maybe foo works on its own stuff. 


The implementation
==================

Of course we are not allowed to use threads. They are defined to be
evil. Me don't like evil. Me like music. Disco. Distributed concurrent
workers. 

Generators
----------

First step is to make use use of generators (see coding style). So
whenever possible or meaningful, a function allows the interpreter
(the postman) to go off and do something else. ::

 def foo(x):
    
     # do complicated stuff
     y=x+2
     yield 1
    
     # more complicated work
     z=x^^2  
     yield 1
    
     #more nasty stuff
     out=x+y+z
     yield 0

 def bar(x):
     for i in range(x):
         j=i+1
         if i%1000:
            yield 1
     out=j
     yield 0

 x= 10000
 f=foo(x)
 b=bar(x)

Now, by repetative calling of f.next() and b.next() both get worked
on, but they don't block each other

Modifiable generators
---------------------

Next step is to have them modifiable. What if you want to modify the x
in foo(x) after having done the second f.next()? For that they get
basically wrapped in a class, by using nifty decorators.

XXX Example code here.


Message passing
---------------

Last idea is to have the message passing going on. A description of
the basic iea can be found at http://kamaelia.sf.net, especially the
axon part. Have a look at the slides. What basically happens is that
each service gets a couple of mailboxes - inbox, outbox, errorbox etc.
Then you can link two boxes together - b.outbox -> f.inbox. Some
process that takes care of the message delivery.

XXX Real descripion has to follow

Now, once this is established, it is of now matter at all where the
services are located - on the same machine, or at the other end of the
internet. Messages are passed, and the services don't have to care how. 
