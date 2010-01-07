# Released into the Public Domain by tav <tav@espians.com>

"""Main entrypoint."""

from ampify.weblite import run_wsgi_app, Application, DEBUG

# ------------------------------------------------------------------------------
# self runner -- app engine cached main() function
# ------------------------------------------------------------------------------

if DEBUG == 2:

    import logging

    from cProfile import Profile
    from pstats import Stats
    from StringIO import StringIO

    def runner():
        run_wsgi_app(Application)

    def main():
        """Profiling main function."""

        profiler = Profile()
        profiler = profiler.runctx("runner()", globals(), locals())
        iostream = StringIO()

        stats = Stats(profiler, stream=iostream)
        stats.sort_stats("time")  # or cumulative
        stats.print_stats(80)     # 80 == how many to print

        # optional:
        # stats.print_callees()
        # stats.print_callers()

        logging.info("Profile data:\n%s", iostream.getvalue())

else:

    def main():
        """Default main function."""

        run_wsgi_app(Application)

# ------------------------------------------------------------------------------
# run in standalone mode
# ------------------------------------------------------------------------------

if __name__ == '__main__':
    main()
