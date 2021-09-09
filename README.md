# Phanpy

Phanpy is dead simple JSON streaming HTTP interface over Postgres to execute your queries.

## Motivation

Unlike [PostgREST](https://postgrest.org/en/v8.0/index.html) Phanpy only embraces the philosophy of 
providing an HTTP interface to Postgres and embraces SQL queries and parameters as first class citizens. 

## Zero Abstraction
Phanpy takes SQL for what it is; a well understood, well document, and well practiced language 
with precise controls that no other DSL has matched so far. That's why projects like Spark, Cassandra etc. adopted
various dialects of SQL. All queries are prepared and executed as is!

## Tutorial

Phanpy requires only two environment variables:
 - `ENV` enviroment of running `prod` means production where it uses production loggers, everything else is development
 - `DB_URL` a full URL according to [pgx](https://github.com/jackc/pgx) specifications e.g. `postgresql://<user>:<password>@localhost:5432/<database>?statement_cache_capacity=128&statement_cache_mode=prepare&pool_max_conns=16`

After launching `phanpy` you can use cURL to run your queries:

```bash
curl -v --data '{"query": "SELECT * FROM generate_series(1, 1000000)", "params": [], "timeout": 30}' http://localhost:8080/
```

## Performance

Here is the example query payload used to generate upto 400 rows at random, with random ordering and some text payload.

```json
{
  "query": "SELECT md5(random()::text) AS some_data, * FROM generate_series(1, (random() * $1)::int) ORDER BY RANDOM()",
  "params": [400],
  "timeout": 30
}
```

Benchmarking with [vegeta](https://github.com/tsenart/vegeta) here are rough benchmarks with everything running on 
(2015 Macbook pro, 2.2 GHz 8 core with 16 GB RAM) same machine (all vegeta, postgres, phanpy):

```bash
➜  phanpy git:(master) cat vegeta.txt | vegeta attack -rate=1000 -duration=30s | vegeta report
Requests      [total, rate, throughput]         30000, 1000.03, 999.78
Duration      [total, attack, wait]             30.007s, 29.999s, 7.478ms
Latencies     [min, mean, 50, 90, 95, 99, max]  400.425µs, 9.288ms, 6.457ms, 17.67ms, 28.689ms, 59.23ms, 166.477ms
Bytes In      [total, mean]                     422840350, 14094.68
Bytes Out     [total, mean]                     4710000, 157.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:30000
Error Set:
```

During all of this maximum memory was reported under 50MB RSS; avg runtime stats looked something like this:

```json
{
  "acquire_count": 101490,
  "acquire_duration": 508260937427,
  "acquired_conns": 16,
  "empty_acquire_count": 12972,
  "idle_conns": 0,
  "max_conns": 16,
  "stats": {
    "time": 1631159177493892000,
    "go_version": "go1.17",
    "go_os": "darwin",
    "go_arch": "amd64",
    "cpu_num": 8,
    "goroutine_num": 213,
    "gomaxprocs": 8,
    "cgo_call_num": 133,
    "memory_alloc": 13717472,
    "memory_total_alloc": 21575081440,
    "memory_sys": 44714768,
    "memory_lookups": 0,
    "memory_mallocs": 374057114,
    "memory_frees": 373986984,
    "memory_stack": 2424832,
    "heap_alloc": 13717472,
    "heap_sys": 35323904,
    "heap_idle": 16744448,
    "heap_inuse": 18579456,
    "heap_released": 6774784,
    "heap_objects": 70130,
    "gc_next": 21642400,
    "gc_last": 1631159177475508000,
    "gc_num": 2348,
    "gc_per_second": 20.843679212802737,
    "gc_pause_per_second": 42.506663,
    "gc_pause": [
      0.19,
      0.232078,
      0.310384,
      0.175009,
      0.151908,
      0.178259,
      0.181853,
      "more ..."
    ]
  },
  "total_conns": 16
}
```
