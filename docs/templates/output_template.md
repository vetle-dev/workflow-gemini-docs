Structure your response exactly like this:

# [Application/Module Name]

## 🏗️ Architecture (C4 Model)
Use the Mermaid C4 syntax to create a specific **Container Diagram**.
Focus on the logical containers (API, Database, Workers) and how data flows between them.

```mermaid
[Insert Mermaid code here]
```

## 🔌 Integrations & Data Flow
Analyze code (e.g., HTTP clients, SQL drivers, queues) to map dependencies.

| Direction | System/Service | Protocol | Purpose | Auth Method |
|--|--|--|--|--|
| **Downstream** | e.g. Stripe API | HTTPS/REST | Payment processing | API Key |
| **Upstream** | e.g. Frontend BFF | gRPC | Fetches user profile | mTLS |
| **Internal** | e.g. Postgres DB | TCP/SQL | Persistent storage | User/Pass |

## ⚙️ Key Configuration & Behavior
Identify environment variables or flags that control behavior (not just standard infra settings).

| Environment Variable | Description | Criticality |
|--|--|--|
| `FEATURE_X_ENABLED` | Enables the new checkout | High |
| `LOG_LEVEL` | Controls verbosity (debug/info) | Low |
| `RETRY_LIMIT` | Max retries for HTTP calls | Medium |

## 🔒 Security Posture
- **Authentication**: [Briefly describe how the app authenticates users/services]
- **Authorization**: [Briefly describe scopes/roles]
- **Data Privacy**: [Does it handle PII/GDPR data? Based on variable names like 'ssn', 'email', etc.]