from ctypes import cdll, c_char_p, c_int, c_size_t
import os
import json
import inspect
import typing
from typing import Any, Callable, Dict, List, Optional, Type, Union
from enum import Enum
from dataclasses import dataclass, field, asdict
import functools
import re

class HTTPMethod(str, Enum):
    GET = "GET"
    POST = "POST"
    PUT = "PUT"
    DELETE = "DELETE"
    PATCH = "PATCH"
    OPTIONS = "OPTIONS"
    HEAD = "HEAD"

class ParamType(str, Enum):
    PATH = "path"
    QUERY = "query"
    HEADER = "header"
    COOKIE = "cookie"

class LogLevel(str, Enum):
    DEBUG = "debug"
    INFO = "info"
    WARN = "warn"
    ERROR = "error"

@dataclass
class Parameter:
    name: str
    in_type: ParamType  # 'in' is a Python keyword, so using 'in_type'
    description: str = ""
    required: bool = False
    type: str = "string"
    schema: Optional[str] = None

@dataclass
class RequestBody:
    description: str = ""
    required: bool = False
    content_type: str = "application/json"
    schema: str = ""

@dataclass
class RouteOptions:
    path: str
    method: HTTPMethod
    description: str = ""
    parameters: List[Parameter] = field(default_factory=list)
    request_body: Optional[RequestBody] = None
    responses: Dict[int, str] = field(default_factory=lambda: {200: "Successful response"})
    dependencies: List[str] = field(default_factory=list)

class Depends:
    """Dependency injection container similar to FastAPI's Depends"""
    def __init__(self, dependency: Callable):
        self.dependency = dependency
        
    def __call__(self):
        return self.dependency()

