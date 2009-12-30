#! /bin/sh

# Released into the Public Domain by tav <tav@espians.com>

echo
echo "# Todo:"
grep "✗" TODO.txt | wc -l

echo
echo "# Done:"
grep "✓" TODO.txt | wc -l

echo
echo "# Total:"
grep "[✓✗]" TODO.txt | wc -l

echo