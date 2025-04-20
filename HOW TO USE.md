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

| **Rank** | **Framework** | **Avg Latency** | **Max Latency** | **Avg Req/Sec** | **Avg Bytes/Sec** |
|:--------:|:--------------|:----------------|:----------------|:----------------|:------------------|
| 🥇 1     | **FastPaze**   | 🟢 0.01 ms       | 18 ms            | 🟢 **28,600**     | **4.84 MB**         |
| 🥈 2     | **Sanic**      | 🟢 0.01 ms       | 13 ms            | 🟢 **13,460**     | 1.7 MB             |
| 🥉 3     | **Starlette**  | 🟡 0.1 ms        | 11 ms            | 5,330            | 847 kB             |
| 4        | **FastAPI**    | 🟡 1.04 ms       | 21 ms            | 3,305            | 730 kB             |
| 5        | **Tornado**    | 🔴 1.5 ms        | 22 ms            | 2,475            | 388 kB             |
| 6        | **Flask**      | 🔴 12.2 ms       | 13 ms            | 🔻 **5**          | 885 B              |

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