class FastPaze:
    def __init__(self):
        try:
            self.lib = cdll.LoadLibrary("./libfastpaze.so")
            self.lib.RegisterRoute.argtypes = [c_char_p, c_char_p, c_char_p, c_char_p]
            self.lib.RegisterRouteWithParams.argtypes = [c_char_p, c_char_p, c_char_p, c_char_p, c_char_p]
            self.lib.RegisterMiddleware.argtypes = [c_char_p, c_int]
            self.lib.RegisterDependency.argtypes = [c_char_p, c_char_p]
            self.lib.SetServerConfig.argtypes = [c_size_t, c_size_t, c_size_t, c_size_t, c_size_t]
            self.lib.SetLogLevel.argtypes = [c_char_p]
            self.lib.SetIncludeDebugData.argtypes = [c_int]
            
            # Store dependencies for injection
            self._dependencies = {}
            
            # Set default log level
            self.set_log_level(LogLevel.INFO)
            
            # Disable debug data in responses by default
            self.include_debug_data(False)
        except OSError as e:
            raise RuntimeError(f"Failed to load libfastpaze.so: {e}")

    def _extract_path_params(self, path: str) -> List[str]:
        """Extract path parameter names from a route path."""
        param_pattern = re.compile(r'{([^{}]+)}')
        return param_pattern.findall(path)

    def route(self, path: str, method: HTTPMethod = HTTPMethod.GET, description: str = "", 
              parameters: List[Parameter] = None, request_body: RequestBody = None,
              responses: Dict[int, str] = None, dependencies: List[Union[str, Depends]] = None):
        """
        Decorator to register a route with FastPaze
        
        Args:
            path: URL path for the endpoint
            method: HTTP method (GET, POST, etc.)
            description: Description for API documentation
            parameters: List of parameters (path, query, header, cookie)
            request_body: Request body specification
            responses: Dictionary of response codes and descriptions
            dependencies: List of dependencies required by this endpoint
        """
        def decorator(func):
            # Extract path parameters
            path_params = self._extract_path_params(path)
            
            # Check if the function signature matches the path parameters
            sig = inspect.signature(func)
            func_params = list(sig.parameters.keys())
            
            # Verify all path parameters are in the function signature
            for param in path_params:
                if param not in func_params:
                    raise ValueError(f"Path parameter '{param}' is not defined in the function signature of '{func.__name__}'")
            
            # Create a wrapper that calls the function with mock parameters
            @functools.wraps(func)
            def wrapper():
                # Create a dictionary of mock path parameters
                mock_params = {}
                for param in path_params:
                    # Use the parameter name as a placeholder in curly braces
                    mock_params[param] = "{" + param + "}"
                
                # Call the original function with mock path parameters
                result = func(**mock_params)
                
                # If result is a dict, convert to JSON string
                if isinstance(result, dict):
                    return json.dumps(result)
                
                # Otherwise, return as string
                return str(result)
            
            # Get the message template from the wrapper
            message_template = wrapper()
            
            # Prepare parameters if provided
            if parameters:
                params_json = json.dumps([{
                    "name": p.name,
                    "in": p.in_type,
                    "description": p.description,
                    "required": p.required,
                    "type": p.type,
                    "schema": p.schema
                } for p in parameters])
                
                self.lib.RegisterRouteWithParams(
                    path.encode('utf-8'),
                    method.value.encode('utf-8'),
                    message_template.encode('utf-8'),
                    description.encode('utf-8'),
                    params_json.encode('utf-8')
                )
            else:
                self.lib.RegisterRoute(
                    path.encode('utf-8'),
                    method.value.encode('utf-8'),
                    message_template.encode('utf-8'),
                    description.encode('utf-8')
                )
            
            # Register dependencies if provided
            if dependencies:
                for dep in dependencies:
                    if isinstance(dep, Depends):
                        # Register the dependency function
                        dep_name = dep.dependency.__name__
                        dep_value = dep.dependency()
                        if isinstance(dep_value, str):
                            self._dependencies[dep_name] = dep_value
                            self.dependency(dep_name, dep_value)
                    elif isinstance(dep, str):
                        # Dependency name is provided directly
                        if dep not in self._dependencies:
                            raise ValueError(f"Dependency '{dep}' not registered")
            
            return func
        return decorator
    
    def get(self, path: str, **kwargs):
        """Shortcut for route with GET method"""
        return self.route(path, method=HTTPMethod.GET, **kwargs)
    
    def post(self, path: str, **kwargs):
        """Shortcut for route with POST method"""
        return self.route(path, method=HTTPMethod.POST, **kwargs)
    
    def put(self, path: str, **kwargs):
        """Shortcut for route with PUT method"""
        return self.route(path, method=HTTPMethod.PUT, **kwargs)
    
    def delete(self, path: str, **kwargs):
        """Shortcut for route with DELETE method"""
        return self.route(path, method=HTTPMethod.DELETE, **kwargs)
    
    def patch(self, path: str, **kwargs):
        """Shortcut for route with PATCH method"""
        return self.route(path, method=HTTPMethod.PATCH, **kwargs)

    def middleware(self, name: str, enabled: bool = True):
        """Register middleware by name"""
        self.lib.RegisterMiddleware(name.encode('utf-8'), c_int(1 if enabled else 0))

    def dependency(self, name: str, value: str):
        """Register a dependency for injection"""
        self._dependencies[name] = value
        self.lib.RegisterDependency(name.encode('utf-8'), value.encode('utf-8'))
    
    def set_log_level(self, level: LogLevel):
        """Set the logging level"""
        self.lib.SetLogLevel(level.value.encode('utf-8'))
    
    def include_debug_data(self, enabled: bool = False):
        """Control whether to include debug data (params, body) in responses"""
        self.lib.SetIncludeDebugData(c_int(1 if enabled else 0))
    
    def configure_server(self, read_timeout: int = 5000, write_timeout: int = 10000, 
                        idle_timeout: int = 30000, max_body_size: int = 4*1024*1024,
                        concurrency: int = 256*1024):
        """Configure server parameters for optimal performance"""
        self.lib.SetServerConfig(
            c_size_t(read_timeout),
            c_size_t(write_timeout),
            c_size_t(idle_timeout),
            c_size_t(max_body_size),
            c_size_t(concurrency)
        )

    def start(self):
        """Start the FastPaze server"""
        self.lib.StartServer()