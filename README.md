![Ampify](https://github.com/downloads/tav/ampify/logo.ampify.smallest.png)

This is the repo for Ampify — an open and decentralised social platform.  
For comparison, perhaps imagine a bastardised mix between Git, Facebook,  
Unix, IRC, App Engine, [Xanadu] and Wikis.

The goal for the final 1.0 release is to have a fully decentralised "internet  
operating system" style platform. We're currently working towards an initial  
release in May 2011 — [see the planfile].

**Getting Started**

    $ make
    $ source environ/ampenv.sh

**Resources**

For more info, please see the `doc` directory or visit:

* online docs: [http://ampify.it]
* mailing list: [http://groups.google.com/group/ampify]
* plan file: [http://ampify.it/planfile.html]
* irc channel: [irc://irc.freenode.net/esp], [irc logs]

**Contribute**

To contribute any patches, simply fork this repository using GitHub  
and create a new branch for your work:

    $ git checkout -b name-for-your-patch

And to submit it for review, make sure you've added yourself to the  
`AUTHORS` file and then run:

    $ git review submit

That's it! Thanks.

**Please Note**

This is very much a work in progress and not much works yet — faster  
development is dependent on your involvement =)

**Credits**

All work by the [Ampify Authors] in this repository has been placed  
into the [public domain]. The major contributors so far have been:

* [tav] — creator of Ampify and BDFL.

* [Mamading Ceesay], evangineer — co-architect.

* [James Arthur], thruflo — implemented various trust map  
  iterations; dolumns; invented thruflo transactions.

* [Sean B. Palmer], sbp — implemented various aspects including  
  field trees; historian; even coined the name Ampify.

* [Mathew Ryden], oierw — designed many aspects of the overlay  
  network, crypto and networking protocols.

* [Yan Minagawa], yncyrydybyl — pioneered experimentation with  
  many of the Ampify concepts and co-designed the interface.

* [Maciej Fijalkowski], fijal — implemented the bridge between  
  WebKit and PyPy-based interpreters; JIT sandbox.

* [Øyvind Selbek], talonlzr — designed aspects of the service  
  architecture, including video encoding.

* [David Pinto], happyseaurchin — co-designed the micro-syntax  
  and elements of the minimal user interface.

* [Jeff Archambeault], jeffarch — the glue that binds us all;  
  Chief Shailar.

See the [authors], [credits] and [pecu allocations] for a full listing  
of all the awesome people who've helped over the years.

—
Enjoy, tav <<tav@espians.com>>





[Xanadu]: http://en.wikipedia.org/wiki/Project_Xanadu
[see the planfile]: http://ampify.it/planfile.html

[Ampify Authors]: http://ampify.it/authors.html
[public domain]: http://ampify.it/unlicense.html

[authors]: http://ampify.it/authors.html
[credits]: http://ampify.it/credits.html
[pecu allocations]: http://tav.espians.com/pecu-allocations-by-tav.html

[http://ampify.it]: http://ampify.it
[http://ampify.it/planfile.html]: http://ampify.it/planfile.html
[http://groups.google.com/group/ampify]: http://groups.google.com/group/ampify
[irc://irc.freenode.net/esp]: irc://irc.freenode.net/esp
[irc logs]: http://irclogs.ampify.it

[David Pinto]: http://twitter.com/happyseaurchin
[James Arthur]: http://thruflo.com
[Jeff Archambeault]: http://www.openideaproject.org/jeffspace
[Maciej Fijalkowski]: http://morepypy.blogspot.com/
[Mamading Ceesay]: http://twitter.com/evangineer
[Mathew Ryden]: https://github.com/oierw
[Øyvind Selbek]: http://twitter.com/talonlzr
[Sean B. Palmer]: http://inamidst.com
[tav]: http://tav.espians.com
[Yan Minagawa]: http://delicious.com/t