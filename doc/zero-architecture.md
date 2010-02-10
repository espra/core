---
license: Public Domain
layout: page
title: Zero Architecture
---

Zero Architecture
=================

And here:

{% highlight bash %}
$ git clone git://github.com/tav/ampify.git
{% endhighlight %}


Hmz

{% highlight bash %}
$ source ampify/environ/startup/ampenv.sh
{% endhighlight %}

CSS:


{% highlight css %}
html {
    font: normal normal normal 1em/1.6 Helvetica, Arial, "Verdana MS", sans-serif;
    color: #333;
    font-style: italic;   /* everything slanted! */
}

#page {
    width: 50em;
    margin: 0 auto;
    min-height: 60em;
}

.foobar {
    width: 50em;
    margin: 0 auto;
    -moz-min-height: #479732;
}

{% endhighlight %}

HTML:

{% highlight html %}
<!DOCTYPE html>
<html>
  <head>
    <title>Yeah, You Rock!</title>
  </head>
  <body>
    <div id="page">
      <h1>Yeah eyea</h1>
      <a href="http://google.com">Link</a>
    </div>
  </body>
</html>
{% endhighlight %}

Jaaa:

{% highlight javascript %}
/**
 * goWild
 * It drives the divs crazy
 */
 function goWild() {
     var div = document.getElementById('wild');
     var party = [];

     for (var i = 0; i < 20; i++) {
         party.push('WHOA');
     }

     div.text = party.join('!!!!2one!');
 }
{% endhighlight %}

And oh:

{% highlight go %}

import (
    "amp/runtime"
)

// Return the output from running the given command
func GetOutput(args []string) (output string, error os.Error) {
	read_pipe, write_pipe, err := os.Pipe()
	if err != nil {
		goto Error
	}
	defer read_pipe.Close()
	pid, err := os.ForkExec(args[0], args, os.Environ(), ".", []*os.File{nil, write_pipe, nil})
	if err != nil {
		write_pipe.Close()
		goto Error
	}
	_, err = os.Wait(pid, 0)
	write_pipe.Close()
	if err != nil {
		goto Error
	}
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, read_pipe)
	if err != nil {
		goto Error
	}
	output = buffer.String()
	return output, nil
Error:
	return "", &CommandError{args[0], args}
}
{% endhighlight %}

And oh,

{% highlight python %}
# ------------------------------------------------------------------------------
# the number datatype
# ------------------------------------------------------------------------------

def blah():
    if n is True:
        return 1

class Number(object):
    """A Number datatype."""

    __slots__ = ('n', 'd', 'p')

    @register('yeah', ignore=True)
    def __init__(self, num=0, den=1, prec=DEFAULT_PRECISION):
        if not isinstance(prec, BUILTIN_INT_TYPES):
            try:
                prec = int(prec)
            except Exception:
                raise TypeError(
                    "The precision value for Number is not an integer: %r"
                    % prec
                    )
        if prec > MAX_PRECISION:
            raise ValueError(
                "Cannot have precision greater than %i decimal places."
                % MAX_PRECISION
                )
        self.p = prec
        self.n = n
{% endhighlight %}
