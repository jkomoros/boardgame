Testing this package assumes that you have a MySQL server running at localhost:3306.

*WARNING*: a database named `TEMPORARY_DATABASE_boardgame_test` will be created and then dropped by this test. Ensure there's no real data in it!

# Connection strings

The connections trings that are passed to storage.Connect() are of a Data Source Name, described at https://github.com/go-sql-driver/mysql#dsn-data-source-name . 

Normally these strings contain the password, and so they shouldn't be checked
into source control. They are generally configured in config.SECRET.json, in the storageconfig section.

Currently the only db name that is supported is `boardgame`

A few examples:

Default for just a basic mamp installation

root:root@tcp(localhost:3306)/boardgame

An example of connecting in prod to a Google Cloud SQL service:

prod:PASSWORD_GOES_HERE@unix(/cloudsql/boardgame-159316:us-east1:prod)/boardgame

where the part after the /cloudsql/ can be derived from running `gcloud sql
instances describe prod`, and noting the connectionName in the result.  Full instructions for that string are here: https://cloud.google.com/appengine/docs/flexible/go/using-cloud-sql


# Creating the database

The `boardgame-util db` tool is designed to help administer your database. 

To set up a database, configure the DSN as described above. Then, sitting in the same folder as config.SECRET.json, run `boardgame-util db setup` (include `--prod` if you want to run on the prod database.

# Making sure the database is up-to-date

Before doing a push to prod it's a good idea to make sure the database is set up correctly with the most recent changes since the last push. Run `boardgame-util db up` to make sure all migrations are applied.

## Rolling deployment of migration 0023

Migration 0023 adds the durable shared-Table transfer columns. Deploy it in
this order when more than one server instance may be live:

1. Deploy binaries containing the 0023 compatibility code to every instance
   while the database is still at 0022. Ordinary Table leases continue using
   the explicit legacy column set; transfer creation remains unavailable.
2. After every older binary has drained, run `boardgame-util db up`.
3. Verify transfer creation, then proceed with later releases.

Do not apply 0023 while pre-compatibility binaries are still running: those
binaries used `SELECT *` and cannot decode the additional columns. For rollback,
first return every instance to a compatibility-capable binary, then run the
0023 down migration. Never remove the columns while a transfer-capable binary
is actively creating transfers.

# Updating the database structure

When making a change to the database structure, create two files in mysql/migrations, named `NNNN_<name-of-change>.down.sql` and `NNNN_<name-of-change>.up.sql` where `NNNN` is the next sequence number. (Don't forget to add them with `git add`)
