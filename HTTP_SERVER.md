# Sentra HTTP Server

This document describes how to use the Sentra HTTP server for rule execution.

## Overview

The HTTP server provides a REST API for executing Sentra rules. It accepts rule execution requests via HTTP POST requests and returns the decision and any attachments.

## Language Features

Sentra supports a rich set of expressions for policy evaluation, including:

- **Arithmetic operations**: `+`, `-`, `*`, `/`, `%`
- **Logical operations**: `and`, `or`, `xor`, `not`
- **Comparison operations**: `==`, `!=`, `<`, `<=`, `>`, `>=`
- **Collection operations**: `in`, `contains`, `matches`
- **Quantifier operations**: `any`, `all`, `filter`, `map`, `distinct`, `reduce`
- **Count operation**: `count` - returns the length of lists, maps, or strings
- **Type checking**: `is defined`, `is empty`, `is not defined`, `is not empty`
- **String operations**: pattern matching with `matches`
- **Collection literals**: `[1, 2, 3]`, `{"key": "value"}`
- **Conditional expressions**: `condition ? true_value : false_value`

## API Endpoints

### Decision Execution

**POST** `/decision/{namespace}/{policy}/{rule}?{runconfig_params}`

Executes a specific rule with the provided facts.

#### Path Parameters

- `namespace`: The namespace containing the policy (can contain multiple segments separated by '/')
- `policy`: The policy name (second to last segment in the path)
- `rule`: The rule name to execute (last segment in the path)

#### Path Structure

The path follows this pattern: `/decision/{namespace}/{policy}/{rule}` where:

- The **last segment** is the rule name
- The **second to last segment** is the policy name
- **Everything before that** is the namespace (can contain multiple `/`-separated segments)

#### Examples

- `/decision/sh/sentra/auth/v1/user/allow` → namespace: `sh/sentra/auth/v1`, policy: `user`, rule: `allow`
- `/decision/org/department/team/policy/rule` → namespace: `org/department/team`, policy: `policy`, rule: `rule`

#### Query Parameters

- `runconfig_params`: Optional configuration parameters for rule execution

#### Request Body

```json
{
  "facts": {
    "key1": "value1",
    "key2": {
      "nested": "object"
    }
  }
}
```

#### Response

```json
{
  "decision": "allow",
  "attachments": {
    "reason": "User has required permissions"
  }
}
```

#### Error Response

```json
{
  "error": "Rule execution failed",
  "message": "Detailed error message"
}
```

### Health Check

**GET** `/health`

Returns the server health status.

#### Response

```json
{
  "status": "healthy",
  "time": "2025-01-27T10:30:00Z"
}
```

## Running the Server

### Command Line Options

```bash
go run main.go serve [options]
```

Options:

- `port`: Port to run the HTTP server on (default: 7529)
- `pack-location`: Path to the policy pack directory (default: ./example_pack)
- `listen`: Address(es) to listen on. Can specify multiple addresses. Special values:
  - `local` → `localhost:port`
  - `local4` → `127.0.0.1:port`
  - `local6` → `[::1]:port`
  - `network` → `:port` (all interfaces)
  - `network4` → `0.0.0.0:port` (IPv4 only)
  - `network6` → `[::]:port` (IPv6 only)
  - Any other value used as-is

### Example

```bash
# Start server with default settings (port 7529, localhost, ./example_pack)
go run main.go serve

# Start server on port 3000 with custom policy pack
go run main.go serve --port 3000 --pack-location /path/to/policy/pack

# Start server on localhost (IPv4/IPv6)
go run main.go serve --listen "local" --port 8080

# Start server on localhost IPv4 only
go run main.go serve --listen "local4" --port 8080

# Start server on localhost IPv6 only
go run main.go serve --listen "local6" --port 8080

# Start server on all interfaces
go run main.go serve --listen "network" --port 8080

# Start server on all IPv4 interfaces
go run main.go serve --listen "network4" --port 8080

# Start server on all IPv6 interfaces
go run main.go serve --listen "network6" --port 8080

# Start server on custom address
go run main.go serve --listen "192.168.1.100:8080"

# Start server on multiple addresses
go run main.go serve --listen "local" --listen "network4" --port 8080

# Start server on both IPv4 and IPv6 localhost
go run main.go serve --listen "local4" --listen "local6" --port 8080
```

## Testing the API

### Using curl

```bash
# Basic rule execution with multi-segment namespace
curl -X POST http://localhost:7529/decision/sh/sentra/auth/v1/user/allow \
  -H "Content-Type: application/json" \
  -d '{
    "facts": {
      "user": {
        "name": "John",
        "age": 25,
        "admin": false,
        "scope": ["read", "write"]
      }
    }
  }'

# Another example with different namespace structure
curl -X POST http://localhost:7529/decision/org/department/team/policy/rule \
  -H "Content-Type: application/json" \
  -d '{
    "facts": {
      "context": {
        "department": "engineering",
        "team": "backend"
      }
    }
  }'

# Health check
curl http://localhost:7529/health
```

## Example Facts File

Create a JSON file with your facts:

```json
{
  "user": {
    "name": "Alice",
    "age": 30,
    "admin": true,
    "scope": ["read", "write", "admin"],
    "department": "engineering"
  },
  "request": {
    "resource": "/api/users",
    "method": "GET",
    "timestamp": "2025-01-27T10:30:00Z"
  },
  "data": {
    "items": ["item1", "item2", "item3"],
    "metadata": {
      "version": "1.0",
      "environment": "production"
    }
  }
}
```

### Example Policy Using Count Expression

Here's an example of how you might use the `count` expression in a policy:

```sentra
namespace example
policy access_control {
  let user_scopes = user.scope
  let scope_count = count user_scopes
  let data_items = data.items
  let item_count = count data_items

  rule allow = scope_count >= 2 and item_count > 0
}
```

This policy allows access only if the user has at least 2 scopes and there are items in the data.

## CORS Support

The server includes CORS headers to allow cross-origin requests:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: POST, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

## Error Handling

The server provides detailed error responses for various scenarios:

- **400 Bad Request**: Invalid path format or malformed JSON
- **405 Method Not Allowed**: Non-POST requests to decision endpoint
- **500 Internal Server Error**: Rule execution failures

## Graceful Shutdown

The server supports graceful shutdown on SIGINT or SIGTERM signals. It will:

1. Stop accepting new connections
2. Wait for existing requests to complete (up to 30 seconds)
3. Shutdown cleanly

## Security Considerations

- The server currently allows all origins (`*`) for CORS
- No authentication or authorization is implemented
- Consider adding these features for production use
- Validate and sanitize all inputs
- Implement rate limiting for production deployments
