# ODS Directory Gateway

ODS Directory Gateway is a Go-based service that provides a unified API facade for ODS API.

---

## Features

- **Gateway API**  
  A single HTTP/JSON entry point for directory-style operations (lookup, listing, search, etc.).

- **Extensible Architecture**  
  Internal packages structured so new backends, transports, or behaviors can be added with minimal changes.

- **Production-Oriented Layout**  
  Uses a standard Go project structure with clear separation between public packages, internal logic, and command binaries.

---

## Project Structure

High-level layout:

- `cmd/`  
  Entry points for binaries (e.g. main server executable, admin tools).

- `api/`  
  API definitions, schemas, and OpenAPI/contract specifications for the HTTP interface.

- `client/`  
  Client library for talking to the gateway (e.g. from other Go services).

- `internal/`  
  Internal application logic, handlers, services, and adapters that are not intended to be imported by external code.

- `pkg/`  
  Publicly reusable Go packages that may be imported by other projects.

- `docs/`  
  Bruno collection, additional documentation, design notes, and diagrams.

- `Makefile`  
  Common development tasks (build, test, lint, etc.).

---

## Getting Started

Install dependencies: 
```shell
make deps  
```

Run API on specific port:
```shell
PORT=8086 make run
```

Import bruno collection from `docs/bruno` folder. 
Choose `local` environment in Bruno interface and update `BASE_URL` if needed. 

Both requests from the Bruno collection should already work. 