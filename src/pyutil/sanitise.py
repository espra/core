# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

r"""
===============
HTML5 Sanitiser
===============

This module provides a single ``sanitise`` function which should protect you
from a wide range of XSS attacks:

  >>> sanitise('<SCRIPT SRC=http://ha.ckers.org/xss.js></SCRIPT>')
  ''

  >>> sanitise("'';!--\"<XSS>=&{()}") == "'';!--\"=&{()}"
  True

  >>> sanitise("<IMG SRC=JaVaScRiPt:alert('XSS')>")
  '<img />'

  >>> sanitise('<INPUT TYPE="image" SRC="javascript:alert(\'XSS\');">')
  '<input type="image" />'

  >>> sanitise('<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="http://ha.ckers.org/xss.js"></SCRIPT>')
  'document.write("'

  >>> sanitise('<SCRIPT a=">\'>" SRC="http://ha.ckers.org/xss.js"></SCRIPT>')
  '\'>" SRC="http://ha.ckers.org/xss.js">'

  >>> sanitise('<<SCRIPT>alert("XSS");//<</SCRIPT>')
  ''

  >>> sanitise('<SCRIPT/SRC="http://ha.ckers.org/xss.js"></SCRIPT>')
  'SRC="http:/ha.ckers.org/xss.js">'

""" # emacs "'

# See http://ha.ckers.org/xss.html for a listing of various XSS attacks

import re

from BeautifulSoup import BeautifulSoup, CData, Comment, ProcessingInstruction

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

create_set = lambda values: frozenset(values.strip().split())
find_cdata = lambda text: isinstance(text, CData)
find_comments = lambda text: isinstance(text, Comment)
find_pi = lambda text: isinstance(text, ProcessingInstruction)

match_js_in_uri_ref = re.compile(
    '(%s)*'
    'j[\s]*(&#x.{1,7})?'
    'a[\s]*(&#x.{1,7})?'
    'v[\s]*(&#x.{1,7})?'
    'a[\s]*(&#x.{1,7})?'
    's[\s]*(&#x.{1,7})?'
    'c[\s]*(&#x.{1,7})?'
    'r[\s]*(&#x.{1,7})?'
    'i[\s]*(&#x.{1,7})?'
    'p[\s]*(&#x.{1,7})?t' % [chr(i) for i in range(33)], re.IGNORECASE
    ).match

match_valid_uri_scheme = re.compile(
    '(http|https|aim|amp|callto|data|dict|dns|fax|fb|feed|file|freenet|ftp|geo|'
    'git|gtalk|im|irc|ircs|itms|lastfm|magnet|mailto|maps|md5|mms|msnim|news|'
    'nntp|psyc|rsync|rtsp|secondlife|sftp|sha|sip|sips|skype|sms|spotify|ssh|'
    'svn|tag|tel|urn|uuid|webcal|xmpp|xri|ymsgr):.*', re.IGNORECASE
    ).match

match_valid_css_value = re.compile(
    '^(#[0-9a-f]+|rgb\(\d+%?,\d*%?,?\d*%?\)?|\d{0,2}\.?\d{0,2}'
    '(cm|deg|em|ex|hz|in|mm|pc|pt|px|s|%|,|\))?)$'
    ).match

# ------------------------------------------------------------------------------
# Some Constants
# ------------------------------------------------------------------------------

VALID_TAGS = create_set("""
    a abbr acronym address area article aside audio b blockquote br button
    canvas caption cite code col colgroup command datalist dd del details dfn
    dialog div dl dt em fieldset figcaption figure footer form h1 h2 h3 h4 h5 h6
    header hgroup hr i img input ins kbd keygen label legend li map mark menu
    meter nav noscript ol optgroup option output p pre progress q rp rt ruby
    samp section select small source span strong sub summary sup table tbody td
    textarea tfoot th thead time tr track ul var video wbr
    """)

# base bdo big body center datagrid dir embed event-source font head html iframe
# link meta object param s script strike style title tt u

VALID_ATTR_PREFIXES = ('aria-', 'data-')

VALID_ATTRS = create_set("""

    class contenteditable contextmenu dir id itemid itemprop itemref itemscope
    itemtype hidden lang role style spellcheck title

    accept accept-charset action alt autocomplete challenge checked cite cols
    colspan coords datetime disabled enctype for form formaction formenctype
    formmethod formnovalidate headers height high href hreflang icon ismap
    keytype kind label list low max maxlength media method min multiple name
    novalidate open optimum pattern ping placeholder preload poster pubdate
    radiogroup readonly rel required reversed rows rowspan scope selected shape
    size span src srclang start step summary type usemap value width wrap

    align border cellpadding cellspacing noshade nowrap

    """)

# accesskey draggable tabindex

