from pyutil.pylzf import compress, decompress

HEADERS = """

GET / HTTP/1.1
host: www.yahoo.com
connection: close
user-agent: Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_5_8; en-us) AppleWebKit/531.21.8 (KHTML, like Gecko) Version/4.0.4 Safari/531.21.10
accept-encoding: gzip
accept-charset: ISO-8859-1,UTF-8;q=0.7,*;q=0.7
cache-control: no
accept-language: en,de;q=0.7,en-us;q=0.3
referer: http://www.google.com

""".strip()

PRESET = "".join("""

OPTIONSGETHEADPOSTPUTDELETETRACEacceptaccept-charsetaccept-encodingaccept-
languageauthorizationexpectfromhostif-modified-sinceif-matchif-none-matchi
f-rangeif-unmodifiedsincemax-forwardsproxy-authorizationrangerefererteuser
-agent10010120020120220320420520630030130230330430530630740040140240340440
5406407408409410411412413414415416417500501502503504505accept-rangesageeta
glocationproxy-authenticatepublicretry-afterservervarywarningwww-authentic
ateallowcontent-basecontent-encodingcache-controlconnectiondatetrailertran
sfer-encodingupgradeviawarningcontent-languagecontent-lengthcontent-locati
oncontent-md5content-rangecontent-typeetagexpireslast-modifiedset-cookieMo
ndayTuesdayWednesdayThursdayFridaySaturdaySundayJanFebMarAprMayJunJulAugSe
pOctNovDecchunkedtext/htmlimage/pngimage/jpgimage/gifapplication/xmlapplic
ation/xhtmltext/plainpublicmax-agecharset=iso-8859-1utf-8gzipdeflateHTTP/1
.1statusversionurlMozillaMacintoshIntelMacGeckoKHTMLVersionAppleWebKitMac
10_5_8OS XSafarigzipen

""".strip().split())

# putting "en" at the end seems to improve compression ratio??

LEN_PRESET = len(PRESET)

COMPRESSED_PRESET = compress(PRESET)[4:]

def compress_with_preset(data):
    initial = compress(PRESET + data)
    pos = 0
    for i, j in zip(initial[4:], COMPRESSED_PRESET):
        if i != j:
            break
        pos += 1
    return pos, initial[:4], initial[pos+4:]

def decompress_with_preset(pos, size, data):
    return decompress(size + COMPRESSED_PRESET[:pos] + data)[LEN_PRESET:]

pos, size, output = compress_with_preset(HEADERS)
input = decompress_with_preset(pos, size, output)

print "Raw Headers: ", len(HEADERS), "bytes"
print "Compressed Headers: ", len(compress(HEADERS)), "bytes"

print "Compressed Headers with Preset:", len(size + output + str(pos)), "bytes"
print "Savings:", len(HEADERS) - len(size + output + str(pos)), "bytes"
