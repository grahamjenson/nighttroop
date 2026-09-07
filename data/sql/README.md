# Nighttroop SQL

Nighttroop uses generated code with SQLc, goverter and a few custom scripts to create the data-layer interface.
All queries in Nighttroop are prepared statements, making it possible to test some aspects of their correctness at compile time.

Having all SQL explicit and in sql, rater than ORM-ish, hopefully makes all interactions more obvious and easier to optimize in the future.

# Control DB

This DB has all the 

# Site DB

This has information that is related to the user, typically mastered elsewhere like AAPS or Loop.
It is the 