gosearchify-demo
================

A demo webapp in Go that uses Searchify's Go IndexTank client to search through the Enron email archive.

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
    Creating murmuring-woodland-6954... done, stack is cedar
    BUILDPACK_URL=https://github.com/kr/heroku-buildpack-go.git
    http://murmuring-woodland-6954.herokuapp.com/ | git@heroku.com:murmuring-woodland-6954.git
    Git remote heroku added
```

Note: you'll need to add the Searchify add-on, which will incur charges (the small plan is less than $1/day), or
sign up for a free trial account at http://www.searchify.com.  To add the Heroku add-on:

```bash
    $ heroku addons:add searchify:small
    Adding searchify:small on murmuring-woodland-6954... done, v5 ($25/mo)
    Use `heroku addons:docs searchify:small` to view documentation.
```

Push to Heroku:

```bash
    $ git push heroku master
    Counting objects: 44, done.
    Delta compression using up to 2 threads.
    Compressing objects: 100% (30/30), done.
    Writing objects: 100% (44/44), 91.50 KiB, done.
    Total 44 (delta 10), reused 44 (delta 10)

    -----> Heroku receiving push
    -----> Fetching custom git buildpack... done
    -----> Go app detected
    -----> Installing Go 1.0.3... done
           Installing Virtualenv... done
           Installing Mercurial... done
           Installing Bazaar... done
    -----> Running: go get ./...
    -----> Discovering process types
           Procfile declares types -> web
    -----> Compiled slug size: 1.3MB
    -----> Launching... done, v6
           http://murmuring-woodland-6954.herokuapp.com deployed to Heroku

    To git@heroku.com:murmuring-woodland-6954.git
     * [new branch]      master -> master
```

Try it out:

```bash
    $ heroku open
```

If it isn't working, check the logs:

```bash
    $ heroku logs
```

If you get an "Index does not exist" error, the most likely cause is that you haven't 
created a Searchify index called "enron".


## Thanks

Thanks to [Keith Rarick](http://xph.us/), [Blake Mizerany](https://github.com/bmizerany),
[Ryan Smith](http://ryandotsmith.heroku.com/),
[Mark McGranaghan](http://mmcgrana.github.com/2012/09/getting-started-with-go-on-heroku.html) and
the IndexTank team.
