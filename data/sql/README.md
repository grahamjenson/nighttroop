# nighttrooop sql

Look, rolling my own ORM-ish-thing is probably not a good idea. But options like GORM, ent, sqlc all come with losing control and added layers.
This way:
1. all code is generated and thus runtime type checking is limited (GORM)
2. all sql statements are obvious
3. all calls are explicit, no *

What I would like is better SQL statement parsing for type checking which can be done at compile time, but that can wait.

At the end of the day I would rather have a lot of explicit SQL in one place, than a bunch of SQL building method calls all over the codebase. 
It helps with security, it helps with speed, and it helps with sanity.