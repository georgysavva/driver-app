![Heetch](heetch.png)

## Test


When you open the Heetch app as a passenger, you are able to see a few drivers surrounding you.
These drivers are usually displayed as a car icon. For the release of a new Zombie-based TV show, we are doing a fun partnership! During the week of the TV show release, we will display a zombie icon instead of the usual car icon for specific drivers.

Heetch drivers send their current coordinates to the backend every five seconds. Our application will use those location updates to differentiate between living and zombie drivers, based on a specific predicate (see below).

To support our growth, we have taken the microservice route. So let’s tackle the basics with a HTTP gateway that either forwards requests or transforms them into [NSQ](https://github.com/nsqio/nsq) messages for asynchronous processing. Then we’ll add services that perform tasks related to our mission of transporting people from A to B.

Your task is to implement three services as follows:

- a `gateway` service that either forwards or transforms requests to be respectively processed synchronously or asynchronously
- a `driver location` service that consumes location update events and stores them
- a `zombie driver` service that allows users to check whether a driver is a zombie or not

### 1. Gateway Service

The `Gateway` service is a _public facing service_.
HTTP requests hitting this service are either transformed into [NSQ](https://github.com/nsqio/nsq) messages or forwarded via HTTP to specific services.

The service must be configurable dynamically by loading the provided `gateway/config.yaml` file to register endpoints during its initialization.

Adding new endpoints shouldn't require any code modification except for the `gateway/config.yaml` file, do not hardcode the values in the code.

#### Public Endpoints

`PATCH /drivers/:id/locations`

**Payload**

```json
{
  "latitude": 48.864193,
  "longitude": 2.350498
}
```

**Role:**

During a typical day, thousands of drivers send their coordinates every 5 seconds to this endpoint.

**Behaviour**

Coordinates received on this endpoint are converted to [NSQ](https://github.com/nsqio/nsq) messages listened by the `Driver Location` service.

---

`GET /drivers/:id`

**Response**

```json
{
  "id": 42,
  "zombie": true
}
```

**Role:**

Users request this endpoint to know if a driver is a zombie.
A driver is a zombie if he has driven less than 500 meters in the last 5 minutes.

**Behaviour**

This endpoint forwards the HTTP request to the `Zombie Driver` service.

### 2. Driver Location Service
The `Driver Location` service is a microservice that consumes drivers' location messages published by the `Gateway` service and stores them in a Redis database.

It also provides an internal endpoint that allows other services to retrieve the drivers' locations, filtered and sorted by their addition date

#### Internal Endpoint

`GET /drivers/:id/locations?minutes=5`

**Response**

```json
[
  {
    "latitude": 48.864193,
    "longitude": 2.350498,
    "updated_at": "2018-04-05T22:36:16Z"
  },
  {
    "latitude": 48.863921,
    "longitude":  2.349211,
    "updated_at": "2018-04-05T22:36:21Z"
  }
]
```

**Role:**

This endpoint is called by the `Zombie Driver` service.

**Behaviour**

For a given driver, returns all the locations from the last 5 minutes (given `minutes=5`).


### 3. Zombie Driver Service
The `Zombie Driver` service is a microservice that determines if a driver is a zombie or not.

#### Internal Endpoint

`GET /drivers/:id`

**Response**

```
{
  "id": 42,
  "zombie": true
}
```

**Role:**

This endpoint is called by the `Gateway` service.

**Predicate**

> A driver is a zombie if he has driven less than 500 meters in the last 5 minutes.


Given that this is the first time we do such a partnership, our operational team mentioned that they might need to change the predicate values (duration and distance) through the duration of the partnership. That would allow them to increase the chances of having passengers encounter zombie drivers. For example, on the second day they might decide that a zombie is a driver that hasn't moved more than 2km over the last 30 minutes. So, bonus points if you make these configurable! ;)


**Behaviour**

Returns the zombie state of a given driver.


### Prerequisites
- handle all failure cases
- your code should be tested
- the gateway should be configured using the `gateway/config.yaml` file
- provide a clear explanation of your approach and design choices (while submitting your pull request)
- provide a proper `README.md`:
  - explaining how to setup and run your code
  - including all information you consider useful for a seamless coworker on-boarding

### Workflow
- write your code in **Go**
- you can use the provided `docker-compose.yaml` file to run NSQ and Redis
- create a new branch
- commit and push to this branch
- submit a pull request once you have finished

We will then write a review for your pull request!

### Bonus

- Add metrics / request tracing / circuit breaker 📈
- Add whatever you think is necessary to make the app awesome ✨

⚠️ Do not add your vendors to the git repository. ⚠️
