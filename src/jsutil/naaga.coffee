# No Copyright (-) 2010 The Ampify Authors. This file is under the
# Public Domain license that can be found in the root LICENSE file.

# Naaga
# =====

fs = require('fs')
sys = require('sys')
ucd = require('./ucd')

Categories = ucd.Categories
CategoryRanges = ucd.CategoryRanges
NumberOfCategoryRanges = CategoryRanges.length
CategoryNames = ucd.CategoryNames

Cc = ucd.Cc
Cf = ucd.Cf
Co = ucd.Co
Cs = ucd.Cs
Ll = ucd.Ll
Lm = ucd.Lm
Lo = ucd.Lo
Lt = ucd.Lt
Lu = ucd.Lu
Mc = ucd.Mc
Me = ucd.Me
Mn = ucd.Mn
Nd = ucd.Nd
Nl = ucd.Nl
No = ucd.No
Pc = ucd.Pc
Pd = ucd.Pd
Pe = ucd.Pe
Pf = ucd.Pf
Pi = ucd.Pi
Po = ucd.Po
Ps = ucd.Ps
Sc = ucd.Sc
Sk = ucd.Sk
Sm = ucd.Sm
So = ucd.So
Zl = ucd.Zl
Zp = ucd.Zp
Zs = ucd.Zs

CasedLetters = ucd.CasedLetters
Letters = ucd.Letters
Marks = ucd.Marks
Numbers = ucd.Numbers
Others = ucd.Others
Punctuations = ucd.Punctuations
Separators = ucd.Separators
Symbols = ucd.Symbols
Unknown = ucd.Unknown

Ascii = 0x80
HighSurrogateStart = 0xD800
HighSurrogateEnd = 0xD8FF
LowSurrogateStart = 0xDC00
LowSurrogateEnd = 0xDFFF
MaxCodepoint = 0x10FFFF
ReplacementChar = 0xFFFD

# Tokenisation
# ------------
#
# The ``tokenise`` function uses a novel approach put forward by sbp (Sean B.
# Palmer). It splits the given ``source`` into tokens of distinct categories
# as specified in Unicode 5.2.
#
# It returns a sequence of ``[category, ascii_flag, codepoints, string]``
# tokens:
#
# * The ``category`` points to a constant representing the unicode category.
#
# * The ``ascii_flag`` indicates whether the segment is comprised of only ASCII
#   characters.
#
# * The ``codepoints`` are the unicode codepoints for use by any normalisation
#   function.
#
# * The ``string`` is the segment of the source for the current token.
tokenise = (source) ->

    tokens = []
    idxAscii = true
    idxCategory = -1
    idxList = []
    idxStart = 0
    HighSurrogate = 0

    for i in [0...source.length]

        # Get the UTF-16 code unit and convert it to a Unicode Codepoint.
        codepoint = source.charCodeAt i
        ascii = false

        # Deal with surrogate pairs.
        if HighSurrogate isnt 0
            if LowSurrogateStart <= codepoint <= LowSurrogateEnd
                codepoint = (HighSurrogate - HighSurrogateStart) *
                            1024 + 65536 +
                            (codepoint - LowSurrogateStart)
            else
                # Replace a malformed surrogate pair with the Unicode
                # Replacement Character.
                codepoint = ReplacementChar
            HighSurrogate = 0
        else if HighSurrogateStart <= codepoint <= HighSurrogateEnd
            HighSurrogate = codepoint
            continue
        # Handle the common case of ASCII before doing a check for the worst
        # case.
        else if codepoint < Ascii
            ascii = true
        # Do a sanity check to ensure that the codepoint is not outside the
        # Unicode Character Set. If so, replace it with the Unicode Replacement
        # Character.
        else if codepoint >= MaxCodepoint
            codepoint = ReplacementChar

        category = Categories[codepoint]

        if typeof category is "undefined"
            for range in CategoryRanges
                if range[0] <= codepoint <= range[1]
                    category = range[2]
                    break
            if typeof category is "undefined"
                category = Unknown

        if category is idxCategory
            idxAscii = idxAscii and ascii
            idxList.push codepoint
        else
            tokens.push [idxCategory, idxCategory, idxList, source[idxStart...i]]
            idxAscii = ascii
            idxCategory = category
            idxList = [codepoint]
            idxStart = i

    tokens.push [idxCategory, idxCategory, idxList, source[idxStart...i]]
    return tokens

parse = ->
    hasProp = Object.hasOwnProperty

evaluate = ->

compile = ->

printTokens = (tokens) ->
  for token in tokens
    sys.puts(JSON.stringify([CategoryNames[token[0]], token[1], token[3]]))
  return

# // Utility Functions
# // -----------------
# //
# function timeit(n, func) {
#     var args,
#         start,
#         stop,
#         i;
#     args = Array.prototype.slice.call(arguments, 2, arguments.length);
#     start = new Date().getTime();
#     for (i = 0; i < n; i++) {
#         func.apply(func, args);
#     }
#     stop = new Date().getTime() - start;
#     sys.puts(stop);
#     return stop;
# }

# function bench(duration) {
#     var args,
#         func,
#         i = 0,
#         name = "unknown function",
#         start,
#         stop,
#         total = 0;
#     args = Array.prototype.slice.call(arguments, 1, arguments.length);
#     if (typeof duration === 'string') {
#         name = duration;
#         duration = args[0];
#         args = args.slice(1, args.length);
#     }
#     if (typeof duration === 'number') {
#         func = args.shift();
#     } else {
#         func = duration;
#         duration = 1;
#     }
#     sys.puts("benching <" + name + "> for " + duration + "s:\n");
#     duration *= 1000;
#     while (true) {
#         start = new Date().getTime();
#         func.apply(func, args);
#         total = total + (new Date().getTime() - start);
#         if (total > duration) {
#             break;
#         }
#         i++;
#     }
#     sys.puts("    " + (total / 1000) + "\t" + i + " runs\n");
#     return total;
# }

root = exports ? this
root.tokenise = tokenise
