# userserver
## introduction
- The project is to implement a basic RESTful HTTP service in Go for a simplified Tantan backend: Adding users and swiping other people in order to find a match.
## require
- You should follow the API specification given in this document. We want the data to be stored in a PostgreSQL database. All input and output should be in JSON format, and the API should return the proper HTTP status codes.

## design

## deploy 

## features

## todo list
- [x] racefully shutdown;
- [x] simple rate limit;
- [x] test basic api;
- [x] design cache;
- [x] add basic unitest(done part);
- [x] add bentch mark;
- [ ] cache data on redis instead of in local mem;
- [ ] Optimize and refact database;
    - [ ] sharding the db;
    - [ ] upsert impl for relation update;
    - [ ] ....
- [ ] refact code;

## benchmark result
- BenchmarkGetUserRelation-8        2000            796181 ns/op
- BenchmarkAddUserRelation-8        1000           1238236 ns/op
- BenchmarkGetUserList-8        2000            609328 ns/op
- BenchmarkAddUserData-8        2000            687842 ns/op

