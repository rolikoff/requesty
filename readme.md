# A simple RESTful web service API example

The example was made to understand the basic principles of how to build a web service with golang.
The idea of this microservice is to count the number of requests to domains and then aggregate the top 10 domains based on the time and number of requests.
It does accept POST requests with data such as `{"timestamp":1635611609, "example.com": 10, "foobar.com": 20}`, dynamically parses that JSON and put all data into `domains` table. `10` and `20` are called counters, that data represents how many times the domain was requested.
There's a few GET endpoints that are used to show top 10 domains requested within last round minute and last round hour.

The app uses [gin web framework](https://github.com/gin-gonic/gin) and SQLite3 for storing the data.

When the app is started, it creates a new DB using the path that is provided during app initialization. Gin binds a few routes to http.Server and starts listening and serving HTTP requests.

## REST API documentation

Since this is a simple example app, all endpoints here are public, no authentication required.

### Domains

Used to collect domain requests.

[Domains](#) `POST /domains`

Used to get top 10 domains requested in last round minute.

[Show top 10 last round minute](#) `GET /domains/statistics/last-minute`

Used to get top 10 domains requested in last round hour.

[Show top 10 last round minute](#) `GET /domains/statistics/last-hour`
