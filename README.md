# API Rate Limiter

This GitHub repository implements a server-side API rate limiter designed to handle a large number of requests in a distributed environment. The rate limiter is flexible, allowing different sets of throttle rules and can be implemented as a separate service or within application code. The goal is to accurately limit excessive requests with low latency, minimal memory usage, and high fault tolerance.

## Requirements

1. **Accurate Rate Limiting**: The system should accurately limit excessive requests.
2. **Low Latency**: The rate limiter should not significantly slow down HTTP response times.
3. **Memory Efficiency**: Use as little memory as possible.
4. **Distributed Rate Limiting**: Support rate limiting across multiple servers or processes.
5. **Exception Handling**: Provide clear exceptions to users when their requests are throttled.
6. **High Fault Tolerance**: Problems with the rate limiter should not affect the entire system.

## Algorithms for Rate Limiting

The repository evaluates different rate limiting algorithms, considering their pros and cons, and chooses the most suitable one from the following options:
- Token Bucket
- Leaking Bucket
- Fixed Window Counter
- Sliding Window Log
- Sliding Window Counter

## Rate Limiting Rules

The rate limiter uses a configuration file to define rules. An example rule for the 'auth' domain is provided:
```yaml
domain: auth
descriptors:
  - key: auth_type
    value: login
    rate_limit:
      unit: minute
      requests_per_unit: 5
```

## Exceeding the Rate Limit

If a request exceeds the rate limit, the API returns an HTTP response code 429 (Too Many Requests) to the client. Depending on use cases, rate-limited requests may be enqueued for later processing.

## Rate Limiter Headers

Clients can determine if they are being throttled and the remaining allowed requests through HTTP response headers:
- **X-Ratelimit-Remaining**: The remaining number of allowed requests within the window.
- **X-Ratelimit-Limit**: Indicates the maximum calls the client can make per time window.
- **X-Ratelimit-Retry-After**: Number of seconds to wait until the client can make a request again without being throttled.

## Detailed Design

Rules are stored on disk, and workers pull them into cache. The rate limiter middleware, upon receiving a request, loads rules from the cache and fetches counters and last request timestamps from Redis cache. Based on this information, the rate limiter decides whether to forward the request to API servers or return a 429 Too Many Requests error to the client. Workers enable parallel and concurrent processing, ensuring scalability.

## Rate Limiter in a Distributed Environment

Scaling the rate limiter for multiple servers involves addressing race conditions and synchronization issues. The provided solution involves using Lua scripts to prevent race conditions without significant system slowdown.

## Monitoring

After implementation, it is crucial to monitor the rate limiter's effectiveness. Analytics data should be gathered to ensure the rate limiting algorithm and rules are achieving the desired outcomes. Adjustments can be made if rules are too strict or if the algorithm is ineffective during sudden traffic increases, such as flash sales.

**Note**: This README provides a high-level overview. For detailed implementation instructions, please refer to the documentation and source code in the repository.
