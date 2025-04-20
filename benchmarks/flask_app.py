from flask import Flask, jsonify

app = Flask(__name__)

@app.route("/hello")
def hello():
    return "Hello"


if __name__ == '__main__':
   app.run(port=8005)
