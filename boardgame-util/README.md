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

DefaultPort is the port (e.g. "8888") to use for the api binary when no port
is specified in environment variables.

### DefaultStaticPort

DefaultStaticPort is the default port to use in `boardgame-util serve` for the
static files. Doesn't do anything in prod (since you typically serve the files
using a more generic static file server like firebase)

### DisableAdminChecking

This option *should only be enabled in dev*. When set to true, it disables all
admin checking. That means that any user can enable admin mode clientside and
then operate as an admin (e.g. make whatever moves they want on a game, view
the state from the perspective of any user, etc).

### OfflineDevMode

If provided, this mode will not check that the firebase auth tokens are
legitimate, will use faux authentication clientside (showing a red dot), and
will disable loading roboto fonts in index.html. This is an extremely
dangerous option and may not be enabled in prod.

Typically you only want to enable this in certain conditions, like when you're developing on an airplane. For that reason it's not common to set this in your config.json; typically you instead pass `--offline-dev-mode` to `boardgaume-util serve` to enable it only temporarily.

### AdminUserIds

When adminmode chcecking is enabled (which is the default, see above), only
users whose userId is in this list will be allowed to enable admin mode. You
can find the userIds in the Firebase user console.

### Storage

Storage is how you configure the parameter to be passed to
Storage.Connect(). Different storage backends have different expectations, and
many are fine with just "". When the server is started up, it will fetch the
connection string from this map that matches the Name of the storage engine in
use.

## DefaultStorageType

DefaultStorageType is the type of storage that will be used in boardgame-util
if you don't specify one explicitly at the command line. If you don't specify
one in your config, boardgame-util will fall back on the default type.

## Games

Games is a tree defining the game packages you want included, listed by their
imports, e.g. "github.com/jkomoros/boardgame/examples/checkers" and
"github.com/jkomoros/test-game/example". You can also provide relative paths that point to a valid game package directory, relative to the location of the config.json file they're contained within. Note that if you provide relative paths to `boardgame-util config add games`, the relative paths will be converted to their import, so if you want a relative path in the config file you need to add it manually.

boardgame-util will use `boardgame/boardgame-util/lib/gamepkg` to load up the
provided imports and verify that they are valid game packages on disk. See the
documentation for that library for more on how it finds and validates packages.
`build/api` and `build/static` will use the information in the packages to
generate code that imports them and creates delegates, and also symlinks folders to the underlying client directories for the static assets.

Note that if modules are enabled, when build.Static() is called with a game that is in a module not yet on disk, it will be fetched and cached if necessary.

## GoogleAnalytics

The analytics ID to use client-side. Often has a form like "UA-321655-11"

## ApiHost

The host that the client should reach out to in that context to reach the api
server. In dev context, almost always "http://localhost:8888". In prod
context, something like "https://example-boardgame.appspot.com".

If not provided, this will be automatically derived based on the mode
(dev/prod), the DefaultPort, and the Firebase configuration. This automatic
derivation is almost always what you want.

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