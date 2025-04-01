#!/bin/bash

cd sql/schema
goose postgres "postgres://frankiebrown:@localhost:5432/gator" down
goose postgres "postgres://frankiebrown:@localhost:5432/gator" up