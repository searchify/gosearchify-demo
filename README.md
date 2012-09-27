gosearchify-demo
================

A demo webapp in Go that uses Searchify's Go IndexTank client

To get started with Go on Heroku, read Mark McGranaghan's
[Getting Started with Go on Heroku](http://mmcgrana.github.com/2012/09/getting-started-with-go-on-heroku.html).

Then, clone the gosearchify-demo from github:

```bash
    $ git clone git://github.com/searchify/gosearchify-demo.git
    $ cd gosearchify-demo
```

Create an app on Heroku:

```bash
    $ heroku create -b https://github.com/kr/heroku-buildpack-go.git
```

Note: you'll need to add the Searchify add-on

Push to Heroku:

```bash
    $ git push heroku master
```

Try it out:

```bash
    $ heroku open
```

If it isn't working, check the logs:

```bash
    $ heroku logs
```


## Thanks

Thanks to [Keith Rarick](http://xph.us/), [Blake Mizerany](https://github.com/bmizerany),
[Ryan Smith](http://ryandotsmith.heroku.com/),
[Mark McGranaghan](http://mmcgrana.github.com/2012/09/getting-started-with-go-on-heroku.html) and
the IndexTank team.
