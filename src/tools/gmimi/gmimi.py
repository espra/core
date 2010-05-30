# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

"""
==============================
Contact Management Utility App
==============================

"""

from cgi import escape
from StringIO import StringIO
from urllib import quote, urlencode
from urllib2 import urlopen

try:
    from xml.etree import cElementTree as ElementTree
except ImportError:
    try:
        import cElementTree as ElementTree
    except ImportError:
        try:
            from xml.etree import ElementTree
        except ImportError:
            from elementtree import ElementTree

import atom

from gdata import contacts
from gdata.contacts.service import ContactsService, ContactsQuery

from google.appengine.api import users
from google.appengine.ext import db
from google.appengine.ext import webapp
from google.appengine.ext.webapp.util import run_wsgi_app

# ------------------------------------------------------------------------------
# App Configuration
# ------------------------------------------------------------------------------

# We define the priority accounts.
PRIORITY = set(['tav@espians.com', 'sofia@turnupthecourage.com'])

# We define headers and footers using super advanced string templating!
TEMPLATE_HEADER = """<!DOCTYPE html>
<html>
<body>
"""

TEMPLATE_FOOTER = """</body></html>"""

# ------------------------------------------------------------------------------
# Datastore Models
# ------------------------------------------------------------------------------

# The app is built on top of 5 simple models:
#
# * Account
# * Contact
# * Group
# * List
# * Subscription
# <yatiblog.comment>

# An Entity Kind representing an Account.
class A(db.Model):

    access = db.StringListProperty(default=None, name='a')
    email = db.StringProperty(name='e')
    hosted = db.BooleanProperty(default=False, name='h')
    imported = db.BooleanProperty(default=False, name='i')
    latest = db.StringProperty(name='l')
    modified = db.DateTimeProperty(auto_now=True, name='m')
    owner = db.StringProperty(name='o')
    password = db.StringProperty(name='p')

# An Entity Kind representing a Contact.
class C(db.Model):

    account = db.StringProperty(name='a')
    data = db.BlobProperty(name='d')
    id = db.StringProperty(name='i')
    refs = db.StringListProperty(default=None, name='r')

# An Entity Kind representing a Group.
class G(db.Model):

    account = db.StringProperty(name='a')
    id = db.StringProperty(name='i')
    name = db.StringProperty(name='n')

# An Entity Kind representing a (Mailing) List.
class L(db.Model):

    account = db.StringProperty(name='a')
    id = db.StringProperty(name='i')
    modified = db.DateTimeProperty(auto_now=True, name='m')
    query = db.TextProperty(name='q')
    refs = db.StringListProperty(default=None, name='r')

# An Entity Kind representing a (Profile) Picture.
class P(db.Model):

    account = db.StringProperty(name='a')
    contact = db.StringProperty(name='c')
    data = db.BlobProperty(name='d')

# An Entity Kind representing a Subscription.
class S(db.Model):

    contact = db.StringProperty(name='c')
    lists = db.StringListProperty(default=None, name='l')

# We use short names to define the Entity Kinds as that saves space on the
# datastore and then alias them to more descriptive classes for general use.
Account = A
Contact = C
Group = G
Picture = P
List = L
Subscription = S

# ------------------------------------------------------------------------------
# A Client For The MadMimi API
# ------------------------------------------------------------------------------

DEFAULT_CONTACT_FIELDS = '"first name","last name","email","tags"'

