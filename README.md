# ODS Directory Gateway

ODS Directory Gateway is a Go-based service that provides a unified API facade for ODS API.

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

Run API locally:
```shell
make run
```

Run dockerized environment:
```shell
make docker-run
```

Import bruno collection from `docs/bruno` folder. 
Choose `local` environment in Bruno interface and update `BASE_URL` if needed. 

Both requests from the Bruno collection should already work. 