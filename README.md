## Suggest areas for improvement given more time

This was my first time working with Docker, Go and proper testing therefore I'm sure there are many ways this project could be improved.
These include both the code structure and better following of Go's conventions, but also API improvements such as:

* Some DB queries/logic could be improved, leading to better performance of the API when dealing with concurrent requests.
* The error handling could be improved in order to give the user better feedback about what happened (e.r. more HTTP error codes corresponding to the specific issues instead of generic ones).
* Addition of more venue related endpoints in order to retrieve infor about tables, updating the number of seats or removing a table.
* Improvement of tests. Addition of DB method specific tests. Currently tests are mostly focused endpoints/handlers which implicitly test the remaining methods.

## Relevant files

The package files are on `cmd/app`

All the testing is done on `cmd/app/main_test.go`

The MySQL DB is initialized on `docker/mysql/dump.sql`


## Running the application
The following command runs the app and a mysql db on docker containers:
```
make docker-up
```

## Running tests the application
Tests can be run on docker with the following command:

```
make docker-test
```

## Cleaning up
```
make docker-down
```

## API guide

### Add a guest to the guestlist

If there is insufficient space at the specified table, throws an error (http.StatusConflict).

```
POST /guest_list/name
body: 
{
    "table": int,
    "accompanying_guests": int
}
response: 
{
    "name": "string"
}
```

### Get the guest list

```
GET /guest_list
response: 
{
    "guests": [
        {
            "name": "string",
            "table": int,
            "accompanying_guests": int
        }, ...
    ]
}
```

### Guest Arrives

A guest may arrive with an entourage that is not the size indicated at the guest list.
If the table is expected to have space for the extras, allow them to come. Otherwise, this method throws an error (http.StatusConflict).

```
PUT /guests/name
body:
{
    "accompanying_guests": int
}
response:
{
    "name": "string"
}
```

### Guest Leaves

When a guest leaves, all their accompanying guests leave as well.

```
DELETE /guests/name
```

### Get arrived guests

```
GET /guests
response: 
{
    "guests": [
        {
            "name": "string",
            "accompanying_guests": int,
            "time_arrived": "string"
        }
    ]
}
```

### Count number of empty seats

```
GET /seats_empty
response:
{
    "seats_empty": int
}
```


### Get the guest - NEW Endpoint

GET /guests/name
response:
{
    "name": "string",
    "table": int,
    "accompanying_guests": int,
	"arrived": bool
}


### Add a new table - NEW Endpoint

NEW useful endpoint

```
POST /venue
body:
{
	"seats": int
}
```