class MadMimi(object):
    """
    The client is straightforward to use:

      >>> mimi = MadMimi('user@foo.com', 'account-api-key')

    You can use it to list existing lists:

      >>> mimi.lists()
      <lists>
        <list subscriber_count="712" name="espians" id="24245"/>
        <list subscriber_count="16" name="family" id="76743"/>
        <list subscriber_count="0" name="test" id="22103"/>
      </lists>

    Delete any of them:

      >>> mimi.delete_list('test')

    Create new ones:

      >>> mimi.add_list('ampify')

    Add new contacts:

      >>> mimi.add_contact(['Tav', 'Espian', 'tav@espians.com'])

    Subscribe contacts to a list:

      >>> mimi.subscribe('tav@espians.com', 'ampify')

    See what lists a contact is subscribed to:

      >>> mimi.subscriptions('tav@espians.com')
      <lists>
        <list subscriber_count="1" name="ampify" id="77461"/>
      </lists>

    And, of course, unsubscribe a contact from a list:

      >>> mimi.unsubscribe('tav@espians.com', 'ampify')

      >>> mimi.subscriptions('tav@espians.com')
      <lists>
      </lists>

    """

    base_url = 'http://madmimi.com/'

    def __init__(self, username, api_key):
        self.username = username
        self.api_key = api_key

    def get(self, method, **params):
        params['username'] = self.username
        params['api_key'] = self.api_key
        url = self.base_url + method + '?' + urlencode(params)
        return urlopen(url).read()

    def post(self, method, **params):
        url = self.base_url + method
        params['username'] = self.username
        params['api_key'] = self.api_key
        return urlopen(url, urlencode(params)).read()

    def lists(self, as_xml=True):
        response = self.get('audience_lists/lists.xml')
        if as_xml:
            return response
        tree, lists = ElementTree.ElementTree(), {}
        tree.parse(StringIO(response))
        for elem in list(tree.getiterator('list')):
            lists[elem.attrib['name']] = elem.attrib['id']
        return lists

    def add_list(self, name):
        return self.post('audience_lists', name=name)

    def delete_list(self, name):
        return self.post('audience_lists/%s' % quote(name), _method='delete')

    def add_contacts(self, contacts_data, fields=DEFAULT_CONTACT_FIELDS):
        output = [fields]
        out = output.append
        for contact in contacts_data:
            line = []
            contact = contact + ['contactmanager']
            for field in contact:
                if '"' in field:
                    field = field.replace('"', '""')
                line.append('"%s"' % field)
            out(','.join(line))
        csv = '\n'.join(output)
        return self.post('audience_members', csv_file=csv)

    def add_contact(self, contact_data, fields=DEFAULT_CONTACT_FIELDS):
        return self.add_contacts([contact_data], fields)

    def subscribe(self, email, list):
        return self.post('audience_lists/%s/add' % quote(list), email=email)

    def unsubscribe(self, email, list):
        return self.post('audience_lists/%s/remove' % quote(list), email=email)

    def subscriptions(self, email):
        return self.get('audience_members/%s/lists.xml' % quote(email))

    # Unfortunately, accessing a CSV export of the list membership only works
    # through the web interface and doesn't work over the API, so you can't find
    # out who the subscribers for a given list are.
    def subscribers(self, list=None, list_id=None):
        if not list_id:
            list_id = self.lists(as_xml=False)[list]
        return self.get('exports/audience/%s.csv' % list_id)

# ------------------------------------------------------------------------------
# Set Up The Client
# ------------------------------------------------------------------------------

from secret import GOOGLE_USERNAME, GOOGLE_PASSWORD

account = Account()
account.access = ['tav@espians.com']
account.email = GOOGLE_USERNAME
account.hosted = True
account.password = GOOGLE_PASSWORD

client = ContactsService()
client.email = GOOGLE_USERNAME
client.password = GOOGLE_PASSWORD
client.source = 'ampify-contactmanager-0.1'
client.account_type = 'HOSTED'
client.contact_list = 'default'
client.ProgrammaticLogin()

# ------------------------------------------------------------------------------
# The Display View
# ------------------------------------------------------------------------------

def display(feed, o):

    for contact in feed.entry:
        o ('<h2>%s</h2>' % escape(contact.title.text))

# ------------------------------------------------------------------------------
# Service Decorator
# ------------------------------------------------------------------------------

# A service name -> function mapping.
SERVICES = {}

# This service decorator provides a super minimal web application framework on
# top of webapp. It works in conjunction with the ``MainHandler`` which does
# the actual dispatching.
def service(name, method='GET', kwargs=True, template=True, login=True):

    def __decorating_function(func):

        def __decorator(handler, o, *__args, **__kwargs):

            # We ensure that the user is authenticated if ``login=True``.
            if login:
                user = users.get_current_user()
                if not user:
                    handler.redirect(
                        users.create_login_url(handler.request.url)
                        )
                    return

            # We map the parameters to keyword arguments if ``kwargs=True``.
            if kwargs and not __kwargs:
                get = handler.request.get
                __kwargs = dict(
                    (str(arg), get(arg))
                    for arg in handler.request.arguments()
                    )

            # We wrap the response in a header and footer template if
            # ``template=True``.
            if template and not handler.templated:
                handler.templated = True
                o (TEMPLATE_HEADER)
                func(handler, o, *__args, **__kwargs)
                o (TEMPLATE_FOOTER)
                return

            func(handler, o, *__args, **__kwargs)

        # Match up the decorator's function name to make debugging easier.
        __decorator.__name__ = func.__name__

        SERVICES[(name, method)] = __decorator

        return __decorator

    return __decorating_function

