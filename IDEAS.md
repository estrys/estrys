#### Timeline polling

Allows you to retrieve a collection of the most recent Tweets and Retweets posted by you and users you follow. This endpoint can return every Tweet created on a timeline over the last 7 days as well as the most recent 800 regardless of creation date.

[Endpoint get-users-id-reverse-chronological](https://developer.twitter.com/en/docs/twitter-api/tweets/timelines/api-reference/get-users-id-reverse-chronological)

##### Mode of operations

* Estrys polls the timeline of different Estrys users (who are Estrys user ?), store and publish them if not already there (others users may have the same tweet in their timeline)
* When restarted, Estrys takes care to get tweets for users starting where it last stopped, for what's in range of the endpoint (thanks to since_id or start_time query parameters)


##### Usecases

##### Limits

User rate limit (User context): 180 requests per 15-minute window per each authenticated user

**Pro**
- Will scale regardless the number of followed accounts per user

**Cons**
- Require to setup Oauth2 authorization
- Will require a twitter account and follow bridged accounts on Twitter


#### Live streaming (not yet implemented)

Streams Tweets in real-time that match the rules estrys added to the stream.
This allow all the flexiblity of the Twitter filtered stream / [rule system](https://developer.twitter.com/en/docs/twitter-api/tweets/filtered-stream/integrate/build-a-rule) like:

[Endpoint get-tweets-search-stream](https://developer.twitter.com/en/docs/twitter-api/tweets/filtered-stream/api-reference/get-tweets-search-stream)

* stream hashtags
*  exclude, or only include tweets with media
* ...

##### Mode of operations

* Estrys has a list of twitter users to follow (how do we specify this, how do we change it ? config, mastodon ?)
* Estrys either
    * constructs the rules based on users you follow
    * you give Estrys twitter rules
* Estrys delete current rules and add its rules
* Estrys subscribes to the stream, store them and publish them

##### Usecases

* I don't care about downtimes / backfilling
* I don't have too much user to follow
* I want my twitter stream rules
* I want to follow hashtags

##### Limits

5 rules max with 512 bytes each, or around 20 users par rule / maximum
So roughly ~100 followed users per instance
50 Tweets/second delivery cap for connections

**Pro**
- Will bridge tweet almost in real time

**Cons**
- The number of followed accounts is very limited