Testing this package assumes that you have a MySQL server running at localhost:3306.

*WARNING*: a database named `TEMPORARY_DATABASE_boardgame_test` will be created and then dropped by this test. Ensure there's no real data in it!

# Connection strings

The connections trings that are passed to storage.Connect() are of a Data Source Name, described at https://github.com/go-sql-driver/mysql#dsn-data-source-name . 

Normally these strings contain the password, and so they shouldn't be checked
into source control. They are generally configured in config.SECRET.json, in the storageconfig section.

A few examples:

Default for just a basic mamp installation

root:root@tcp(localhost:3306)/boardgame

An example of connecting in prod to a Google Cloud SQL service:

prod:PASSWORD_GOES_HERE@unix(/cloudsql/boardgame-159316:us-east1:prod)/boardgame

where the part after the /cloudsql/ can be derived from running `gcloud sql
instances describe prod`, and noting the connectionName in the result.  Full instructions for that string are here: https://cloud.google.com/appengine/docs/flexible/go/using-cloud-sql


# Databases in production

The engine will only create tables when testmode is true. Otherwise, it will
just assume they exist.

create_tables.sql contains the SQL necessary to create the tables at the
current version.

## Generating create_tables.sql

After making a change that would affect the schema, go to main_test.go, flip
outputTables to true, run go test, then flip it back off. Then go through and
remove the (<time>) at the end of each line.
