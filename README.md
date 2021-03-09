# MCmanager [![CircleCI](https://circleci.com/gh/jlmeeker/mcmanager.svg?style=svg)](https://circleci.com/gh/jlmeeker/mcmanager)

## Status

Even though MCmanager is used by myself, it is still in its early stage of development.  You are encouraged to use it and submit issues for bug fixes and feature requests.

Note that we will try our best to avoid any backward incompatibility.  If we do run into something, it will be documented.

## About

MCmanager is a web front-end for managing any number of locally running Minecraft servers. The interface is designed with both mobile and desktop in mind. It lets you...

* Self-contained, single binary.  Just build/install and run (no need to place HTML template files anywyere... that isn't supported right now anyway)
* Running Minecraft server instances aren't attached to the mcmanager process, so restarting mcmanager (should) be fine and not kill any running servers.
* Log in with your Minecraft account to manage your servers.
* Create, start, stop, and even delete server instances (vanilla only, for now).
* See which users are OPs and who is playing on each server.
* See the latest minecraft.net news.
* Tested on Linux so YMMV with other OS's
* Uses rcon to communicate with the running instances
* ... more features coming soon

## Configuration

No configuration is currently possible, but it will be a necessity before long.  See [Todo](#todo)

```
TBD
```

## Installation

Ensure you have Go >= 1.16.0 installed and set up on your machine, then run the following command:

```
$ go install github.com/jlmeeker/mcmanager@latest 
```

## CLI Reference

```
Usage of mcmanager:
  -listen string
        address to listen for http traffic (default "127.0.0.1:8080")
  -storage string
        where to store server data
```

## Usage

Usage is very straightforward, and options are defined with the default values (use `-h`). Here is an example:

```
$ mcmanager --storage <path/to/storagedir>
```

**NOTE**: If you want your minecraft servers to be available outside your local network, you will need to configure port forwarding on your router/firewall.  MCmanager DOES NOT do this for you.  The server port is shown on the "Servers" page.

**CAUTION**: MCmanager does NOT provide TLS support.  Since logins use existing Minecraft accounts, it is STRONGLY RECOMMENDED that you leave the --listen value as the default and run a proxy service (there are many, Caddy works well) that can provide TLS for you.  This isn't a huge concern if you run this solely inside a home network, but don't expose it to the internet before securing it.  You have been warned. (all mcmanger -> minecraft.net traffic IS over HTTPS, this notice is only about the communication from your web browser to mcmanager)

## Todo
- [x] Authentication (minecraf.net credentials)
  - [x] any authenticated user can create a server instance
  - [ ] restrict who can be "owners", instead of everyone (if desired)
- [x] Authorization
  - [ ] super user: mcmanager account that can see and control all running instances regardless of ownership.
  - owner:
    - [x] delete
    - [x] world re-gen
  - op:
    - [x] op add
    - [x] whitelist add
    - [x] weather
    - [x] time
    - [x] backup
    - [x] save
    - [x] start
    - [x] stop
- Configurable:
  - [ ] news sources (home page content)
  - [x] host name (commane-line flag)
  - [ ] automated backups/saves
  - [ ] schedules for starting/stopping instances (?)
  - more and more and more
- Support server versions:
  - [x] vanilla
  - [x] spigot
  - [ ] craftbukkit
  - plugins:
    - [ ] upload jar
    - [ ] upload config
    - [ ] edit config
    - [ ] delete jar & configs
  - others...
- Instance:
  - [x] port auto-assigned
  - [x] view server address
  - [x] view players online
  - backups:
    - [ ] automated
    - [x] manual
  - save:
    - [ ] automated
    - [x] manual
  - [ ] configuration editing
  - [x] hardcore (on create)
  - [x] game mode (on create)
  - [x] specify seed (on create)
  - [x] choose release (on create)
  - [x] pvp choice (on create)
  - [x] world type (on create)
  - MOTD:
    - [x] set (on create)
    - [ ] change
    - [x] view
  - ops:
    - [x] add
    - [ ] remove
    - [x] view
  - whitelist:
    - [x] add
    - [ ] remove
    - [x] view
    - [x] enable (on create)
    - [x] disable (on create)
  - controls:
    - [x] time set day
    - [x] weather clear
    - [x] world re-gen
    
And so much more.... (please submit a feature request issue)

## License

MIT