# autofocus autoplay controls formtarget loop target

# axis background balance bgcolor bgproperties bordercolor bordercolordark
# bordercolorlight bottompadding ch char charoff charset choff clear color
# compact data datafld datapagesize datasrc default delay dynsrc end face frame
# galleryimg gutter hidefocus hspace inputmode leftspacing longdesc loopcount
# loopend loopstart lowsrc nohref point-size pqg prompt repeat-max repeat-min
# replace rev rightspacing rules scope suppress template toppadding unselectable
# urn valign variable volume vrml vspace

ATTRS_WITH_URI_REFS = frozenset(['action', 'cite', 'formaction', 'href', 'src'])

# background longdesc

VALID_CSS_PROPERTIES = create_set("""

    azimuth background-color border border-bottom border-bottom-color
    border-bottom-left-radius border-bottom-right-radius border-bottom-style
    border-bottom-width border-collapse border-color border-left
    border-left-color border-left-style border-left-width border-radius
    border-right border-right-color border-right-style border-right-width
    border-spacing border-style border-top border-top-color
    border-top-left-radius border-top-right-radius border-top-style
    border-top-width border-width caption-side clear color cursor direction
    display elevation empty-cells float font font-family font-size
    font-size-adjust font-stretch font-style font-variant font-weight height
    ime-mode layout-flow layout-grid layout-grid-char layout-grid-char-spacing
    layout-grid-line layout-grid-mode layout-grid-type letter-spacing
    line-height line-break list-style list-style-position list-style-type marks
    max-height max-width min-height min-width outline-color outline-style
    outline-width overflow overflow-x overflow-y padding padding-bottom
    padding-left padding-right padding-top pause pause-after pause-before pitch
    pitch-range richness ruby-align ruby-overhang ruby-position speak
    speak-header speak-numeral speak-punctuation speech-rate stress table-layout
    text-align text-align-last text-autospace text-decoration text-justify
    text-kashida-space text-overflow text-shadow text-transform unicode-bidi
    vertical-align voice-family volume white-space width word-break word-spacing
    word-wrap writing-mode

    """)

# background* box-shadow* clip column* content counter counter-increment cue*
# filter include-source layer* left/right/top/bottom margin list-style-image
# marker-offset orphans page* play-during position quotes scrollbar* size
# text-indent text-underline-position widows
# z-index zoom

VALID_CSS_KEYWORDS = create_set("""

    left-side far-left left center-left center center-right right
    far-right right-side leftwards rightwards

    below level above lower higher

    alias all-scroll cell col-resize copy count-down count-up count-up-down
    crosshair default grab grabbing hand help move no-drop not-allowed pointer
    progress spinning text vertical-text wait

    caption icon menu message-box small-caption status-bar

    large larger medium small smaller x-large x-small xx-large xx-small

    condensed expanded extra-condensed extra-expanded narrower normal
    semi-condensed semi-expanded ultra-condensed ultra-expanded wider

    active inactive deactivated

    horizontal vertical-ideographic

    both char fixed line loose strict

    inside outside

    armenian circle cjk-ideographic decimal decimal-leading-zero disc georgian
    hebrew katakana-iroha hiragana hiragana-iroha katakana lower-alpha
    lower-greek lower-latin lower-roman square upper-alpha upper-latin
    upper-roman

    crop cross

    medium thick thin

    hidden scroll visible

    high low medium x-high x-low

    above center distribute-letter distribute-space inline left line-edge right
    whitespace

    always code continuous digits once spell-out

    fast medium slow x-fast x-slow

    ideograph-alpha ideograph-numeric ideograph-parenthesis ideograph-space

    distribute distribute-all-lines inter-cluster inter-ideograph inter-word
    newspaper

    clip ellipsis

    capitalize lowercase uppercase

    bidi-override embed

    baseline bottom middle sub super text-bottom text-top top

    collapse hidden hide show visible

    male female child

    loud medium silent soft x-loud x-soft

    break-all break-word keep-all

    lr-tb tb-rl

    arial arial-black comic-sans-ms constantina consolas courier courier-new
    cursive fantasy franklin-gothic-medium geneva georgia helvetica
    helvetica-neue impact lucida-console lucida-grande lucida-sans-unicode
    monaco monospace palatino-linotype sans-serif serif tahoma times
    times-new-roman trebuchet-ms verdana

    !important aqua auto black block blue bold bolder both bottom brown center
    collapse compact dashed dotted double embed fixed fuchsia gray green groove
    hide hidden inherit inline inline-block inset invert italic justify left
    lighter lime line-through list-item ltr marker maroon medium navy none
    normal nowrap oblique olive outset overline pointer pre purple red ridge
    right rtl run-in separate show silver small-caps solid strict table
    table-cell table-column table-column-group table-footer-group
    table-header-group table-row table-row-group teal top transparent underline
    white yellow

    """)

