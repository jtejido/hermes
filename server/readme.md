# Server

Server that will serve files from a hermes instance.

This serves as a basic sample as it implements basic handlers for Get, Set, Delete and Stats.

## Running

To get the parameters:

```
./server -h 
```


## Usage

Once running, adding things to it is as simple as:

```
curl -v -XPUT localhost:8080/hermes/api/cache/example -d "yey"
```

And fetching it is as simple as:

```
curl -v -XGET localhost:8080/hermes/api/cache/example
```

It returns accepts and responds in a form of []byte, so converting and transforming it should be done on the client.

You can also do the following:

```
curl -v -XDELETE localhost:8080/hermes/api/cache/example
curl -v -XGET localhost:8080/hermes/api/stats  // also json {hit, miss, collission, delhit and delmiss}
curl -v -XGET localhost:8080/hermes/api/clear
curl -v -XGET localhost:8080/hermes/api/filterClear // if filter is enabled
```