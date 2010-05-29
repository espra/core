One of the topics I was most interested in - how is data actually
stored in the indexes and how it can be queried.

An example
==========

Lets say, we had one big question::

  What are the top songs by female singers in the last 30 years?

How would it be answered?

Index structure
---------------

For writing out the indexes, we use a little shortcut. Instead of
writing::

 10 foo:bar    20 alias_for:10/1
 ----------    -----------------
 1  ...  ()    1  ...    michael
 2  ...  ()

we just write::

 10 foo:bar
 ------- -------
 1  michael   ()


Ok? Let's start.::


 10 is_a:metatype  # Our metatypes
 ----------------
 1  artist     ()  
 2  song       ()
 3  topt_chart ()
 

 20 properties_of:artist
 -----------------------
 1  gender            ()

 21 properties_of:songs
 ----------------------
 1  title            ()
 2  artist           ()

 22 properties_of:topt_chart
 ---------------------------
 1  chartdate             ()
 2  top ten               ()

Now we have our types, and their properties. Next we need some artists::
 
 40 is_a:artist
 ----------------
 1  prince     ()
 2  madonna    ()
 3  m. jackson ()
 4  b. spears  ()
 5  z style    ()
 6  p. hilton  ()

And their gender::

 50 gender_of:prince
 -------------------
 1              male

 51 gender_of:zstyle
 -------------------
 1            female
 
 ...

Some songs and their titles::

 100 is_a:song      110 title_of:100/1
 -------------      ------------------
 1          ()      1      purple rain
 2          ()
 3          ()      111 title_of:100/2
 4          ()      ---------------------
 5          ()      2  one night in paris
       
 ...

Also, we need to know the artists for a song::
 
 120 artist_for:100/1   121 artist_for:100/2
 --------------------   --------------------
 1               40/1   1               40/4
                        2               40/6

Next, some charts, and their dates::

 140 is_a:topt_chart  150 chartdate:120/1
 -------------------  -------------------
 1                ()  1            1/1970
 2                ()  
 3                ()  151 chartdate:120/2
                      -------------------
 ...                  1            2/1985


Now, the actual chart entries::

 150 top_ten_for:140/1  151 top_ten_for:140/2
 ---------------------  ---------------------
 1               100/1  1               100/2
 2               100/3  2             2005/27 #yes, some other index

 ...



Query language
==============

There are some keywords used by the query language:

- for
- in
- is
- not
- if

Using these keywords the query on indexes looks quite familiar to the 
python minded reader::

  for chart in topt_charts
    for song in chart.top_ten
      if song.artist.gender = "female" and chart.date>1974
        print song.title

To actually get that running it's quite clear that the query engine
has to do a lot of checking for metatypes under the hood. If it
encounters the `for chart in topt_charts` it will look if `topt_chart`
iss a metatype, and then treat `chart.top_ten` accordingly. 

Speeding up
===========

Reversed indexes
----------------

Ok, all those indexes where  quite easy to understand, and easy to
have an overview of? 

Well, if so, we might just as well add a couple of indexes to speed up 
the query (see below). The rule is quite simple - for every::

 property1:value1
 ----------------
 item1
 item2
 ...

We create also::

 item1:property1   item2:property1
 ---------------   ---------------
 value1            value2

So, we are having Property-Value-Item, and Item-Property-Value. It
might be necessary to create even more indexes, but this is defered to
future.

Using this rule, we also create::

 200 40/4:artist  #b.spears:artist
 ---------------
 1         100/2  #song          

 ...

The good thing  is that "indexes cost nothing" (tm), (c) tav. :-)


Grouped indexes
---------------

To speed up things even more we have grouped indexes::

 300 (is_a:artist + artist.gender=="female")
 -------------------------------------------
 1                                index/40/5
 ...

These will be generated on the fly, and updated as needed. How is the
update done? By putting lots of sensors everywhere. You update an
index - it will fire of an event which gets noticed by the other
indexes, which will then update themselves. 

Distributed query
=================

Of course query can span more than one entity - the question is, how
the query is going to be implemented. There seem to be two ways:

1. Grab all indexes from the other entities that might become
   important to your query, and do the necessary intersections on your
   local entity.

2. Pass on the query, get the results, and do maybe a last
   intersection on the local entity.

Most likely it's going to be a mixture. The distributed query part
seems to be a part which has lot of potential to optimization.


