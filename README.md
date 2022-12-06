# estrys 🐦️🐘

Estrys allow you to bridge some twitter activity to mastodon.

[![codecov](https://codecov.io/gh/estrys/estrys/branch/main/graph/badge.svg?token=J92W6PYFEE)](https://codecov.io/gh/estrys/estrys)

## Why ?

Migrating to mastodon you miss some twitter activity ? Estrys to the rescue !
Mastodon is great, but unfortunately there is still some actors of your Twitter feed that do not have migrated ~~yet~~
to mastodon. This project will allow you to bridge some Twitter user on mastodon.

## Features

- ✅ Work with an essential Twitter API account
- ❌ Backfill tweets

The following Twitter items/actions are currently bridged by Estrys:

- **Tweets**
  - [ ] Tweets
  - [ ] Retweets
  - [ ] Replies
- **Users**
  - [x] Bio
  - [x] Follower/Following/Tweets count
  - [x] Profile image
- **Actions**
  - [ ] Follow
  - [ ] Unfollow

ℹ️ For now the bridge is **unidirectional** FROM Twitter TO mastodon.

That mean that you will not be able to post messages from mastodon to Twitter.

### ⚠️ Usage anti-patterns

**Do not bridge too many Twitter accounts on a single instance**

Estrys instances are limited by the Twitter API rate limits.
We want people to use Estry without having to ask Twitter staff for an elevated access.
If you have an elevated access you could probably use it at scale, but bear in mind that this project was not designed
for that initially.

**The intention behind this project is not to duplicate all twitter within mastodon.**

Please do bridge Twitter users wisely.
The more you follow the most slow it will be to bridge the content (because of API rates limitations)
Also Twitter and Mastodon are two distinct ecosystems and the purpose of this project is not to mirror Twitter to mastodon.
Finally, if you want to replicate a full Twitter feed with thousands of users, you'll also have to follow thousands of users on mastodon.

## Global Limitations

With a Twitter essential account, the following global limitations will apply

* 500k tweets per month

## Getting started

### Create a Twitter app

Head to https://developer.twitter.com/en/portal/petition/use-case to apply for developer account and create an app.

It is recommended to setup a dedicated project for Estrys since it will try to use 100% of your rate limit.

## Config

The default configuration can be found in the `.env` file.
All config options are documented in the `.env` file.

The precedence order is:
* `.env` file
* `.env.local` file
* Environment variables

Every value can be set using env variables using `ESTRYS_` prefix.
For example if you want to change the listen address you can do `ESTRYS_ADDRESS=127.0.0.1:1337`

## How it works

* Estrys manage a list of Twitter users to follow
  * To follow a new user, an instance owner should send a message to the instance admin to ask to bridge a new Twitter user
* Estrys will round over the different Twitter user it's instructed to follow, poll for tweets, if not already stored, store them and publish them.
* [not yet implemented] When restarted, Estrys takes care to get tweets for users starting where it last stopped
* Estrys manages rate limits, it knows the API limits and take care of it in the polling to be as live as the Twitter api allows us
  * ⚠️ That mean that for a given API key Estrys will try to use 100% of your limits

### Limitations

App rate limit (Application-only): 1500 requests per 15-minute window shared among all users of your app

Example:
Your maximum rate is `1500 / 15 * 60 ~= 1.66` req/sec
So if you follow 100 users, you'll have to wait `100 * 1.6 = 166` seconds (~2:47 min) before being able to refresh tweets again for this user.

**Cons**
- Will be less and less live as the instance will follow more accounts
  - Can be mitigated if you use an Oauth2 login (not yet implemented, see [IDEAS](IDEAS.md) for upcoming workarounds)

## Contribute

### Run locally

```bash
cp docker-compose.override.example.yml docker-compose.override.yml
# update the config according to your env if needed
cp .env.local.example .env.local
# configure the app env vars

# start the db and run database migrations
docker-compose up db -d
make db-migrate

# Now run everything
docker-compose up
```

### Run tests

`make test`

## Alternatives

* [Birdsitelive](https://github.com/NicolasConstant/BirdsiteLive)
  * This will require an elevated access
  * From the author : this is more like a sandbox, not really meant for production usage

## Why this name ?

bird (Twitter) + elephant (mastodon) = [elephant bird](https://en.wikipedia.org/wiki/Elephant_bird)

This look like an ostrich, and ostrich in Welsh it is said estrys 😉

Sounds far-fetched, definitively, but it's hard to find good names ...