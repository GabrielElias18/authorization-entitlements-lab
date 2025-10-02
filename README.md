# Fine Grain Entitlements Lab: Authorization Architecture Comparison

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🎯 Project Focus: Authorization System Patterns

This project implements and compares different authorization architectures using practical side-by-side comparison of typical access control use cases, focusing on:

### 🏆 **Authorization Approaches**
- **SpiceDB (Zanzibar-style)**: External authorization engine with relationship modeling
- **Neo4j (GraphDB)**: Database-embedded authorization with graph queries
- **GraphQL + PostgreSQL**: Application-layer authorization with relational data

⚠️ Disclaimer  
This project is personal research and for educational purposes only.  
All examples and financial sector use cases are **generic and hypothetical**. 

## 🚀 Why These Models?

### **SpiceDB (Zanzibar) - External Authorization Engine**
- Purpose-built for fine-grained access control
- Declarative schema with relationship modeling
- Caveat expressions for complex business rules
- Distributed architecture with consistency guarantees
- Separation of authorization logic from application code

### **Neo4j (GraphDB) - Database-Embedded Authorization**
- Native graph relationship modeling
- Cypher queries for complex permission logic
- Visual representation of access patterns
- Efficient traversal of multi-hop relationships
- Single system for data and authorization logic

### **GraphQL + PostgreSQL - Application-Layer Authorization**
- Familiar SQL patterns for business logic
- ACID transactions for data integrity
- Rich querying with GraphQL interface
- Direct integration with application logic
- Mature ecosystem with established tooling
- Reference implementation for comparison

## 🏗️ Architecture Patterns

```
┌─────────────────────────────────────────────────────────────┐
│                    Go gRPC Service                          │
│              (Centralized Entitlement API)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   SpiceDB   │  │   GraphQL   │  │    Neo4j    │          │
│  │ (External)  │  │(Application)│  │ (Embedded)  │          │
│  │             │  │             │  │             │          │
│  │ • Schema    │  │ • ACID      │  │ • Graph     │          │
│  │ • Caveats   │  │ • SQL       │  │ • Cypher    │          │
│  │ • Tuples    │  │ • ORM       │  │ • Visual    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘

External Engine     │ Application Layer  │ Database Embedded
• Dedicated system  │ • Integrated logic │ • Query-based rules
• Relationship model│ • Business context │ • Relationship graph
• Distributed scale │ • Transactional    │ • Single system
```

## 📊 Implementation Characteristics

| Aspect | SpiceDB | GraphQL + PostgreSQL | Neo4j GraphDB |
|--------|---------|----------------------|---------------|
| **Authorization Model** | External engine | Application-integrated | Database-embedded |
| **Query Approach** | Relationship + caveats | SQL + business logic | Cypher graph queries |
| **Schema Management** | Declarative model | Database migrations | Graph schema |
| **Complex Rules** | Caveat expressions | Application code | Query logic |
| **Scalability Pattern** | Distributed authorization | Application scaling | Database clustering |
| **Development Model** | Schema-first | Code-first | Query-first |

### 🔍 Architectural Trade-offs
- **SpiceDB**: Specialized for authorization, external dependency, caveat flexibility
- **Neo4j**: Visual modeling, single system, query complexity
- **GraphQL**: Familiar patterns, application coupling, transactional consistency. Anti pattern with authroisation logic in application.


## Hypothetical Use Cases for Fine-Grained Access Control

### **Core Access Patterns**
1. **Resource Ownership**: Direct user-account relationships.
2. **Delegated Access**: Time-bound delegation with limits
3. **Organizational Roles**: Role-based access across accounts
4. **Delegated Access**: professional access with limited privileges
5. **Document Downloads**: Controlled document access

### **Complex Scenarios**
- **Amount-based limits** for delegated access
- **Time-window validation** for temporary access
- **Multi-hop relationships** through organizations
- **Context-aware permissions** with caveats

## 🚀 Quick Start

### **1. Start the Go Service**
```bash
cd go-entitlement-service
PORT=50052 go run cmd/server/main.go &
```

### **2. Run Canonical Tests**
```bash
# Test all implementations
./spicedb_canonical_tests.sh    # SpiceDB (18 tests)
./graphql_canonical_tests.sh    # GraphQL (28 tests)  
./neo4j_canonical_tests.sh      # Neo4j (18 tests)
```

### **3. Individual Backend Setup**

#### **SpiceDB (Zanzibar)**
```bash
cd spicedb-model/spicedb-config
docker-compose up -d
cd .. && ./setup.sh
```

#### **GraphQL + PostgreSQL**
```bash
cd graphql-model
docker-compose up -d
npm install
npm run dev  # Starts GraphQL server on :4000
```

#### **Neo4j (Experimental)**
```bash
cd graphdb-model/neo4j-config
docker-compose up -d
cd .. && npm install
```

## 🧪 Testing Strategy

### **Canonical Test Suite**
- **64 total tests** across all implementations
- **Identical scenarios** for fair comparison (SpiceDB vs Neo4j: 18 each)
- **Edge case coverage** (unknown users, accounts)
- **Functional verification** for each backend

### **Test Categories**
- ✅ **Positive Cases**: Valid permissions
- ❌ **Negative Cases**: Denied access
- 🔄 **Edge Cases**: Unknown entities, empty inputs
- ⏰ **Time-based**: Expired access validation
- 💰 **Amount-based**: Transaction limits

