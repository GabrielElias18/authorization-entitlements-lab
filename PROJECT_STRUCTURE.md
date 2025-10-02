# Project Structure: Authorization Architecture Lab

This document provides a detailed overview of the repository structure and architectural components for the authorization comparison project.

## 📁 Repository Layout

```
labs/
├── README.md                       # Main project documentation
├── COMPARISON.md                   # SpiceDB vs Neo4j implementation analysis
├── EXAMPLES.md                     # Resource authorization use cases & test scenarios
│
├── go-entitlement-service/         # Central gRPC API Gateway
│   ├── cmd/server/                 # Server entry point
│   │   └── main.go                 # Application bootstrap
│   ├── proto/                      # Protocol Buffer definitions
│   │   ├── entitlement.proto       # Service contract
│   │   ├── entitlement.pb.go       # Generated Go code
│   │   └── entitlement_grpc.pb.go  # Generated gRPC server/client
│   ├── internal/                   # Private implementation packages
│   │   ├── service/                # Business logic orchestration
│   │   │   └── service.go          # Entitlement service implementation
│   │   ├── spicedb/                # SpiceDB client adapter
│   │   │   └── client.go           # SpiceDB integration
│   │   ├── graphql/                # GraphQL client adapter
│   │   │   └── client.go           # GraphQL/PostgreSQL integration
│   │   └── neo4j/                  # Neo4j client adapter
│   │       └── client.go           # Neo4j integration
│   ├── go.mod                      # Go module dependencies
│   ├── test_grpc_api.sh            # gRPC API validation tests
│   └── test_permissions.sh         # Permission check tests
│
├── spicedb-model/                  # SpiceDB (Zanzibar) Implementation
│   ├── spicedb-config/             # Docker compose configuration
│   │   └── docker-compose.yml      # SpiceDB + PostgreSQL containers
│   ├── model.zaml                  # Authorization schema (Zanzibar)
│   ├── tuples.csv                  # Relationship data (user-resource mappings)
│   ├── schema.json                 # Schema validation
│   ├── setup.sh                    # Initialization script
│   ├── reset.sh                    # Clean state script
│   ├── load_relationships.sh       # Import relationship tuples
│   ├── run_permission_checks.sh    # Test suite runner
│   ├── tests/                      # Test scenarios
│   ├── Makefile                    # Build automation
│   └── README.md                   # SpiceDB setup guide
│
├── graphql-model/                  # GraphQL + PostgreSQL Implementation
│   ├── src/                        # Application source
│   │   ├── server.ts               # GraphQL server entry point
│   │   ├── entitlement-resolvers.ts # Permission check resolvers
│   │   └── generated/prisma/       # Prisma ORM types
│   ├── prisma/                     # Database configuration
│   │   └── schema.prisma           # Data model definition
│   ├── init-scripts/               # Database initialization
│   ├── schema.graphql              # GraphQL schema definition
│   ├── docker-compose.yml          # PostgreSQL container
│   ├── package.json                # Node.js dependencies
│   ├── tsconfig.json               # TypeScript configuration
│   ├── setup.sh                    # Database setup & migration
│   ├── reset.sh                    # Clean database state
│   ├── run_permission_checks.sh    # Test suite runner
│   └── README.md                   # GraphQL setup guide
│
├── graphdb-model/                  # Neo4j (GraphDB) Implementation
│   ├── neo4j-config/               # Docker compose configuration
│   │   └── docker-compose.yml      # Neo4j container
│   ├── schema.cypher               # Graph schema definition
│   ├── sample-data.cypher          # Test data (nodes & relationships)
│   ├── permission-queries.cypher   # Authorization query templates
│   ├── check-permission.ts         # Permission check implementation
│   ├── package.json                # Node.js dependencies
│   └── README.md                   # Neo4j setup guide
│
└── Test Orchestration Scripts
    ├── spicedb_canonical_tests.sh  # SpiceDB test runner (18 tests)
    ├── graphql_canonical_tests.sh  # GraphQL test runner (28 tests)
    └── neo4j_canonical_tests.sh    # Neo4j test runner (18 tests)
```

---

## 🏗️ Architecture Components

### 1. **go-entitlement-service** (Central API Gateway)
**Purpose:** Unified gRPC interface for all authorization backends

**Key Responsibilities:**
- Expose consistent permission check API via gRPC
- Route requests to SpiceDB, GraphQL, or Neo4j backends
- Abstract implementation details from clients
- Provide protocol buffer contracts for type safety

**Technology Stack:**
- **Language:** Go 1.21+
- **RPC Framework:** gRPC with Protocol Buffers
- **Architecture Pattern:** Adapter pattern for backend abstraction

**API Contract:**
```protobuf
service EntitlementService {
  rpc CheckPermission(PermissionRequest) returns (PermissionResponse);
}
```

---

### 2. **spicedb-model** (External Authorization Engine)
**Purpose:** Zanzibar-style dedicated authorization system

**Key Components:**
- `model.zaml` - Declarative authorization schema with caveats
- `tuples.csv` - Relationship data (subject-relation-object tuples)
- `setup.sh` - Automated schema migration and data loading

**Architecture Pattern:** External authorization service (Google Zanzibar)

**Authorization Model:**
- **Resources:** `resource`, `delegation`, `organization`
- **Relations:** `owner`, `delegated_access`, `auditor_access`, `member`
- **Permissions:** `can_initiate_transaction`, `can_view_records`, `can_access_document`
- **Caveats:** Time-bound validation, amount limits, date ranges

**Runtime Environment:**
- Docker container running SpiceDB
- PostgreSQL backend for persistence
- Port: `50051` (gRPC)

