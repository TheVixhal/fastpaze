## 📂 Project Structure

```
fastpaze/
├── benchmarks/        # Benchmark scripts for various frameworks
│   ├── benchmark.sh
│   ├── fastapi_app.py
│   ├── fastpaze_app.py
│   ├── flask_app.py
│   ├── sanic_app.py
│   ├── starlette_app.py
│   ├── tornado_app.py
│   └── node_app.js
│
├── examples/          # Example Python usage
│   └── example.py
│
├── fastpaze/          # Core framework (Go + Python interface)
│   ├── __init__.py
│   ├── fastpaze.py
│   ├── libfastpaze.go
│   ├── go.mod
│   └── go.sum
│
├── LICENSE
├── CONTRIBUTION.md
├── README.md
├── requirements.txt
├── setup.py
└── .gitignore
```

---

## 🚀 Getting Started

### 🔧 Installation

```bash
# 1. Clone the repository
git clone https://github.com/thevixhal/fastpaze.git

# 2. Navigate into the core directory
cd fastpaze

# 3. Move into the Go source directory
cd fastpaze

# 4. Compile the Go code into a shared library
go build -buildmode=c-shared -o libfastpaze.so 

libfastpaze.go

# 5. Go back to the root and install the Python package
cd ..

pip install -e .
```

---

## ▶️ Run Example

```bash
cd examples

python example.py
```

> ✅ The server will start at `http://localhost:port`

---

## 🔍 Benchmarking

To compare FastPaze with other frameworks (FastAPI, Flask, Sanic, etc.), run:

```bash
cd benchmarks

chmod +x benchmark.sh

./benchmark.sh
```


---

## 🚀 Benchmark Results

Benchmarks were run using [`autocannon`](https://github.com/mcollina/autocannon) with the default configuration:

```bash
autocannon -c 5 -d 1 -p http://localhost:PORT/hello
```

| Framework     | Requests/sec | Latency (ms) | Throughput (MB/sec) |
|---------------|--------------|--------------|----------------------|
| **FastPaze** (Python + Go) | 🏆 **65,000** | 1.3 | 58.2 |
| FastAPI       | 19,000       | 5.1          | 20.3                 |
| Flask         | 8,500        | 11.2         | 10.1                 |
| Starlette     | 18,000       | 5.5          | 19.8                 |
| Sanic         | 21,000       | 4.6          | 22.4                 |
| Tornado       | 13,000       | 6.8          | 15.1                 |


> **Note**: All tests were run locally on a machine with AMD RYZEN 5, 8GB RAM. FastPaze outperforms other Python frameworks by leveraging Go's high-speed HTTP engine under the hood.

---

---

## 🧪 Example Output

```
$ curl http://localhost:9090/hello/world
{"message": "Hello, world!"}
```

---

## 📦 Package Info

### Editable Install
Running `pip install -e .` installs FastPaze in *editable mode*, meaning any changes to `fastpaze/` will instantly reflect when used in Python — ideal for development.

---
