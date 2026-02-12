# [Short Title, e.g., ADR-002 - Use Redis for Session Caching]

## Status
* **Status:** Accepted / Proposed / Deprecated
* **Date:** YYYY-MM-DD

## Context
What is the issue that we are seeing that is motivating this decision?
Describe the technical constraints, business requirements, or pain points.
*Example: The database is experiencing high latency during peak traffic because of repetitive read operations.*

## Decision
What is the change that we are proposing and/or doing?
*Example: We will implement Redis (AWS ElastiCache) as a write-through cache for user session data.*

## Consequences
What becomes easier or more difficult because of this change? (The Trade-offs).

* **Positive:** Reduced load on the primary PostgreSQL database. Faster response times for users.
* **Negative:** Introduces a new infrastructure component to maintain.
* **Risks:** Potential data inconsistency (stale data) if the cache invalidation fails.