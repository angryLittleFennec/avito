#!/bin/bash
# Wait for a little while to let the application start and run migrations
sleep 10

# Now run the SQL script
psql -U $POSTGRES_USER -d $POSTGRES_DB -a -f /docker-entrypoint-initdb.d/init.sql
