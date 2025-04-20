from fastpaze import FastPaze, Parameter, ParamType, RequestBody, Depends, HTTPMethod, LogLevel

# Create FastPaze application
app = FastPaze()

# Set log level to debug for development
app.set_log_level(LogLevel.DEBUG)

# Don't include debug data in responses
app.include_debug_data(False)

# Enable middleware
app.middleware("logging", enabled=True)

# Route with path parameters
@app.get("/users/{user_id}", 
    description="Get user by ID",
    parameters=[
        Parameter(
            name="user_id",
            in_type=ParamType.PATH,
            description="The user ID",
            required=True,
            type="integer"
        )
    ]
)
def get_user(user_id):
    # Return a dictionary with placeholders that will be replaced
    # The {user_id} placeholder will be replaced with the actual value
    return {
        "message": f"User details retrieved",
        "user_id": user_id,
        "username": "user_" + user_id,
        "email": f"user{user_id}@example.com"
    }

# Route with multiple path parameters
@app.get("/posts/{post_id}/comments/{comment_id}", 
    description="Get a specific comment on a post",
    parameters=[
        Parameter(
            name="post_id",
            in_type=ParamType.PATH,
            description="The post ID",
            required=True,
            type="integer"
        ),
        Parameter(
            name="comment_id",
            in_type=ParamType.PATH,
            description="The comment ID",
            required=True,
            type="integer"
        )
    ]
)
def get_comment(post_id, comment_id):
    # Both post_id and comment_id will be replaced with actual values
    return {
        "message": "Comment details retrieved",
        "post_id": post_id,
        "comment_id": comment_id,
        "content": f"This is comment {comment_id} on post {post_id}"
    }

# Simple hello endpoint
@app.get("/hello", description="Hello endpoint")
def hello():
    return {
        "message": "Hello, World!",
        "version": "1.0"
    }

if __name__ == "__main__":
    print("Starting FastPaze server...")
    app.start(port=8000)