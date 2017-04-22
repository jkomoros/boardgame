Testing this package assumes that you have a MySQL server running at localhost:3306, and that it has a database called boardgame_test.

*WARNING*: the database boardgame_test will be cleared out during the test!

If you don't have such a database, run the following query (once connected to the db):

```
CREATE DATABASE boardgame_test;
```

# Databases in production

The engine will only create tables when testmode is true. Otherwise, it will
just assume they exist.

create_tables.sql contains the SQL necessary to create the tables at the
current version.

## Generating create_tables.sql

After making a change that would affect the schema, go to main_test.go, flip
outputTables to true, run go test, then flip it back off. Then go through and
remove the (<time>) at the end of each line.
