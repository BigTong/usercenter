# userserver
## introduction
- The project is to implement a basic RESTful HTTP service in Go for a simplified Tantan backend: Adding users and swiping other people in order to find a match.
## require
- You should follow the API specification given in this document. We want the data to be stored in a PostgreSQL database. All input and output should be in JSON format, and the API should return the proper HTTP status codes.
## designe

## deploy 

## features

## todo list
- [x] racefully shutdown;
- [x] rate limit;
- [x] test basic api;
- [x] design cache;
- [ ] add basic unitest(done part);
- [ ] add bentch mark;
- [ ] cache data on redis instead of in local mem;
- [ ] Optimize and refact database;
    - [ ] sharding the db;
    - [ ] upsert impl for relation update;
    - [ ] ....
- [ ] refact code;

## benchmark result
