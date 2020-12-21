# ASCII

ASCII is a rest-based service that allows users to transform an PNG image into ASCII art. In hindsight I should have picked a better name for this project. This project was written as a coding exercise.

## API
The API is simple.
  1. **Create a new ASCII image: `POST /images`**
  - Method: POST
  - Body: binary representation of a PNG image
  - Headers: 
    - *[experimental]* `async: bool (optional, default = false)`. See notes for explanation
  - Response: a uuid string associated with the ASCII image
  - Notes: 
    - returns 400 if not a valid PNG image
    - this endpoint does not return the image itself but rather the uuid for the image resource
    - the default behaviour of the endpoint is to return the uuid of the ascii image resource when the ascii image has finished generating. Thus a successful return means the ascii image is ready to be fetched.
    - *[experimental]* if the async header value is set to true, the endpoint will instead return as soon as an uuid for the image is generated. The image itself could still be generating. Use the GET endpoint to fetch its status/value.
  2. **Fetch an existing/creating ASCII image: `GET /images/{uuid}`**
  - Method: GET
  - Url Param: `uuid: uuid of the image from the Create endpoint`
  - Response: 
    - `status: string {finished/generating/error}`
    - `error: string`
    - `asciiData: string`
  - Notes:
    - returns 404 if the uuid is not an existing resource
  3. **List all ASCII images: `GET /images`**
  - Method: GET
  - Response: `[]uuid`
  - Notes:
    - will only return list of completed images right now

## Implementation Details
Aside from the API's functional specs I also focused on adding some bells and whistles to make this code-base more representative of an actual service I'd deploy to production
  - Logging
    - I use logrus which outputs logs in a defined logging format (i.e json). 
    This is important in a production environment if we want the logs to get scraped to a log aggregator stack.
  - API tracking 
    - Each api request should be trackable via a correlationID. The correlationID for each request will be included in the error response so customers can refer to them for support.
    Instead of manually setting and inserting correlationIDs into logs/responses in the main application code, 
    I wrote a middleware (in cmd/server/middleware.go) that auto-decorates and auto-injects a correlationID.
    The middleware also has naive API latency tracking capabilities (logs request start and request end).
  - Request timeouts
    - Not entirely applicable for this particular spec but sometimes the server itself may want to impose a strict deadline for each request.
    I also added a dynamic timeout middleware that can dynamically adjust the timeout per request based on the request itself (i.e larger images => longer timeouts)
  - Async workflows
    - If the ascii conversion is cpu/io-bound then it doesn't make sense for the initial POST call to block until the entire processing has finished. Thus I added extra functionality for the POST endpoint to be async. The request immediately returns an uuid while the processing happens async. 
    The caller can use the GET endpoint to poll for status.
    - I marked this async version as experimental because it uses my own implementation of async Tasks in Go (https://github.com/eriksywu/go-async) that has not been 100% production-tested. 
  - Why timeouts and async?
    - I believe long-living TCP connections breaks the implied contract/behaviour for REST APIs. There could also be too many things that go wrong. For example, certain go REST libraries do not handle tcp resets all that well - which most L3 loadbalancers rely on to keep NAT ports open. 
    - Use websockets or grpc if we want to maintain a long-living TCP connection.
    