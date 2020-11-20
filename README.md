## Overview

When you open the app as a passenger, you can see a few drivers surrounding you.
These drivers are usually displayed as a car icon. For the release of a new Zombie-based TV show, we want to display a zombie icon instead of the usual car icon for specific drivers.

Drivers send their current coordinates to the backend every five seconds. Our application will use those location updates to differentiate between living and zombie drivers, based on a specific predicate (see below).

The app consists of three microservices:

- a `gateway` HTTP gateway service that either forwards or transforms requests into [NSQ](https://github.com/nsqio/nsq) messages to be respectively processed synchronously or asynchronously
- a `driver location` service that consumes location update events and stores them
- a `zombie driver` service that allows users to check whether a driver is a zombie or not

### 1. Gateway Service

The `Gateway` service is a _public facing service_.
HTTP requests hitting this service are either transformed into [NSQ](https://github.com/nsqio/nsq) messages or forwarded via HTTP to specific services.

The service is dynamically configured by loading the provided `gateway/config.yaml` file to register endpoints during its initialization.

Adding new endpoints don't require any code modification except for the `gateway/config.yaml` file.

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


The predicate values (duration and distance) are configurable. That allows us to increase the chances of having passengers encounter zombie drivers. For example, a zombie is a driver that hasn't moved more than 2km over the last 30 minutes.


**Behaviour**

Returns the zombie state of a given driver.

## Implementation details
- All services follow clean/hex architecture
- The code is tested. All tests are running without any external dependency and don’t require any specific environment.
- The code is protected by `golangci-lint`
- The app is configurable via `.yaml` files
- All services packaged with Docker

## Setup

In the project root do:

- Run all tests: `make test`
- Build Docker images for each service: `make all`
- Run everything `docker-compose up`
