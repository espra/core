![Ampify](https://cloud.github.com/downloads/tav/ampify/logo.ampify.smallest.png)

This is the repo for Ampify — an open and decentralised app platform. It is  
intended as a successor to the Open Web and as a replacement for proprietary  
platforms like Facebook and iOS.

**Please Note**

This is very much a work in progress and not much works yet — faster  
development is dependent on your involvement =)

We're currently working towards an initial release in early 2016, with a  
production-ready 1.0 release at the end of 2016.

**Community**

Please join us on `#esp` on `irc.freenode.net`:

* irc channel: [irc://irc.freenode.net/esp], [irc logs]

**Contribute**

We're writing a tool called `revue` to make it really easy for you to  
contribute in a way that's easy to test and code review. You will be able  
to install `revue` with the following command:

```bash
$ go get github.com/tav/revue
```

Then, to contribute any patches, simply create a new branch for your  
work:

```bash
$ revue open
```

And to submit it for review, make sure you've added yourself to the  
`AUTHORS` file and run:

```bash
$ revue submit
```

That's it! Thanks.

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

And, in addition, thanks to [BrowserStack] for providing a free account  
for us to do cross-browser testing.

—  
Enjoy, tav <<tav@espians.com>>


[Ampify Authors]: https://github.com/tav/ampify/blob/master/AUTHORS.md
[authors]: https://github.com/tav/ampify/blob/master/AUTHORS.md
[BrowserStack]: http://www.browserstack.com/
[credits]: https://github.com/tav/ampify/blob/master/doc/credits.md
[irc://irc.freenode.net/esp]: irc://irc.freenode.net/esp
[irc logs]: http://irclogs.ampify.net
[pecu allocations]: http://tav.espians.com/pecu-allocations-by-tav.html
[public domain]: https://github.com/tav/ampify/blob/master/UNLICENSE.md

[David Pinto]: https://twitter.com/happyseaurchin
[James Arthur]: https://github.com/thruflo
[Jeff Archambeault]: https://github.com/jeffarch
[Maciej Fijalkowski]: https://github.com/fijal
[Mamading Ceesay]: https://twitter.com/evangineer
[Mathew Ryden]: https://github.com/oierw
[Øyvind Selbek]: https://twitter.com/talonlzr
[Seyi Ogunyemi]: https://github.com/micrypt
[Sean B. Palmer]: https://github.com/sbp
[tav]: http://tav.espians.com
[Tom Salfield]: https://github.com/salfield
[Yan Minagawa]: https://github.com/yncyrydybyl
