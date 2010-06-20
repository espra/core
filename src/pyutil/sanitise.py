# No Copyright (-) 2008-2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

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
    'git|gtalk|im|irc|ircs|itms|lastfm|magnet|mailto|maps|md5|msnim|news|nntp|'
    'psyc|rsync|rtsp|secondlife|sftp|sha|sip|sips|skype|sms|spotify|ssh|svn|tag|'
    'tel|urn|uuid|webcal|xmpp|xri|ymsgr):.*', re.IGNORECASE
    ).match

match_valid_css_value = re.compile(
    '^(#[0-9a-f]+|rgb\(\d+%?,\d*%?,?\d*%?\)?|\d{0,2}\.?\d{0,2}'
    '(cm|deg|em|ex|in|mm|pc|pt|px|%|,|\))?)$'
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
    border-bottom-style border-bottom-width border-collapse border-color
    border-left border-left-color border-left-style border-left-width
    border-right border-right-color border-right-style border-right-width
    border-spacing border-style border-top border-top-color border-top-style
    border-top-width border-width caption-side clear color cursor direction
    display elevation empty-cells float font font-family font-size font-style
    font-variant font-weight height letter-spacing line-height overflow padding
    padding-bottom padding-left padding-right padding-top pause pause-after
    pause-before pitch pitch-range richness speak speak-header speak-numeral
    speak-punctuation speech-rate stress text-align text-decoration text-indent
    unicode-bidi vertical-align voice-family volume white-space width

    """)

# clip content counter counter-increment cue* filter

VALID_CSS_KEYWORDS = create_set("""

    left-side far-left left center-left center center-right right
    far-right right-side leftwards rightwards

    below level above lower higher

    alias all-scroll cell col-resize copy count-down count-up count-up-down
    crosshair default grab grabbing hand help move no-drop not-allowed pointer
    progress spinning text vertical-text wait

    !important aqua auto black block blue bold both bottom brown center collapse
    compact dashed dotted double embed fuchsia gray green groove hide hidden
    inherit inline inline-block inset italic left lime list-item ltr marker
    maroon medium navy none normal nowrap olive outset pointer purple red ridge
    right rtl run-in separate show silver solid table table-cell table-column
    table-column-group table-footer-group table-header-group table-row
    table-row-group teal top transparent underline white yellow

    """)

# *-resize (for cursor)

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
    strip_pi=True, rel_whitelist=None, second_run=True
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
                continue
            if attr == 'id':
                if not val.startswith(secure_id_prefix):
                    continue
            elif attr == 'style':
                new_style = []; add_style = new_style.append
                for style in val.split(';'):
                    style = style.strip().split(':', 1)
                    if len(style) != 2:
                        continue
                    prop, value = style
                    prop = prop.strip()
                    if prop not in valid_css_properties:
                        continue
                    value = value.strip()
                    valid = True
                    for part in value.split():
                        if part not in valid_css_keywords:
                            if not match_valid_css_value(part):
                                valid = False
                    if not valid:
                        continue
                    add_style("%s: %s;" % (prop, value))
                if not new_style:
                    continue
                val = ''.join(new_style)
            elif attr == 'class':
                val = ' '.join(
                    klass for klass in val.split()
                    if klass in valid_css_classes
                    )
                if not val:
                    continue
            elif attr in attrs_with_uri_refs:
                if not match_valid_uri_scheme(val):
                    continue
                elif match_js_in_uri_ref(val):
                    continue
            elif (attr == 'rel') and rel_whitelist:
                if val.split(':')[0] not in rel_whitelist:
                    continue
            append((attr, val))

        tag.attrs = tag_attrs

    # protects against: <<SCRIPT>script>("XSS");//<</SCRIPT>

    if not second_run:
        return sanitise(
            soup.renderContents(), valid_tags, valid_attrs, valid_attr_prefixes,
            attrs_with_uri_refs, valid_css_properties, valid_css_keywords,
            valid_css_classes, secure_id_prefix, strip_cdata, strip_comments,
            strip_pi, rel_whitelist, True
            )

    return soup.renderContents().replace('<script>', '').replace('<script ', '')
