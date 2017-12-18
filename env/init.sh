#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE dbmain;
    \c dbmain;
    CREATE TABLE person (
    "id" serial primary key,
    "name" varchar(80),
    "job" varchar(100),
    "salary" real,
    "month" date
)
EOSQL