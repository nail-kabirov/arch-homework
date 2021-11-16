#!/bin/bash

# add two users
curl -L -X POST 'arch.homework/api/v1/user' -H 'Content-Type: application/json' --data-raw '{"username": "johndoe","firstName": "John","lastName": "Doe","email": "john@doe.com","phone": "+71002003040"}'
curl -L -X POST 'arch.homework/api/v1/user' -H 'Content-Type: application/json' --data-raw '{"username": "janedoe","firstName": "Jane","lastName": "Doe","email": "jane@doe.com","phone": "+71002003050"}'

while true; do
    # get all users
    ab -n 1000 -c 50 http://arch.homework/api/v1/users
    # get non-existent user
    ERR_COUNT=$((1 + $RANDOM % 40))
    ab -n $ERR_COUNT -c $ERR_COUNT http://arch.homework/api/v1/user/00000000-0000-0000-0000-000000000000
    sleep 3
done