# ------------------------------------------------------------------------------
# The Main Handler
# ------------------------------------------------------------------------------

# The MainHandler dispatches to the appropriate service functions.
class MainHandler(webapp.RequestHandler):

    def get(self):

        args = filter(None, self.request.path.split('/'))
        if args:
            service_name = args[0]
            args = args[1:]
        else:
            service_name = 'root'
            args = []

        self.templated = False

        service = SERVICES[(service_name, 'GET')]
        service(self, self.response.out.write, *args)

    def post(self):
        pass

# ------------------------------------------------------------------------------
# The Logout Handler
# ------------------------------------------------------------------------------

# This simple request handler will logout the current user and redirect to
# either a specified ``return_to`` url parameter or the site root.
@service('logout', login=False, kwargs=True)
def logout(handler, o, return_to='/'):
    handler.redirect(users.create_logout_url(return_to))

# ------------------------------------------------------------------------------
# The Accounts Handler
# ------------------------------------------------------------------------------

class AccountsHandler(webapp.RequestHandler):

    def get(self, o):
        owner = str(users.get_current_user())
        method = None
        id = self.request.path.split('/')[2:]
        if len(id) == 2:
            id, method = id
        elif id:
            id = id[0]
        if id == 'create':
            o ("<h2>Create A New Account</h2>")
        elif id:
            key = db.Key.from_path('A', id)
            account = db.get(key)
            if account:
                if account.owner == owner:
                    self.render_account(o, account)
                else:
                    o ("Sorry you're not the owner of account %r." % escape(id))
            else:
                o ("Couldn't find account %r." % escape(id))
        else:
            o ('<h1>Accounts</h1>')
            accounts = Account.all().filter('o =', owner).fetch(100)
            for account in accounts:
                self.render_account(o, account)

    def render_account(self, o, account):
        o ('<h2>%s</h2>' % account.email)

# ------------------------------------------------------------------------------
# The Import Handler
# ------------------------------------------------------------------------------

class ImportHandler(webapp.RequestHandler):

    def get(self):
        Account.all().filter('imported =', False).fetch(100)

@service('root')
def home(handler, o, *args, **kwargs):

    o('<h3>%r</h3>' % repr(args))
    o('<h3>%r</h3>' % kwargs)

    query = ContactsQuery(client.GetFeedUri())
    query.orderby = 'lastmodified'
    query['sortorder'] = 'descending'

    feed = client.GetContactsFeed(query.ToUri())

    display(feed, o)

@service('bar', kwargs=True)
def bar(handler, o, *args, **kwargs):
    o('<h3>%s</h3>' % repr(args))
    o('<h3>%r</h3>' % kwargs)

@service('foo', kwargs=True)
def home(handler, o, *args, **kwargs):

    o('<h3>%s</h3>' % repr(args))
    o('<h3>%r</h3>' % kwargs)

    bar(handler, o, 'foo', **kwargs)

# ------------------------------------------------------------------------------
# The Main Function
# ------------------------------------------------------------------------------

def main():
    application = webapp.WSGIApplication([
        ('.*', MainHandler),
        ], debug=True)
    run_wsgi_app(application)


if __name__ == '__main__':
    main()


# r -> it looks for everything in the Name field starting with r
# #joblabs -> looks for every contact with #jobslabs in the notes
# "something" -> looks for every contact with something in the notes
# foo@blah.com -> looks for the contact with that in the Email field
# @joblabs -> looks for every contact matching the @joblabs audience

# @joblabs = #joblabs
# @some-event = #joblabs OR #socbizmod
# @lcl = #joblabs OR #socbizmod OR (#supporters #family)

# the @lists will get synced to madmimi audiences


#l = List()
#l.query = "#family OR #supporters"
#l.refs = ['#family', '#supporters']

# twitter
# facebook
# editing