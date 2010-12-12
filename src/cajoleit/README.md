The Cajoleit API Server provides an easy to use HTTP API for transforming  
raw JavaScript and CoffeeScript into "cajoled" code. See the docs for more  
info: [http://ampify.it/cajoleit.html]


**Installing Your Own Instance**

* Create a new application with [Google App Engine].

* Update `appengine.xml` with the application ID.

* Run `./deps` to grab the various dependencies.

* Run `ant runserver` to test that everything works in your local instance.

* Run `ant update` to deploy to a production instance at App Engine.


—  
Enjoy, tav <<tav@espians.com>>



[http://ampify.it/cajoleit.html]: http://ampify.it/cajoleit.html
[Google App Engine]: http://code.google.com/appengine/