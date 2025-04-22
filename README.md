
---

# 🚀 FastPaze Framework

**FastPaze** is a high-performance API framework that marries the speed of **Go** with the simplicity of **Python**. It uses Go’s `fasthttp` for ultra-fast HTTP handling and offers a Python interface inspired by **FastAPI** to define routes, middleware, and dependencies.

---

<img src="https://i.ibb.co/h1TLrmQg/Screenshot-2025-04-22-144033.png" alt="Screenshot-2025-04-22-144033" border="0">

---

## ✨ Features

- ⚡ **High Performance**: Built on Go’s `fasthttp`, capable of handling hundreds of thousands of RPS.
- 🐍 **Pythonic Interface**: Define APIs using decorators and classes, similar to FastAPI.
- 🧩 **Middleware Support**: Plug-and-play middleware: Logging, CORS, Rate Limiting, and more.
- 🧪 **Dependency Injection**: Use simple Python functions to inject dependencies into routes.
- 🔧 **Configurable Server**: Tune concurrency, body size, timeouts, and port.
- 🔒 **Thread-Safe**: Leveraging Go's concurrency model with Python coordination.
- 📜 **Path Parameters**: Dynamic URL handling with regex-based parameter parsing.
- 🐞 **Debug Mode**: Include request data in responses during development.
- 🧾 **Structured Logging**: Choose from DEBUG, INFO, WARN, and ERROR levels.

---

---

## 🛠️ Prerequisites

- **Go**: ≥ 1.18
- **Python**: ≥ 3.8
- **C Compiler**: e.g., `gcc`

### Go Dependencies
```bash
go get github.com/valyala/fasthttp
go get github.com/go-playground/validator/v10
```

### Python Dependencies
No external packages required; uses Python standard library only.

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
go build -o libfastpaze.so -buildmode=c-shared libfastpaze.go

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


## 💻 Basic Usage

```python
from fastpaze import FastPaze, HTTPMethod, Parameter, ParamType

app = FastPaze()

@app.get("/hello/{name}", 
         description="Greet a user by name",
         parameters=[Parameter(name="name", in_type=ParamType.PATH, required=True)])
def hello(name: str):
    return {"message": f"Hello, {name}!"}

app.configure_server(port=":9090")
app.start()
```

```bash
python example.py
# → Server runs at http://localhost:9090
```

---

## 💡 Advanced Example

```python
from fastpaze import FastPaze, HTTPMethod, Parameter, ParamType, LogLevel, Depends

app = FastPaze()

# Dependency
def auth_token():
    return "Bearer mock-token"

# Middleware
app.middleware("logging", enabled=True)
app.middleware("cors", enabled=True)
app.middleware("rate_limiter", enabled=True)

# Register Dependency
app.dependency("auth", auth_token())

@app.get("/users/{user_id}",
         description="Get user details",
         parameters=[
             Parameter(name="user_id", in_type=ParamType.PATH, required=True),
             Parameter(name="details", in_type=ParamType.QUERY, type="boolean")
         ],
         dependencies=[Depends(auth_token)])
def get_user(user_id: str):
    return {
        "message": f"User {user_id}",
        "auth_token": "{auth}"
    }

# Server Settings
app.configure_server(
    read_timeout=10000,
    write_timeout=15000,
    idle_timeout=60000,
    max_body_size=8*1024*1024,
    concurrency=512*1024,
    port=":8081"
)

app.set_log_level(LogLevel.DEBUG)
app.include_debug_data(True)
app.start()
```

---

## ⚙️ Server Configuration Options

```python
app.configure_server(
    read_timeout=5000,
    write_timeout=10000,
    idle_timeout=30000,
    max_body_size=4*1024*1024,
    concurrency=256*1024,
    port=":9090"
)
```

---

## 🧩 Middleware

```python
app.middleware("logging", enabled=True)
app.middleware("cors", enabled=True)
app.middleware("rate_limiter", enabled=True)
```

---

## 🔁 Dependency Injection

```python
def database_connection():
    return "mysql://localhost:3306"

app.dependency("db", database_connection())

@app.get("/data", dependencies=[Depends(database_connection)])
def get_data():
    return {"message": "Data fetched", "db": "{db}"}
```

---

## 🧪 Debugging

```python
app.set_log_level(LogLevel.DEBUG)
app.include_debug_data(True)
```

---

## 🧬 Architecture

FastPaze’s architecture bridges the performance of Go with the expressiveness of Python:

- 🟨 **Go Layer**: `fasthttp` handles HTTP processing and concurrency.
- 🟦 **Python Layer**: Provides a user-friendly interface, decorators, and logic definition.
- 🧩 **Integration**: Python uses `ctypes` to load the Go shared library (`libfastpaze.so`).


---

## 🌐 API Testing

```bash
curl http://localhost:9090/hello/world
# → {"message":"Hello, world!"}
```

Use **Postman**, **curl**, or a browser.

---

## ⚠️ Troubleshooting

| Problem | Solution |
|--------|----------|
| `libfastpaze.so not found` | Ensure it exists in the same directory or update `lib_path` |
| Port already in use | Choose a different port using `configure_server` |
| Invalid port | Must be an integer from 1 to 65535 |
| Missing images | Check `images/` folder and filenames |

---

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

## 📦 Package Info

### Editable Install
Running `pip install -e .` installs FastPaze in *editable mode*, meaning any changes to `fastpaze/` will instantly reflect when used in Python — ideal for development.

---

## ✨ Contributing

Check [`CONTRIBUTION.md`](./CONTRIBUTION.md) for guidelines. PRs welcome!

---

## 📜 License

Licensed under the [MIT License](./LICENSE).

---

