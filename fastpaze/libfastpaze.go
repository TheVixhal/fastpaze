package main

import (
	"C"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
)

// ApiResponse for JSON response
type ApiResponse struct {
	Message string      `json:"message" validate:"required"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse for structured error responses
type ErrorResponse struct {
	Error       string `json:"error" validate:"required"`
	StatusCode  int    `json:"status_code" validate:"required"`
	Description string `json:"description,omitempty"`
}

// RouteInfo stores route metadata for OpenAPI and handling
type RouteInfo struct {
	Path        string
	Method      string
	Message     string // Store the message directly
	Description string
	Parameters  []ParameterInfo
	RequestBody *RequestBodyInfo
	Responses   map[int]string
	// Path parameter regex for matching
	PathRegex    *regexp.Regexp
	ParamNames   []string
	Dependencies []string // Names of dependencies this route requires
}

// RequestBodyInfo for OpenAPI documentation
type RequestBodyInfo struct {
	Description string `json:"description"`
	Required    bool   `json:"required"`
	ContentType string `json:"content_type"` // e.g., "application/json"
	Schema      string `json:"schema"`       // JSON schema or reference
}

// ParameterInfo for OpenAPI documentation
type ParameterInfo struct {
	Name        string `json:"name"`
	In          string `json:"in"` // e.g., "query", "path", "header"
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Schema      string `json:"schema,omitempty"` // JSON schema for complex types
}

// OpenAPI structure for API documentation
type OpenAPI struct {
	OpenAPI    string                            `json:"openapi"`
	Info       map[string]string                 `json:"info"`
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components map[string]interface{}            `json:"components"`
}

// Global variables with thread-safe access
var (
	routes           = make(map[string]RouteInfo)
	routesMu         sync.RWMutex
	validate         = validator.New()
	middlewares      = []func(fasthttp.RequestHandler) fasthttp.RequestHandler{}
	middlewaresMu    sync.RWMutex
	dependencies     = make(map[string]interface{})
	depsMu           sync.RWMutex
	logLevel         = "info" // Default log level
	includeDebugData = false  // Don't include debug data in responses by default

	// Server configuration
	serverConfig = ServerConfig{
		ReadTimeout:        5 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        30 * time.Second,
		MaxRequestBodySize: 4 * 1024 * 1024, // 4MB
		Concurrency:        256 * 1024,      // 256K concurrent connections
		TCPKeepAlive:       true,
		ReduceMemoryUsage:  true,
		Port:               ":8080", // Default port
	}
)

// ServerConfig holds tunable server parameters
type ServerConfig struct {
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxRequestBodySize int
	Concurrency        int
	TCPKeepAlive       bool
	ReduceMemoryUsage  bool
	Port               string // Port to listen on (e.g., ":8080")
}

//export SetIncludeDebugData
func SetIncludeDebugData(cEnabled int) {
	enabled := cEnabled != 0
	includeDebugData = enabled
	logInfo("Include debug data in responses: %v", enabled)
}

// Middleware registration
//
//export RegisterMiddleware
func RegisterMiddleware(cName uintptr, cEnabled int) {
	namePtr := (*C.char)(unsafe.Pointer(cName))
	if namePtr == nil {
		log.Println("Error: cName is nil in RegisterMiddleware")
		return
	}
	name := C.GoString(namePtr)

	enabled := cEnabled != 0
	if !enabled {
		logDebug("Middleware %s is disabled", name)
		return
	}

	middlewaresMu.Lock()
	defer middlewaresMu.Unlock()
	switch name {
	case "logging":
		middlewares = append(middlewares, loggingMiddleware)
		logInfo("Registered middleware: %s", name)
	case "cors":
		middlewares = append(middlewares, corsMiddleware)
		logInfo("Registered middleware: %s", name)
	case "rate_limiter":
		middlewares = append(middlewares, rateLimiterMiddleware)
		logInfo("Registered middleware: %s", name)
	default:
		logWarning("Unknown middleware: %s", name)
	}
}

// Logging middleware
func loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		logInfo("%s %s from %s in %v - Status: %d",
			string(ctx.Method()), string(ctx.Path()),
			ctx.RemoteAddr(), time.Since(start), ctx.Response.StatusCode())
	}
}

// CORS middleware
func corsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		next(ctx)
	}
}

// Rate limiter middleware (simple implementation)
func rateLimiterMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	var (
		mu         sync.Mutex
		clients    = make(map[string][]time.Time)
		maxReqs    = 100              // Max requests per window
		windowSize = 60 * time.Second // Window size (1 minute)
	)

	return func(ctx *fasthttp.RequestCtx) {
		ip := ctx.RemoteAddr().String()

		mu.Lock()
		// Clean up old timestamps
		now := time.Now()
		if _, exists := clients[ip]; exists {
			var validTimes []time.Time
			for _, t := range clients[ip] {
				if now.Sub(t) < windowSize {
					validTimes = append(validTimes, t)
				}
			}
			clients[ip] = validTimes
		}

		// Check rate limit
		if len(clients[ip]) >= maxReqs {
			mu.Unlock()
			ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
			ctx.SetBodyString(`{"error":"Rate limit exceeded","status_code":429,"description":"Too many requests, please try again later"}`)
			return
		}

		// Add current request timestamp
		clients[ip] = append(clients[ip], now)
		mu.Unlock()

		next(ctx)
	}
}

// Dependency injection context (e.g., for auth or DB)
//
//export RegisterDependency
func RegisterDependency(cName uintptr, cValue uintptr) {
	namePtr := (*C.char)(unsafe.Pointer(cName))
	valuePtr := (*C.char)(unsafe.Pointer(cValue))
	if namePtr == nil || valuePtr == nil {
		logError("One or more parameters are nil in RegisterDependency")
		return
	}
	name := C.GoString(namePtr)
	value := C.GoString(valuePtr)
	depsMu.Lock()
	dependencies[name] = value
	depsMu.Unlock()
	logInfo("Registered dependency: %s", name)
}

// GetDependency retrieves a dependency by name
func GetDependency(name string) (interface{}, bool) {
	depsMu.RLock()
	defer depsMu.RUnlock()
	val, exists := dependencies[name]
	return val, exists
}

// Set log level
//
//export SetLogLevel
func SetLogLevel(cLevel uintptr) {
	levelPtr := (*C.char)(unsafe.Pointer(cLevel))
	if levelPtr == nil {
		return
	}
	level := C.GoString(levelPtr)
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if validLevels[level] {
		logLevel = level
		logInfo("Log level set to: %s", level)
	} else {
		logWarning("Invalid log level: %s. Using default: %s", level, logLevel)
	}
}

// Logger functions with levels
func logDebug(format string, args ...interface{}) {
	if logLevel == "debug" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel == "debug" || logLevel == "info" {
		log.Printf("[INFO] "+format, args...)
	}
}

func logWarning(format string, args ...interface{}) {
	if logLevel == "debug" || logLevel == "info" || logLevel == "warn" {
		log.Printf("[WARN] "+format, args...)
	}
}

func logError(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// Convert path with parameters to regex pattern
func pathToRegex(path string) (*regexp.Regexp, []string) {
	paramNames := []string{}
	// Match {param} pattern
	paramPattern := regexp.MustCompile(`\{([^{}]+)\}`)

	// Find all parameter names
	matches := paramPattern.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		if len(match) > 1 {
			paramNames = append(paramNames, match[1])
		}
	}

	// Replace {param} with regex capture group
	regexPattern := paramPattern.ReplaceAllString(path, `([^/]+)`)
	// Escape forward slashes and add start/end anchors
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		logError("Failed to compile regex for path %s: %v", path, err)
		return nil, paramNames
	}

	return regex, paramNames
}

//export RegisterRoute
func RegisterRoute(cPath uintptr, cMethod uintptr, cMessage uintptr, cDesc uintptr) {
	pathPtr := (*C.char)(unsafe.Pointer(cPath))
	methodPtr := (*C.char)(unsafe.Pointer(cMethod))
	messagePtr := (*C.char)(unsafe.Pointer(cMessage))
	descPtr := (*C.char)(unsafe.Pointer(cDesc))

	if pathPtr == nil || methodPtr == nil || messagePtr == nil || descPtr == nil {
		logError("One or more parameters are nil in RegisterRoute")
		return
	}

	path := C.GoString(pathPtr)
	method := strings.ToUpper(C.GoString(methodPtr))
	message := C.GoString(messagePtr)
	desc := C.GoString(descPtr)

	logInfo("Registering route: %s for method: %s with message: %s", path, method, message)

	// Convert path with parameters to regex
	pathRegex, paramNames := pathToRegex(path)

	routesMu.Lock()
	key := path + method
	routes[key] = RouteInfo{
		Path:        path,
		Method:      method,
		Message:     message,
		Description: desc,
		Parameters:  []ParameterInfo{},
		PathRegex:   pathRegex,
		ParamNames:  paramNames,
		Responses: map[int]string{
			200: "Successful response",
			400: "Bad request",
			404: "Not found",
			500: "Internal server error",
		},
	}
	logInfo("Route registered with key: %s", key)
	routesMu.Unlock()
}

//export RegisterRouteWithParams
func RegisterRouteWithParams(cPath uintptr, cMethod uintptr, cMessage uintptr, cDesc uintptr, cParamsJSON uintptr) {
	pathPtr := (*C.char)(unsafe.Pointer(cPath))
	methodPtr := (*C.char)(unsafe.Pointer(cMethod))
	messagePtr := (*C.char)(unsafe.Pointer(cMessage))
	descPtr := (*C.char)(unsafe.Pointer(cDesc))
	paramsJSONPtr := (*C.char)(unsafe.Pointer(cParamsJSON))

	if pathPtr == nil || methodPtr == nil || messagePtr == nil || descPtr == nil || paramsJSONPtr == nil {
		logError("One or more parameters are nil in RegisterRouteWithParams")
		return
	}

	path := C.GoString(pathPtr)
	method := strings.ToUpper(C.GoString(methodPtr))
	message := C.GoString(messagePtr)
	desc := C.GoString(descPtr)
	paramsJSON := C.GoString(paramsJSONPtr)

	var params []ParameterInfo
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		logError("Error parsing params JSON: %v", err)
		return
	}

	logInfo("Registering route with params: %s for method: %s", path, method)

	// Convert path with parameters to regex
	pathRegex, paramNames := pathToRegex(path)

	routesMu.Lock()
	key := path + method
	routes[key] = RouteInfo{
		Path:        path,
		Method:      method,
		Message:     message,
		Description: desc,
		Parameters:  params,
		PathRegex:   pathRegex,
		ParamNames:  paramNames,
		Responses: map[int]string{
			200: "Successful response",
			400: "Bad request",
			404: "Not found",
			500: "Internal server error",
		},
	}
	logInfo("Route registered with key: %s and %d parameters", key, len(params))
	routesMu.Unlock()
}

// Find matching route with path parameters
func findMatchingRoute(path string, method string) (RouteInfo, map[string]string, bool) {
	pathParams := make(map[string]string)

	// First try exact match
	routesMu.RLock()
	defer routesMu.RUnlock()

	key := path + method
	if route, exists := routes[key]; exists {
		return route, pathParams, true
	}

	// If no exact match, try regex matching for routes with path parameters
	for _, route := range routes {
		if route.Method == method && route.PathRegex != nil {
			matches := route.PathRegex.FindStringSubmatch(path)
			if len(matches) > 1 {
				// Extract path parameters
				for i, name := range route.ParamNames {
					if i+1 < len(matches) {
						pathParams[name] = matches[i+1]
					}
				}
				return route, pathParams, true
			}
		}
	}

	return RouteInfo{}, pathParams, false
}

// ServeOpenAPI generates the OpenAPI JSON
func ServeOpenAPI(ctx *fasthttp.RequestCtx) {
	openapi := OpenAPI{
		OpenAPI: "3.0.0",
		Info: map[string]string{
			"title":       "FastPaze API",
			"version":     "1.0.0",
			"description": "High-performance API built with FastPaze and Go",
		},
		Paths:      make(map[string]map[string]interface{}),
		Components: make(map[string]interface{}),
	}

	routesMu.RLock()
	for _, route := range routes {
		if _, exists := openapi.Paths[route.Path]; !exists {
			openapi.Paths[route.Path] = make(map[string]interface{})
		}

		// Build parameters array for OpenAPI
		parameters := make([]map[string]interface{}, 0)
		for _, param := range route.Parameters {
			parameters = append(parameters, map[string]interface{}{
				"name":        param.Name,
				"in":          param.In,
				"description": param.Description,
				"required":    param.Required,
				"schema":      map[string]string{"type": param.Type},
			})
		}

		// Add the route to OpenAPI paths
		openapi.Paths[route.Path][strings.ToLower(route.Method)] = map[string]interface{}{
			"summary":    route.Description,
			"parameters": parameters,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": route.Responses[200],
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"message": map[string]string{"type": "string"},
									"data":    map[string]string{"type": "object"},
								},
							},
						},
					},
				},
			},
		}
	}
	routesMu.RUnlock()

	ctx.Response.Header.Set("Content-Type", "application/json")
	if err := json.NewEncoder(ctx).Encode(openapi); err != nil {
		ctx.Error(`{"error": "Failed to generate OpenAPI"}`, fasthttp.StatusInternalServerError)
	}
}

// Handler for serving Swagger UI
func ServeSwaggerUI(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())

	// Default to index.html if the path is just /swagger/ or /swagger
	if path == "/swagger/" || path == "/swagger" {
		path = "/swagger/index.html"
	}

	// Remove /swagger prefix to get the file path
	filePath := strings.TrimPrefix(path, "/swagger")
	if filePath == "" {
		filePath = "/index.html"
	}

	// Serve the file from the swagger-ui directory
	ctx.SendFile("swagger-ui" + filePath)
}

// Parse request parameters
func parseRequestParams(ctx *fasthttp.RequestCtx, route RouteInfo, pathParams map[string]string) map[string]string {
	params := make(map[string]string)

	// Add path parameters
	for k, v := range pathParams {
		params[k] = v
	}

	// Parse query parameters
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})

	// Parse header parameters
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		// Convert to lowercase for case-insensitive matching
		params["header:"+strings.ToLower(headerKey)] = string(value)
	})

	return params
}

// Process a message template with path parameters
func processMessageTemplate(message string, pathParams map[string]string) string {
	// Check if message is a JSON string
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(message), &jsonData); err == nil {
		// Process each field in the JSON recursively
		processJSONFields(jsonData, pathParams)

		// Convert back to JSON
		result, err := json.Marshal(jsonData)
		if err == nil {
			return string(result)
		}
	}

	// If not JSON or JSON processing failed, do simple string replacement
	result := message
	for k, v := range pathParams {
		placeholder := "{" + k + "}"
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}

// Recursively process JSON fields to replace placeholders
func processJSONFields(data map[string]interface{}, pathParams map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Replace placeholders in string values
			for paramName, paramValue := range pathParams {
				placeholder := "{" + paramName + "}"
				if strings.Contains(v, placeholder) {
					data[key] = strings.ReplaceAll(v, placeholder, paramValue)
				}
			}
		case map[string]interface{}:
			// Recursively process nested objects
			processJSONFields(v, pathParams)
		case []interface{}:
			// Process arrays
			for i, item := range v {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					processJSONFields(nestedMap, pathParams)
				} else if strItem, ok := item.(string); ok {
					// Replace placeholders in string array items
					for paramName, paramValue := range pathParams {
						placeholder := "{" + paramName + "}"
						if strings.Contains(strItem, placeholder) {
							v[i] = strings.ReplaceAll(strItem, placeholder, paramValue)
						}
					}
				}
			}
		}
	}
}

// Main request handler using fasthttp
func requestHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	// Handle OpenAPI and Swagger UI routes
	if path == "/openapi.json" {
		ServeOpenAPI(ctx)
		return
	}

	if strings.HasPrefix(path, "/swagger") {
		ServeSwaggerUI(ctx)
		return
	}

	// Find matching route (including path parameter handling)
	route, pathParams, exists := findMatchingRoute(path, method)

	if !exists {
		// Check if the path exists with a different method
		var supportedMethods []string
		routesMu.RLock()
		for _, rt := range routes {
			// Check both exact path and regex path
			if rt.Path == path || (rt.PathRegex != nil && rt.PathRegex.MatchString(path)) {
				supportedMethods = append(supportedMethods, rt.Method)
			}
		}
		routesMu.RUnlock()

		errorMsg := fmt.Sprintf("Route not found for %s %s", method, path)
		if len(supportedMethods) > 0 {
			errorMsg = fmt.Sprintf("%s - Try using method %s", errorMsg, strings.Join(supportedMethods, " or "))
		}

		logWarning("Route not found: %s %s", method, path)

		// Return error in appropriate format
		errResp := ErrorResponse{
			Error:       errorMsg,
			StatusCode:  fasthttp.StatusNotFound,
			Description: "The requested endpoint does not exist",
		}

		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		ctx.Response.Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx).Encode(errResp)
		return
	}

	logDebug("Route found for %s %s, processing request", method, path)

	// Parse request parameters
	params := parseRequestParams(ctx, route, pathParams)

	// Parse request body if present
	var requestBody map[string]interface{}
	if len(ctx.Request.Body()) > 0 {
		contentType := string(ctx.Request.Header.Peek("Content-Type"))
		if strings.Contains(contentType, "application/json") {
			if err := json.Unmarshal(ctx.Request.Body(), &requestBody); err != nil {
				logWarning("Failed to parse request body: %v", err)
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				json.NewEncoder(ctx).Encode(ErrorResponse{
					Error:       "Invalid JSON body",
					StatusCode:  fasthttp.StatusBadRequest,
					Description: err.Error(),
				})
				return
			}
		}
	}

	// Process the message template with path parameters
	processedMessage := processMessageTemplate(route.Message, pathParams)

	// Check if the processed message is valid JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(processedMessage), &responseData); err == nil {
		// If it's valid JSON, use it directly as the response
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(processedMessage)
		return
	}

	// If not valid JSON, use the standard ApiResponse format
	response := ApiResponse{
		Message: processedMessage,
	}

	// Add parameters to response data if debug data inclusion is enabled
	if includeDebugData {
		debugData := make(map[string]interface{})
		if len(params) > 0 {
			debugData["params"] = params
		}
		if requestBody != nil {
			debugData["body"] = requestBody
		}
		if len(debugData) > 0 {
			response.Data = debugData
		}
	} else {
		// Always include path parameters in the response data
		if len(pathParams) > 0 {
			response.Data = pathParams
		}
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	if err := json.NewEncoder(ctx).Encode(response); err != nil {
		logError("Error encoding response: %v", err)
		ctx.Error(`{"error": "Internal server error"}`, fasthttp.StatusInternalServerError)
		return
	}
}

// Apply middleware to handler
func applyMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	middlewaresMu.RLock()
	defer middlewaresMu.RUnlock()

	result := handler
	// Apply middleware in reverse order (last registered is executed first)
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}
	return result
}

//export StartServer
func StartServer() {
	// Create a channel to signal shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Apply middleware to the main handler
	handler := applyMiddleware(requestHandler)

	// Configure the server
	server := &fasthttp.Server{
		Handler:                       handler,
		ReadTimeout:                   serverConfig.ReadTimeout,
		WriteTimeout:                  serverConfig.WriteTimeout,
		IdleTimeout:                   serverConfig.IdleTimeout,
		MaxRequestBodySize:            serverConfig.MaxRequestBodySize,
		Concurrency:                   serverConfig.Concurrency,
		TCPKeepalive:                  serverConfig.TCPKeepAlive,
		ReduceMemoryUsage:             serverConfig.ReduceMemoryUsage,
		DisableHeaderNamesNormalizing: true, // For better performance
	}

	// Start the server in a goroutine
	go func() {
		logInfo("FastPaze server running on http://localhost%s", serverConfig.Port)
		logInfo("API docs available at http://localhost%s/swagger/", serverConfig.Port)
		if err := server.ListenAndServe(serverConfig.Port); err != nil {
			logError("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stop
	logInfo("Shutting down server...")

	// Shutdown the server
	if err := server.Shutdown(); err != nil {
		logError("Server shutdown error: %v", err)
	}

	logInfo("Server stopped gracefully")
}

//export SetServerConfig
func SetServerConfig(cReadTimeout uintptr, cWriteTimeout uintptr, cIdleTimeout uintptr, cMaxBodySize uintptr, cConcurrency uintptr, cPort uintptr) {
	readTimeout := int64(cReadTimeout)
	writeTimeout := int64(cWriteTimeout)
	idleTimeout := int64(cIdleTimeout)
	maxBodySize := int(cMaxBodySize)
	concurrency := int(cConcurrency)
	portPtr := (*C.char)(unsafe.Pointer(cPort))

	if readTimeout > 0 {
		serverConfig.ReadTimeout = time.Duration(readTimeout) * time.Millisecond
	}
	if writeTimeout > 0 {
		serverConfig.WriteTimeout = time.Duration(writeTimeout) * time.Millisecond
	}
	if idleTimeout > 0 {
		serverConfig.IdleTimeout = time.Duration(idleTimeout) * time.Millisecond
	}
	if maxBodySize > 0 {
		serverConfig.MaxRequestBodySize = maxBodySize
	}
	if concurrency > 0 {
		serverConfig.Concurrency = concurrency
	}
	if portPtr != nil {
		port := C.GoString(portPtr)
		if port != "" {
			serverConfig.Port = port
		}
	}

	logInfo("Server config updated: ReadTimeout=%v, WriteTimeout=%v, IdleTimeout=%v, MaxBodySize=%d, Concurrency=%d, Port=%s",
		serverConfig.ReadTimeout, serverConfig.WriteTimeout, serverConfig.IdleTimeout,
		serverConfig.MaxRequestBodySize, serverConfig.Concurrency, serverConfig.Port)
}

func main() {
	// Required for shared library
}