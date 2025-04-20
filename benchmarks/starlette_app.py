from starlette.applications import Starlette
from starlette.responses import JSONResponse
from starlette.routing import Route

async def hello(request):
    return JSONResponse({"message": "Hello from Starlette"})

app = Starlette(debug=False, routes=[Route("/hello", hello)])
