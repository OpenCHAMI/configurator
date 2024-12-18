## 
## Run these tests after starting server with `configurator serve...`
##

# Generate a `example` config with default plugin and template
GET http://127.0.0.1:3334/generate?target=example
HTTP 200

# Create a new target using the API
POST http://127.0.0.1:3334/targets
{
    "name": "test",
    "plugin": "example",
    "templates": [{
        "contents": "This is an example template used with the example plugin."
    }]
}
HTTP 200

# Test the new target just add from POST above
GET http://127.0.0.1:3334/generate?target=example
HTTP 200