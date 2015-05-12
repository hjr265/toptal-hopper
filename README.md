# Hopper

Hopper is a prototype web application deployment automation tool built to accompany an article on Toptal Blog.

## Usage

To compile this, you need the [Go distribution installed](http://golang.org/doc/install) on your computer. You can download and compile this program as you would [do for any simple Go program](http://golang.org/doc/code.html#Command).

~~~
mkdir hopper
cd hopper
export GOPATH=`pwd`
go get github.com/hjr265/toptal-hopper
go install github.com/hjr265/toptal-hopper
~~~

To run this on your server, make sure Git and Bash is installed and a valid configuration file (config.tml) exists in the current directory.

~~~
# config.tml

[core]
addr = ":26590"

[buildpack]
url = "https://github.com/heroku/heroku-buildpack-nodejs.git"

[app]
repo = "hjr265/hopper-hello.js" # github_username/repository_name

	[app.env]
	GREETING = "Hello"

	[app.procs]
	web = 1

[hook]
secret = ""
~~~

~~~
./hopper
~~~
