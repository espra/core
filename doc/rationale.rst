---
layout: page
license: Public Domain
title: Rationale
---

Rationale
=========

Ampify is an open source competitor to the various proprietary platforms that
currently threaten to fragment the Open Web [Neuberg-2008]_. It addresses the
fact that technological complexity is fast becoming unmanageable:

.. class:: math center

  `Platforms`  ✕  `APIs`  ✕  `Devices`  ＝  `Brittle.Complexity`

Innovation slows down as developers try to assess, learn, code, test and deploy
for the various platforms -- again and again and again. Our collective efforts
are wasted on working around technical incompatabilities and shortcomings.

.. raw:: html

    <div class="center"><img
      src="http://static.ampify.net/img.dev-platform-options-2009.png" alt=""
      width="512px" height="384px" class="boxed"
      /></div>

We desperately need a period of convergence at a much higher level than what we
currently get from web browsers. Incompatibility on issues like identity,
structured data, media delivery and interfaces keep giving opportunities for
closed platforms to gain adoption.

The fact is, the Web is over 20 years old [Berners-Lee-1989]_ and was never
developed with today's applications in mind. As a result countless developers
have had to independently solve the same problems over and over, e.g. messaging,
scalability, security, &c.

.. class:: sidebox

  "Ease developer pain and take innovation to the next curve"

We need solutions that will take innovation to the next curve [Kawasaki-2006]_.
Ampify is one such solution. It builds on top of the Open Web principles and
will hopefully be a worthy successor to the Web someday.

The design and development of Ampify is driven by a few key principles:

* Decentralisation and openness.
* Ease of use.
* Speed, security and scalability.
* Simplicity beyond complexity.

Wherever appropriate, Ampify makes use of existing open source technologies as
much as possible, e.g. Caja_, Chromium_, CoffeeScript_, Dirac_, FFmpeg_,
FreeSWITCH_, Git_, Go_, jQuery_, Keyspace_, Mapnik_, `Native Client`_, Node.js_,
PyPy_, Python_, QuantLib_, Redis_, Ruby_, Sizzle_, Theora_, V8_, WebM_, &c.

Ampify is without doubt a very ambitious undertaking. And, despite having had
over 10 years of research, it is going to be quite a challenge to pull off. But
if you'd like to help make it a reality, do join us on the ``#esp`` channel on
``irc.freenode.net``. All are welcome!

.. raw:: html

  <div style="text-align: center; margin-top: 9px;">
    <form action="http://webchat.freenode.net/" method="get">
      <button style="padding: 2px 6px 3px;">Click to join #esp</button>
      <input type="hidden" name="channels" value="esp" />
   </form>
  </div>


References
----------

.. [Berners-Lee-1989]

    `Information Management: A Proposal
    <http://www.w3.org/History/1989/proposal.html>`_

    Tim Berners-Lee, CERN, 1989.

.. [Kawasaki-2006]

    `The Art of Innovation
    <http://blog.guykawasaki.com/2006/01/the_art_of_inno.html>`_

    Guy Kawasaki, 2006

.. [Neuberg-2008]

    `What Is the Open Web and Why Is It Important?
    <http://codinginparadise.org/weblog/2008/04/whats-open-web-and-why-is-it-important.html>`_

    Brad Neuberg, April 2008.

.. _Caja: http://code.google.com/p/google-caja/
.. _Chromium: http://www.chromium.org
.. _CoffeeScript: http://jashkenas.github.com/coffee-script/
.. _Dirac: http://diracvideo.org
.. _FFmpeg: http://ffmpeg.org
.. _FreeSWITCH: http://www.freeswitch.org
.. _Git: http://git-scm.com
.. _Go: http://golang.org
.. _jQuery: http://jquery.com
.. _Keyspace: http://scalien.com/keyspace/
.. _Mapnik: http://mapnik.org
.. _Native Client: http://code.google.com/p/nativeclient/
.. _Node.js: http://nodejs.org
.. _PyPy: http://codespeak.net/pypy/dist/pypy/doc/
.. _Python: http://www.python.org
.. _QuantLib: http://quantlib.org
.. _Redis: http://code.google.com/p/redis/
.. _Ruby: http://www.ruby-lang.org
.. _Sizzle: http://sizzlejs.com
.. _Theora: http://www.theora.org
.. _V8: http://code.google.com/p/v8/
.. _WebM: http://www.webmproject.org
