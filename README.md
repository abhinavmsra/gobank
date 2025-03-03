# Concurrent Money Transfer System

This project provides a **concurrent money transfer system** using **Go**.

---

## Docker Setup

To set up the project using **Docker**, follow these steps:

1. Build and start the container (runs in the background):
  ```sh
   docker compose up -d
  ```


2.	Access the running container:
  ```sh
  docker compose exec -it app bash
  ```

3. Run the API server:

  ```sh
  docker compose exec -it app bash -c "go run ."
  ```

## Seed Data

By default, the system loads initial account balances from `accounts.json`.

```json
[
  {"name": "Mark", "balance": 100},
  {"name": "Jane", "balance": 50},
  {"name": "Adam", "balance": 0}
]
```

## API

1. Transfer Money
	•	Send a **POST** request to `/transfer` with the following JSON payload:

```json
{
  "from": "Mark",
  "to": "Jane",
  "amount": 30
}
```

```sh
curl -X POST http://localhost:3000/transfer -H "Content-Type: application/json" -d '{"from":"Mark", "to":"Jane", "amount":30}'
```

2. Expected API Response

On success, the response returns the updated balances:

```json
{
  "from": "Mark",
  "from_balance": 70,
  "to": "Jane",
  "to_balance": 80
}
```

If an error occurs (e.g., insufficient funds), the response includes 400 status code with an error message:

```sh
"insufficient funds"
```

## Tests

To run unit tests inside the Docker container:

```sh
docker compose exec -it app bash -c "go test -v"
```

## Locking Strategy

To ensure safe concurrent balance updates, we use Go’s `sync.Mutex` for locking.

### How It Works:

1.	Each account has its own `sync.Mutex` lock.

2.	Before updating balances, the system locks both the sender and receiver accounts.

3.	Locks are always acquired in a consistent order (alphabetical order of account names) to prevent deadlocks (Mark before Jane).This ensures that if two transactions happen simultaneously, they always lock in the same order, avoiding deadlocks.

4.	Once the transaction is complete, the locks are released.
