# Distributed Object Storage Engine

A distributed object storage system built from scratch in Go — files are split into fixed-size, content-addressed chunks, distributed across independent storage nodes, and reassembled on retrieval. Built to understand the core mechanics behind systems like S3 and GFS: chunking, node distribution, coordination, and (soon) replication and fault tolerance.

Not a "Google Drive clone" — this is a scoped-down engine focused on the distributed-systems fundamentals: dumb, replaceable storage nodes and a coordinator that splits, distributes, tracks, and reassembles data across them.

## Why this project

Most fresher backend portfolios stop at CRUD APIs. This project was built to go one level deeper — into the infrastructure patterns that back real storage systems: content-addressed chunking, round-robin data distribution, and coordinator/worker architecture, all implemented with Go's concurrency primitives.

## Architecture

```
                     ┌─────────────┐
   Client  ───PUT───▶│ Coordinator │
                     └──────┬──────┘
                            │ splits file into chunks,
                            │ distributes via round robin
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
         ┌─────────┐  ┌─────────┐  ┌─────────┐
         │ Node 1  │  │ Node 2  │  │ Node 3  │
         │ :8081   │  │ :8082   │  │ :8083   │
         └─────────┘  └─────────┘  └─────────┘
```

**Storage nodes** — Deliberately "dumb" and identical. Each one is a standalone HTTP server exposing:
- `PUT /chunks/{id}` — stores raw bytes to disk under the given ID
- `GET /chunks/{id}` — retrieves them back, with proper 404 handling for missing chunks

Nodes have no knowledge of each other, no knowledge of files, and no knowledge of the coordinator. They just store and serve bytes. Port and storage directory are configurable via CLI flags, so the same binary can run as any number of independent nodes.

**Chunker** — Splits an incoming file into fixed-size chunks (4MB each) and content-addresses each chunk with its SHA-256 hash, which doubles as the chunk's unique ID (identical content always produces the same ID).

**Coordinator** — The only component with global knowledge. On upload, it:
1. Reads the incoming file
2. Splits it into chunks via the chunker
3. Distributes each chunk to a storage node using round-robin selection
4. Verifies each node accepted the chunk (checks response status, not just network success)
5. Records two in-memory maps: `fileName → ordered chunk IDs` and `chunkID → node address`

On download, it reverses the process — looks up the file's chunk list, resolves each chunk's node, fetches them in order, and reassembles the original file.

## Tech stack

- **Go 1.22** — `net/http` with method-based routing (`http.ServeMux` path patterns), no external frameworks
- Standard library only: `crypto/sha256`, `flag`, `io`, `os`

## Project structure

```
Backend/
├── go.mod
├── chunker/
│   └── chunker.go       # fixed-size chunking + SHA-256 content addressing
├── server/
│   └── server.go        # storage node (PUT/GET chunks, disk-backed)
└── coordinator/
    └── coordinator.go   # upload/download orchestration, round-robin distribution
```

## Running it locally

Start three storage nodes, each with its own port and storage directory:

```bash
cd server
go run server.go -port=8081 -storage="./storage1"
go run server.go -port=8082 -storage="./storage2"
go run server.go -port=8083 -storage="./storage3"
```

Start the coordinator:

```bash
cd coordinator
go run coordinator.go -port=9000
```

Upload a file:

```bash
curl -X PUT http://localhost:9000/files/myfile.txt -d "some file content here"
```

Download it back:

```bash
curl http://localhost:9000/files/myfile.txt
```

## What's implemented

- [x] Fixed-size file chunking with SHA-256 content addressing
- [x] Standalone storage nodes with disk-backed chunk PUT/GET
- [x] Configurable nodes via CLI flags (port, storage directory)
- [x] Coordinator: upload flow — chunk, distribute (round robin), verify, track
- [x] Coordinator: download flow — lookup, fetch, reassemble

## Roadmap

- [ ] Replication — write each chunk to 2+ nodes so a single node failure doesn't lose data
- [ ] Checksum verification on the download path using the SHA-256 hashes already computed at chunk time
- [ ] Concurrent chunk transfer via goroutines instead of sequential upload/download
- [ ] Failure handling — detect a down node and serve from a replica
- [ ] Persist coordinator metadata (currently in-memory only, lost on restart)

## Design notes

- **Chunk size**: fixed at 4MB. Large enough to avoid excessive per-chunk overhead, small enough to enable parallel transfer once concurrency is added.
- **Node placement strategy**: round robin for v1 — simple, stateless, and reasonably load-balanced for similarly-sized chunks. Load-aware placement is a possible future improvement once nodes report real-time load back to the coordinator.
- **Nodes are intentionally dumb**: no cross-node awareness, no replication logic on the node side. All intelligence lives in the coordinator, keeping storage nodes trivially replaceable.