---

### 3. **graphql-model** (Application-Layer Authorization)
**Purpose:** Traditional application-integrated authorization with SQL

**Key Components:**
- `prisma/schema.prisma` - Relational data model (users, resources, delegations)
- `src/entitlement-resolvers.ts` - Business logic for permission checks
- GraphQL query layer for flexible API access

**Architecture Pattern:** Application-layer authorization

**Authorization Approach:**
- SQL queries with datetime/amount comparisons
- Transaction-based consistency (ACID)
- Prisma ORM for type-safe database access

**Runtime Environment:**
- Node.js + TypeScript server
- PostgreSQL database
- Apollo GraphQL server on port `4000`

---

### 4. **graphdb-model** (Database-Embedded Authorization)
**Purpose:** Graph-based authorization with Cypher queries

**Key Components:**
- `schema.cypher` - Graph node/relationship definitions
- `sample-data.cypher` - Test users, resources, and access paths
- `permission-queries.cypher` - Cypher templates for authorization checks
- `check-permission.ts` - TypeScript client for Neo4j

**Architecture Pattern:** Database-embedded authorization

**Authorization Approach:**
- Graph traversal for relationship-based permissions
- Cypher queries with WHERE clauses for business rules
- Visual relationship modeling

**Runtime Environment:**
- Neo4j Docker container
- Bolt protocol on port `7687`
- Neo4j Browser on port `7474`

---

## 🔄 Data Flow Architecture

### Permission Check Flow (via gRPC Service)

```
Client Application
       ↓
gRPC Request (CheckPermission)
       ↓
go-entitlement-service
       ↓
┌──────┴──────┬──────────────┬──────────────┐
│ SpiceDB     │ GraphQL      │ Neo4j        │
│ Adapter     │ Adapter      │ Adapter      │
└──────┬──────┴──────────────┴──────────────┘
       ↓              ↓              ↓
┌─────────────┐ ┌──────────┐ ┌─────────────┐
│  SpiceDB    │ │PostgreSQL│ │   Neo4j     │
│ (External)  │ │ (SQL)    │ │  (Cypher)   │
└─────────────┘ └──────────┘ └─────────────┘
```

### Direct Backend Access (Development/Testing)

```
Test Scripts
       ↓
┌──────┴──────┬──────────────┬──────────────┐
│ SpiceDB CLI │ GraphQL API  │ Neo4j Client │
│ (zed)       │ (HTTP POST)  │ (bolt://)    │
└─────────────┴──────────────┴──────────────┘
```

---

## 🧪 Test Infrastructure

### Test Suite Organization

| Test Script | Backend | Test Count | Coverage |
|------------|---------|------------|----------|
| `spicedb_canonical_tests.sh` | SpiceDB | 18 | Time limits, amount limits, delegation |
| `graphql_canonical_tests.sh` | GraphQL+PostgreSQL | 28 | Extended edge cases |
| `neo4j_canonical_tests.sh` | Neo4j | 18 | Identical to SpiceDB tests |

### Test Categories
1. **Positive Cases** - Valid permissions (✅)
2. **Negative Cases** - Denied access (❌)
3. **Time-Based Tests** - Expired/future access validation
4. **Amount-Based Tests** - Transaction limit enforcement
5. **Edge Cases** - Unknown users, invalid resources

---

## 🛠️ Development Workflow

### Initial Setup
```bash
# Start SpiceDB
cd spicedb-model/spicedb-config && docker-compose up -d
cd .. && ./setup.sh

# Start GraphQL
cd graphql-model && docker-compose up -d
npm install && npm run dev

# Start Neo4j
cd graphdb-model/neo4j-config && docker-compose up -d
```

### Start gRPC Service
```bash
cd go-entitlement-service
PORT=50052 go run cmd/server/main.go
```

### Run Tests
```bash
# From labs/ root
./spicedb_canonical_tests.sh
./graphql_canonical_tests.sh
./neo4j_canonical_tests.sh
```

---

## 📊 Functional Parity Matrix

| Feature | SpiceDB | GraphQL | Neo4j | Status |
|---------|---------|---------|-------|--------|
| Direct ownership | ✅ | ✅ | ✅ | **Parity** |
| Time-bound delegation | ✅ | ✅ | ✅ | **Parity** |
| Amount-based limits | ✅ | ✅ | ✅ | **Parity** |
| Multi-hop relationships | ✅ | ✅ | ✅ | **Parity** |
| Organizational roles | ✅ | ✅ | ✅ | **Parity** |
| Negative cases | ✅ | ✅ | ✅ | **Parity** |

**Result:** All three implementations achieve 100% functional parity for the defined test scenarios.

---

## 🔑 Key Design Decisions

### 1. **Why gRPC for the Service Layer?**
- High performance binary protocol
- Strong typing with Protocol Buffers
- Native support in Go
- Language-agnostic client generation

### 2. **Why Three Implementations?**
- **SpiceDB:** Industry-standard external authorization (Google Zanzibar pattern)
- **GraphQL:** Common application-layer pattern for comparison
- **Neo4j:** Graph-native alternative for relationship-heavy authorization

### 3. **Why Generic Resource Model?**
- Demonstrates authorization patterns applicable to any domain
- Avoids coupling to specific business logic
- Educational focus on architectural patterns

---

## 📚 Related Documentation

- **[README.md](README.md)** - Project overview and quick start
- **[COMPARISON.md](COMPARISON.md)** - Detailed architectural comparison
- **[EXAMPLES.md](EXAMPLES.md)** - Authorization use cases and test scenarios

---

**Last Updated:** October 2, 2025
