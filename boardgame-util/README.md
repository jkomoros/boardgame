## Providing configuration to boardgame-util

You configure boardgame-util via one or more config.json files. The name of the
file must be config.json, or config.*.json, where * is anything other than
"SECRET". These files must be in the directory you will start the command from.

Some of the information in these files is secret and should not be committed
to source control, while some is not. You may have two files: one named
generically, for any information it's OK to check into source control. If you
have a file named "config.SECRET.json", then that will also be loaded up, and
override anything in the base file (or used directly if nothing is in the base
file). This allows you to add a line to your .gitignore that makes it
impossible for you to accidentally commit the information in
config.SECRET.json. Generally you want to keep all of the non-secret aspects
in config.

Within a config file there are three sub configs: "base", dev" and "prod".
"base" is never used directly, but it sets defaults that "dev" and "prod" will
override and extend.

Both "dev" and "prod" have the same possible fields to set. The server picks
which one to use at start up based on the GIN_MODE environment variable.

### AllowedOrigins

AllowedOrigins*is a comma-delimited list of origins to use in CORS that should
be allowed to access your endpoint.

### DefaultPort

DefaultPort is the port (e.g. "8080") to use when no port is specified in
environment variables.

### DisableAdminChecking

This option *should only be enabled in dev*. When set to true, it disables all
admin checking. That means that any user can enable admin mode clientside and
then operate as an admin (e.g. make whatever moves they want on a game, view
the state from the perspective of any user, etc).

### AdminUserIds

When adminmode chcecking is enabled (which is the default, see above), only
users whose userId is in this list will be allowed to enable admin mode. You
can find the userIds in the Firebase user console.

### StorageConfig

StorageConfig is how you configure the parameter to be passed to
Storage.Connect(). Different storage backends have different expectations, and
many are fine with just "". When the server is started up, it will fetch the
connection string from this map that matches the Name of the storage engine in
use.

## DefaultStorageType

DefaultStorageType is the type of storage that will be used in boardgame-util
if you don't specify one explicitly at the command line. If you don't specify
one in your config, boardgame-util will fall back on the default type.

## Games

Games is a tree defining the game packages you want included. This isn't used
for antyhing yet, but will allow commands in `boardgame-util` to know which
games you want to by default operate on.


## GoogleAnalytics

The analytics ID to use client-side. Often has a form like "UA-321655-11"

## ApiHost

The host that the client should reach out to in that context to reach the api
server. In dev context, almost always "http://localhost:8888". In prod
context, something like "https://example-boardgame.appspot.com"

## Firebase

Firebase is a sub-object that contains the configuration for firebase. The
fields are a straight forward application of what you get from firebase.
Mostly used when generating client_config.json.

### ApiKey

Has a form like "AIzaSyDi0hhBgLPbpJabcVCDzDkk8zuFpb9XadM"

### AuthDomain

Has a form like "example-boardgame.firebaseapp.com"

### DatabaseURL

Has a form like "https://example-boardgame.firebaseio.com"

### ProjectID

Has a form like "example-boardgame" . This field is also used in api server to
validate logins.

### StorageBucket

Has a form like "example-boardgame.appspot.com"

### MessagingSenderID

Has a form like "138149526364"