---
license: Public Domain
layout: page
title: jsutil.builtins
---

The module provides some utility functions and extensions of builtin objects to
complement the standard Javascript environment.

The `Object.extend` method allows you to easily update an object with the
contents of another:

    > a = {'a': 1}

    > b = {'b': 2}

    > a.extend(b)
    {'a': 1, 'b': 2}

    > foo = 2

    > bar = 3

    > foo + bar
    5
