# Simple Go Load Balancer

This is a minimal load balancer written in Go for learning purposes. It distributes incoming HTTP requests across multiple backend servers using a simple round-robin strategy.

---

## Features

- Basic round-robin load balancing  
- Command-line configuration for backend servers  
- Lightweight and easy to run locally  

---

## Requirements

- Go 1.20 or newer  
- Bun (for running the example backend servers)  

---

## Running the Example

### 1. Start the backend servers
Each backend server runs a simple HTTP handler written in Bun (JavaScript/TypeScript).

Run these commands in separate terminals:

```bash
bun run ./server.ts --port 3001
bun run ./server.ts --port 3002
bun run ./server.ts --port 3003
```

Each server will start listening on its respective port.

---

### 2. Start the load balancer

The load balancer will listen on port 8000 and distribute incoming requests among the three backend servers.

```bash
go run main.go -backends http://localhost:3001,http://localhost:3002,http://localhost:3003 -port 8000
```

---

### 3. Test the setup

Open another terminal and send test requests to the load balancer:

```bash
curl http://localhost:8000
```

Repeat the request several times to see the load balancer distribute traffic across the different backend servers.

---

### 4. Stop the servers

Use `Ctrl+C` in each terminal to stop the processes.

---

## References

This project is based on the tutorial from [kasvith.me](https://kasvith.me/posts/lets-create-a-simple-lb-go/).