# tornado_app.py
import tornado.ioloop
import tornado.web

class MainHandler(tornado.web.RequestHandler):
    def get(self):
        self.write({"message": "Hello from Tornado!"})

def make_app():
    return tornado.web.Application([
        (r"/", MainHandler),
    ])

if __name__ == "__main__":
    app = make_app()
    app.listen(8004)
    print("Running Tornado on http://localhost:8004")
    tornado.ioloop.IOLoop.current().start()
