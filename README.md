![Ampify](https://github.com/downloads/tav/ampify/logo.ampify.smallest.png)

This is the repo for Ampify — an open and decentralised app platform. It is  
intended as a successor to the Open Web and as a replacement for closed  
platforms like iOS and Facebook.

We're currently working towards an initial release for mid-2013, with a final  
1.0 release at the end of 2013 — [see the planfile for more info].

**Getting Started**

    $ make
    $ source environ/ampenv.sh

**Resources**

For more info, please see the `doc` directory or visit:

* online docs: [http://ampify.net]
* irc channel: [irc://irc.freenode.net/esp], [irc logs]
* mailing list: [http://groups.google.com/group/ampify]
* plan file: [http://plan.ampify.net]
* review server: [https://gitreview.com/ampify]

**Contribute**

We've made a tool called `git-review` to make it really easy for you  
to contribute. If you haven't used it before, just install it with  
the following command:

    $ sudo easy_install git-review

Then, to contribute any patches, simply create a new branch for your  
work:

    $ git checkout -b name-for-your-patch

And to submit it for review, make sure you've added yourself to the  
`AUTHORS` file and run:

    $ git review submit

That's it! Thanks.

**Please Note**

This is very much a work in progress and not much works yet — faster  
development is dependent on your involvement =)

**Credits**

All work by the [Ampify Authors] in this repository has been placed  
into the [public domain]. The major contributors so far have been:

* [tav] — creator of Ampify and BDFL.

* [Mamading Ceesay], evangineer — helped think through many of  
  the facets of Ampify.

* [Sean B. Palmer], sbp — implemented various aspects including  
  field trees; historian; even coined the name Ampify.

* [Mathew Ryden], oierw — designed many aspects of the overlay  
  network, crypto and networking protocols.

* [Tom Salfield], salfield — helped work through a lot of the  
  datastore and application programming layers.

* [Yan Minagawa], yncyrydybyl — pioneered experimentation with  
  many of the Ampify concepts and co-designed the interface.

* [Øyvind Selbek], talonlzr — designed aspects of the service  
  architecture, including video encoding.

* [James Arthur], thruflo — implemented various trust map  
  iterations; dolumns; invented thruflo transactions.

* [Seyi Ogunyemi], micrypt — working on actually implementing  
  the core of Ampify!

* [Maciej Fijalkowski], fijal — implemented the bridge between  
  WebKit and PyPy-based interpreters; JIT sandbox.

* [David Pinto], happyseaurchin — co-designed the micro-syntax  
  and elements of the minimal user interface.

* [Jeff Archambeault], jeffarch — the glue that binds us all;  
  Chief Shailar.

See the [authors], [credits] and [pecu allocations] for a full listing  
of all the awesome people who've helped over the years.

—  
Enjoy, tav <<tav@espians.com>>





[Xanadu]: http://en.wikipedia.org/wiki/Project_Xanadu
[see the planfile for more info]: http://plan.ampify.net

[Ampify Authors]: http://ampify.net/authors.html
[public domain]: http://ampify.net/unlicense.html

[authors]: http://ampify.net/authors.html
[credits]: http://ampify.net/credits.html
[pecu allocations]: http://tav.espians.com/pecu-allocations-by-tav.html

[http://ampify.net]: http://ampify.net
[http://groups.google.com/group/ampify]: http://groups.google.com/group/ampify
[https://gitreview.com/ampify]: https://gitreview.com/ampify
[http://plan.ampify.net]: http://plan.ampify.net
[irc://irc.freenode.net/esp]: irc://irc.freenode.net/esp
[irc logs]: http://irclogs.ampify.net

[David Pinto]: http://twitter.com/happyseaurchin
[James Arthur]: http://thruflo.com
[Jeff Archambeault]: http://www.openideaproject.org/jeffspace
[Maciej Fijalkowski]: http://morepypy.blogspot.com/
[Mamading Ceesay]: http://twitter.com/evangineer
[Mathew Ryden]: https://github.com/oierw
[Øyvind Selbek]: http://twitter.com/talonlzr
[Seyi Ogunyemi]: http://micrypt.com
[Sean B. Palmer]: http://inamidst.com
[tav]: http://tav.espians.com
[Tom Salfield]: https://twitter.com/tsalfield
[Yan Minagawa]: http://delicious.com/t
