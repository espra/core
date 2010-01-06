#! /bin/sh

# Released into the Public Domain by tav <tav@espians.com>

echo
echo "# Todo:"
grep "✗" TODO.rst | wc -l

echo
echo "# Done:"
grep "✓" TODO.rst | wc -l

echo
echo "# Total:"
grep "[✓✗]" TODO.rst | wc -l

echo