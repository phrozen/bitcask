# bitcask

**Work In Progress**

A pure Go implementation of the [BitCask](https://riak.com/assets/bitcask-intro.pdf) model with no dependecies (hopefully).

It is NOT meant to be a 1:1 replacement for Bitcask as [Riak](https://riak.com/) uses it, nor it is meant to be fully compatible with original Erlang's Bitcask database files.

It IS meant to be an idiomatic Go implementation of the Log-Structured Merge (LSM) Trees with a Write Ahead Log (WAL) model that Bitcask uses. It IS meant to be performant and easy to read through the code, and will implement AT LEAST all the basic features described on [BitCask's model design paper](https://riak.com/assets/bitcask-intro.pdf).

## Introduction

> Bitcask is an Erlang application that provides an API for storing and retrieving key/value data into a log-structured hash table.

It is a fast persistent key/value store with predictable performance and strong fault tolerance, features as per the design document:

- low latency per item read or written
- high throughput, especially when writing an incoming stream of random items
- ability to handle datasets much larger than RAM w/o degradation
- crash friendliness, both in terms of fast recovery and not losing data
- ease of backup and restore
- a relatively simple, understandable (and thus supportable) code structure and data format
- predictable behavior under heavy access load or large volume

You can find the original Bitcask Erlang implementation at: https://github.com/basho/bitcask

## Roadmap

- [x] Create new database
- [ ] Open existing database
- [ ] Verify lock for opening existing database
- [x] List, Put, Get, Delete keys and values
- [ ] Default sane options and config
- [ ] Flushable write buffer and Sync configuration
- [ ] Search, Scan, Prefix, Indexing (Radix tree?)
- [ ] Merge and compact database files
- [ ] Create hint files for fat load times
- [ ] Global time expiration of keys
- [ ] Extensive unit testing and benchmarks
- [ ] More...

## API

```go

func bitcask.New(directory string, options bitcask.Opts) (*bitcask.Bitcask, error)

func (bc *Bitcask) Put(key, value []byte) error

func (bc *Bitcask) Get(key []byte) ([]byte, error)

func (bc *Bitcask) Delete(key []byte) error

```

## Benchmarks

Soon...