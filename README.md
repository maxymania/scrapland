# scrapland
A (kind-of) web framework / web ducttape / toolkit, that allows you to glue together fine websites.

Scrapland is a collection of loosly coupled packages, that fullfill different jobs such as Web Scraping,
accessing FastCGI servers such as php-fpm, compose multiple `http.Handler`s together into one.

# Modules:

## Endorsed Public API Packages

- [container](https://godoc.org/github.com/maxymania/scrapland/container)
- [fcgibinding](https://godoc.org/github.com/maxymania/scrapland/fcgibinding)
- [override](https://godoc.org/github.com/maxymania/scrapland/override)
- [tmplhelp](https://godoc.org/github.com/maxymania/scrapland/tmplhelp)
- [webscrape](https://godoc.org/github.com/maxymania/scrapland/webscrape)

## Internally used Low(er) Level Packages

- [fcgiclient](https://godoc.org/github.com/maxymania/scrapland/fcgiclient)
- [htmlscrape](https://godoc.org/github.com/maxymania/scrapland/htmlscrape)

## Obsolete/Unstable/Useless Packages (may varnish)

- portlet
