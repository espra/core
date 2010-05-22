ueberinterfacecontroller - I love long words, I am german. But we also
might call it uic, if you insist.


After more or less knowing by know what the plex is about, the
question arises on how to actually use it. Where do I click? What
displays an image? Am I using a browser, or do I use the command line?


We decouple the interface from the application logic - so that you can
write an application that supports a command line as well as a web
display or a fance 3d interface.

Interface model
===============

The interface model is pretty much mvc - model, view controller.

model
  The actual data we are working on - like the text in a document

view
  The display of the text. The text is just a huge collection of
  (hopefully not completely random) bytes. A view is the actual
  display, e.g. on the screen.

controller
  This modifies and controls the model - the actual data.

The basic idea is now to have basically two sets of mvc - on on the
server, one on the client.::

  Client Side                             Server Side

       local view                           view
                 \                         /
  User           helper----------------uic
                /      \                   \ 
            input       \___________________service

XXX yep, there should be a decent graphic

What does that mean? The service (server side) provides the actual
application. But its not rendering that itself, e.g. to an html page.
Instead a helper is placed on the client side (Javascript) that is
able to deal with user input and can place widgets on the screen. On
the server side there is a uic that manages most of the information
flow between the helper and the service. But sometimes there is also
direct communication between the service and the helper.
The view on the server side is a service transforming data to
displayable stuff, which is then transfered to the client. 

The most important thing to have in mind is that just as with the
whole disco working thing events are flowing back and forth between
the components. This especially means that you don't load a new page
in your browser for every click - instead a javascript picks up the
event, transfers it to the server, which might then send a view to the
client which contains a picture or whatever.

Example
=======

XXX t, help me with that

You have a picture application. You use a web browser as the client.
The application contains a big button, that says next picture.

1. An 'empty' page gets transfered to the client, containing the
   necessary javascript and css files, so that now communication
   between the server and the client can take place. Especially the
   helper and some initial variables are now on the client

2. The helper creates the initial button on the page

3. You click the button

4. The helper picks up the javascript event, and sends out an event to
   the uic.

5. The uic transfers the event to the service, which grabs a new
   picture.

6. The picture  gets transfered to the view, which
   creates a nice html with the picture and its caption (using events
   of course)

7. The result is then transfered to the uic, which sends it to the helper.

8. The helper places the html information on the client side.

Voila.

The biggest import thing is no notice the decoupling of the 'click'
and the answer. With the click you trigger an event, which gets
transfered. An incoming event might trigger the display of a new
picture. You could also imagine the service on the server side sending
out new pictures every 20 seconds. Full dia show mode...


Existing toolkits
=================

A short discussion was how to deal with existing toolkits - how do we
integrate existing applications of your platform of non-choice, or how
do we integrate into them?

- Using the filesystem. Use plexdrive to mount the plex as it was a
  drive, just the webfolders do.

- Provide nice widgets that work with the plex

- Try to hook up with the help system.

- Provide a good enough helper api so that third party developers can
  just write decent helpers for the applications.




