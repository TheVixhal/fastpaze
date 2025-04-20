from fastpaze import FastPaze, LogLevel

# Create FastPaze application
app = FastPaze()

# Configure server for max performance
app.configure_server(
    read_timeout=5,  # 5ms for fast reads 
    write_timeout=5,  # 5ms for quick writes
    idle_timeout=60000,  # 60s for keep-alive
    max_body_size=65536,  # 64KB for GET-only APIs
    concurrency=1000000,  # 1M for massive concurrency
    port=":8007",
)

app.set_log_level(LogLevel.INFO)

# Simple hello endpoint
@app.get("/hello")
def hello():
    return {
        "message": "Hello, World!",
        "version": "1.0"
    }

if __name__ == "__main__":
    print("Starting FastPaze server with optimized config...")
    app.start()