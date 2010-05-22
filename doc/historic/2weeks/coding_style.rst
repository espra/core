Python
======

We use Python 2.4 Release

Language
========

Plain English is going to be used.

From the irc:

< oierw> or whatever langauge that actually exists and isn't tav's
perversion of something that works.

Functional style
================

Classes
-------

Classes are only used where necessary, as opposed to whenever
possible. Instead a more functional style approach is used. One reason
is to have better maintainable code. Also, when it comes to
Interfaces, they mostly belong to the Java world - but lucky for us,
we live in the python world.
XXX example of the form: with classes like that, but we use this.

Decorators
----------

We are going to use decorators quite extensively. This comes along with
the functional style of coding. The reason is to have better
maintainable code, and they are also needed for the events. XXX
example.

Function attributes
-------------------

Did you know you can attach attributes to functions as if they were
intances of classes?::
  
  def foo():
      print foo.bar

  foo.bar = 'hello world'
  foo()


Naming
======

For the naming we are using CamelCase only for classnames, for other
purposes underscores are used.

function names
  All lowercase. First word is a verb. `read_this`

variables
  Lowercase, but a noun. `entity`

Classes
  CamelCase. `MyClass`

Interfaces
  Like a class, but with a capital 'I' in front of it.
  `ISuperInterface`


Files
=====

The code is written using *no tabs*. Instead 4 spaces are used for
indentation. Assume the width of the text to be 76 characters, but
this is not a must: you can also have long lines. But if you wrap, 76
is your number. All the text files are to be unicode.

Importing
=========

When importing, try to avoid `from foo import *`. Do direct imports of
the form `from foo import bar`. Exception: `from pyutil import *`.
Pyutil is the package where all our nice nifty utilities reside. 



Generators + continuation
=========================

Generators are used quite a lot. Example for a generator::

  def foo():
      yield 'hello'
      yield 'world'

  g = foo()
  print g.next()
  print g.next()

This will print::

  hello
  world

foo is a function, g is a generator.

This means, that whenever `g.next` is called, python proceeds in foo, until a
yield is reached, and continues from there on subsequent calls to
`g.next`. 

An example from the code would be XXX example.

The problem is that parameters for the function are only evaluated
when the generator is created. But one might also want to pass some
data or do things between calls. Enter the continuation hack.

XXX continuation hack. The basic idea is that you bind the generator
to a class, so that you can use self.foo attributes within the
generator. They are then also accessible from outside.

And to make things even more interesting, there is the pyutil.copy
tool to copy generators - and preserving the state.

overloaded operators
====================

To make things more readable two operators are overloaded: the
bitshift operator '>>' and the symbol for logical 'or': '|'. The
latter one is now used as a pipe symbol. 
XXX example. 

The former one is used to XXX example

doctest
=======

As a combination of documentation and testing framework doctest is
going to be used.