## 📁 Project Structure

```
labs/
├── go-entitlement-service/     # 🚀 Central gRPC service
├── spicedb-model/              # 🔐 SpiceDB (Zanzibar)
├── graphql-model/              # 🗄️ GraphQL + PostgreSQL  
├── graphdb-model/              # 🕸️ Neo4j (Experimental)
├── spicedb_canonical_tests.sh  # ✅ SpiceDB tests
├── graphql_canonical_tests.sh  # ✅ GraphQL tests
├── neo4j_canonical_tests.sh    # ✅ Neo4j tests
├── README.md                   # 📖 This file
├── PROJECT_STRUCTURE.md        # 🏗️ Architecture docs
├── COMPARISON.md              # 📊 Implementation comparison
└── USECASES.md                # 💼 Use case documentation
```

## 🎯 Architectural Recommendations

### **Choose SpiceDB When:**
- Building dedicated authorization services
- Need complex caveat-based rules
- Scaling authorization across multiple applications
- Want separation of concerns between auth and business logic

### **Choose GraphQL + PostgreSQL When:**
- Familiar with relational database patterns
- Need tight integration with business logic
- Want ACID transaction guarantees
- Working with existing PostgreSQL infrastructure

### **Choose Neo4j When:**
- Modeling complex relationship hierarchies
- Need visual representation of access patterns
- Want to embed authorization logic in database queries
- Exploring graph-based authorization patterns

## 🏗 Architectural Comparison: SpiceDB vs GraphDB

After implementing identical authorization scenarios in both systems, here's the architectural analysis:

### **Core Architecture Differences**

| **Aspect** | **SpiceDB** | **GraphDB (Neo4j)** | **Developer Impact** |
|------------|-------------|---------------------|---------------------|
| **Architecture** | External authorization engine | Business logic in data store | GraphDB: Single system to manage |
| **Query Language** | Caveat expressions + Zanzibar | Cypher graph queries | GraphDB: More familiar SQL-like syntax |
| **Time Handling** | Dynamic caveat evaluation | DateTime comparisons in queries | SpiceDB: More flexible, GraphDB: More transparent |
| **Performance Profile** | Network + computation overhead | Direct database query execution | GraphDB: Fewer network hops |
| **Scalability** | Distributed authorization | Neo4j cluster scalability | SpiceDB: Purpose-built scaling |
| **Development Model** | Schema + relationship tuples | Graph schema + Cypher logic | GraphDB: More intuitive for DB developers |

### **Developer Experience: Adding New Permissions**

#### **Simple Permission: "can_transfer_funds"**

**SpiceDB Approach:**
```yaml
# Update model.zaml schema
definition account {
  relation delegate: user with transfer_limit
  permission can_transfer_funds = owner + delegate
}

caveat transfer_limit(amount int, max_amount int) {
  amount <= max_amount
}
```

**GraphDB Approach:**
```cypher
// Add relationship and query function
MATCH (user:User)-[rel:CAN_TRANSFER]->(account:Account)
WHERE $amount <= rel.max_amount
RETURN count(rel) > 0 as canTransfer
```

#### **Complex Logic: "Approval Workflows with User Tiers"**

**SpiceDB:** Declarative caveat expressions handle complex rules elegantly.

**GraphDB:** Business logic embedded directly in Cypher queries

### **Scaling Considerations**

| **Factor** | **SpiceDB** | **GraphDB** | **Winner** |
|------------|-------------|-------------|-----------|
| **Learning Curve** | Zanzibar concepts + caveats | Cypher + graph thinking | **GraphDB** |
| **Adding Simple Permissions** | Schema + tuple load | Single Cypher query | **GraphDB** |
| **Complex Business Logic** | Caveat expressions | Embedded query logic | **SpiceDB** |
| **Performance Optimization** | Relationship modeling | Index tuning + query optimization | **SpiceDB** |
| **Schema Evolution** | Versioned migrations | Database migrations | **Tie** |
| **Error Handling** | Authorization-specific errors | Database + application errors | **SpiceDB** |

### **Recommendation**

- **Choose SpiceDB** for: Complex authorization rules, distributed systems, large-scale requirements
- **Choose GraphDB** for: Rapid prototyping, familiar database patterns, rich relationship modeling

Both approaches achieved **100% functional parity** with identical test results across 18 comprehensive scenarios.

## 🔧 Technology Stack

### **Core Service**
- **Go 1.21+**: High-performance gRPC service
- **Protocol Buffers**: Efficient binary serialization
- **gRPC**: Low-latency RPC framework

### **Backend Databases**
- **SpiceDB**: Zanzibar-style permission system
- **PostgreSQL**: ACID-compliant relational database
- **Neo4j**: Graph database for complex relationships

### **Development Tools**
- **Docker Compose**: Local development environment
- **Prisma ORM**: Type-safe database access
- **GraphQL**: Flexible query interface

## 📚 Documentation

- [**PROJECT_STRUCTURE.md**](PROJECT_STRUCTURE.md) - Detailed architecture
- [**COMPARISON.md**](COMPARISON.md) - Implementation analysis
- [**USECASES.md**](USECASES.md) - Complete use case documentation

---

**Last Updated:** September 28, 2025  
**Status:** ✅ All implementations passing canonical tests  
**Coverage:** 100% functional parity between SpiceDB and Neo4j models


