Hookworm Heroku Example
=======================

This is an example of how to deploy a hookworm "instance" to Heroku or a
Heroku work-alike such as dokku or Flynn.  Steps to verify, which are
roughly the same as any other Heroku thing:

``` bash
git clone https://github.com/modcloth-labs/hookworm-heroku-example.git
cd hookworm-heroku-example
heroku create -b https://github.com/ddollar/heroku-buildpack-multi.git
# ...
git push heroku master
```