# *-resize (for cursor)
# font-family/size
# text-decoration: blink
# voice-family

VALID_CSS_CLASSES = frozenset()

# ------------------------------------------------------------------------------
# Core Function
# ------------------------------------------------------------------------------

def sanitise(
    html, valid_tags=VALID_TAGS, valid_attrs=VALID_ATTRS,
    valid_attr_prefixes=VALID_ATTR_PREFIXES,
    attrs_with_uri_refs=ATTRS_WITH_URI_REFS,
    valid_css_properties=VALID_CSS_PROPERTIES,
    valid_css_keywords=VALID_CSS_KEYWORDS,
    valid_css_classes=VALID_CSS_CLASSES,
    secure_id_prefix='local-', strip_cdata=True, strip_comments=True,
    strip_pi=True, rel_whitelist=None, second_run=False,
    allow_relative_urls=False, encoding='utf-8'
    ):
    """Return a sanitised version of the provided HTML."""

    soup = BeautifulSoup(html)

    if strip_cdata:
        for cdata in soup.findAll(text=find_cdata):
            cdata.extract()

    if strip_comments:
        for comment in soup.findAll(text=find_comments):
            comment.extract()

    if strip_pi:
        for pi in soup.findAll(text=find_pi):
            pi.extract()

    for tag in soup.findAll(True):

        if tag.name not in valid_tags:
            tag.hidden = True
            continue

        tag_attrs = []; append = tag_attrs.append

        for attr, val in tag.attrs:
            if attr not in valid_attrs:
                for prefix in valid_attr_prefixes:
                    if attr.startswith(prefix):
                        append((attr, val))
                        break
                continue
            if attr == 'id':
                if not val.startswith(secure_id_prefix):
                    continue
            elif attr == 'style':
                new_style = []; add_style = new_style.append
                for style in val.split(u';'):
                    style = style.strip().split(u':', 1)
                    if len(style) != 2:
                        continue
                    prop, value = style
                    prop = prop.strip().lower()
                    if prop not in valid_css_properties:
                        continue
                    components = []; add_component = components.append
                    segments = filter(None, value.split(u','))
                    segcount = len(segments) - 1
                    for idx, segment in enumerate(segments):
                        current = None
                        in_quotes = 0
                        for char in segment:
                            if current is None:
                                if char.isspace():
                                    continue
                                if char == '"' or char == "'":
                                    in_quotes = 1
                                    current = u''
                                else:
                                    current = char
                                continue
                            if in_quotes:
                                if char == '"' or char == "'":
                                    in_quotes = 0
                                    add_component(current)
                                    current = None
                                else:
                                    current += char
                            else:
                                if char.isspace():
                                    add_component(current)
                                    current = None
                                else:
                                    current += char
                        if current:
                            add_component(current)
                        if idx != segcount:
                            add_component(u',')
                    valid = True
                    new_value = []; add_part = new_value.append
                    for component in components:
                        if component == u',':
                            add_part(u', ')
                            continue
                        norm = u'-'.join(component.lower().split())
                        if norm not in valid_css_keywords:
                            if not match_valid_css_value(norm):
                                valid = False
                                break
                        if u' ' in component:
                            add_part(u' "%s"' % component)
                        else:
                            add_part(u' %s' % component)
                    if not valid:
                        continue
                    add_style(u"%s: %s;" % (prop, ''.join(new_value)))
                if not new_style:
                    continue
                val = u''.join(new_style)
            elif attr == 'class':
                val = u' '.join(
                    klass for klass in val.split()
                    if klass in valid_css_classes
                    )
                if not val:
                    continue
            elif attr in attrs_with_uri_refs:
                if allow_relative_urls:
                    if match_js_in_uri_ref(val):
                        continue
                elif not match_valid_uri_scheme(val):
                    continue
            elif (attr == 'rel') and rel_whitelist:
                if val.split(u':')[0] not in rel_whitelist:
                    continue
            append((attr, val))

        tag.attrs = tag_attrs

    # protects against: <<SCRIPT>script>("XSS");//<</SCRIPT>

    if not second_run:
        return sanitise(
            unicode(soup.renderContents(), encoding), valid_tags, valid_attrs,
            valid_attr_prefixes, attrs_with_uri_refs, valid_css_properties,
            valid_css_keywords, valid_css_classes, secure_id_prefix,
            strip_cdata, strip_comments, strip_pi, rel_whitelist, True,
            allow_relative_urls, encoding
            )

    return unicode(soup.renderContents(), encoding)

# ------------------------------------------------------------------------------
# Run Tests
# ------------------------------------------------------------------------------

if __name__ == '__main__':

    import doctest
    doctest.testmod()
