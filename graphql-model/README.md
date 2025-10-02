# GraphQL Model: Resource Authorization (Reference Implementation)

## ⚠️ Architectural Note

This GraphQL implementation serves as a **reference implementation** demonstrating application-layer authorization patterns. After comprehensive evaluation against SpiceDB and Neo4j, this approach is **not recommended** for production authorization systems due to architectural anti-patterns.

**Use this implementation for:**
- Understanding application-layer authorization patterns
- Educational comparison with specialized approaches  
- API gateway patterns (over dedicated authorization engines)
- Rapid prototyping and demos

**For production authorization, see:**
- [SpiceDB Model](../spicedb-model/) - External authorization engine
- [Neo4j Model](../graphdb-model/) - Database-embedded authorization

This directory contains the schema, data, and automation for testing application-layer entitlement patterns using GraphQL, PostgreSQL, and Prisma.

## Use Cases Covered

- View Records
- Initiate transactions
- Access documents
- Transaction Limits
- Delegated access
- Time Bound Access
- Auditor access (distinct from delegation)
- Access and delegation scoped by organization membership

## 🏗 Authorization Architecture Pattern

### Application-Layer Authorization
This implementation follows the **application-layer authorization** pattern where:
- Authorization logic resides in application code (GraphQL resolvers)
- Business rules are implemented as TypeScript functions  
- Permission checks happen at the API layer
- Data is stored in traditional relational format (PostgreSQL)

### Why This Pattern Has Limitations
- **Coupling**: Authorization tightly coupled to application logic
- **Scalability**: All auth decisions bottleneck through app layer
- **Maintenance**: Permission changes require code deployments  
- **Consistency**: No centralized authorization model
- **Specialization**: General-purpose API vs authorization-focused systems

## 🚀 Quick Start (Reference Implementation)

> **Note**: This is for educational/reference purposes. For production authorization, consider [SpiceDB](../spicedb-model/) or [Neo4j](../graphdb-model/) approaches.

### Educational Use
```bash
# 1. Setup for learning/comparison
bash setup.sh

# 2. Explore application-layer patterns
npm run dev

# 3. Compare with other approaches  
bash run_permission_checks.sh

# 4. Reset data (if needed)
bash reset.sh
```

## Contents
- `schema.graphql` — GraphQL schema definition
- `prisma/schema.prisma` — Database schema (Prisma ORM)
- `init-scripts/01-seed-data.sql` — Canonical data loading (aligned with SpiceDB/GraphDB)
- `setup.sh` — One-command setup: starts PostgreSQL, runs migrations, loads data
- `reset.sh` — Clear and reload data (keeps database running)
- `run_permission_checks.sh` — Comprehensive test script for permissions
- `docker-compose.yml` — Configuration for running PostgreSQL locally
- `src/` — GraphQL server and resolvers

## Workflow

### 1. Initial Setup
```bash
bash setup.sh
```
This single command will:
- Start PostgreSQL container
- Install Node.js dependencies
- Run Prisma migrations
- Load canonical data from `init-scripts/01-seed-data.sql`

### 2. Start GraphQL Server
```bash
npm run dev
```
This starts the GraphQL server on http://localhost:4000/

### 3. Run Tests
```bash
bash run_permission_checks.sh
```
This will check all key permissions and scenarios, printing pass/fail for each test.

### 4. Access GraphQL Playground
Open [http://localhost:4000/](http://localhost:4000/) in your browser.

### 5. Reset Data (Optional)
```bash
bash reset.sh
```
Clear all data and reload from files (useful after schema changes).

## Authentication

The system uses **HTTP header-based authentication** for demonstration:

```json
{
  "x-user-id": "david"
}
```

## 📊 Evaluation Results vs SpiceDB/Neo4j

### What We Learned
After implementing identical authorization scenarios across all three approaches:

| **Aspect** | **GraphQL** | **SpiceDB** | **Neo4j** | **Winner** |
|------------|-------------|-------------|-----------|------------|
| **Authorization Specialization** | General-purpose API | Purpose-built engine | Query-based rules | **SpiceDB** |
| **Rule Complexity** | Application code | Caveat expressions | Cypher logic | **SpiceDB** |
| **Development Speed** | Familiar patterns | Learning curve | Query-first | **GraphQL** |
| **Production Scaling** | App bottlenecks | Distributed auth | DB clustering | **SpiceDB** |
| **Maintenance** | Code deployments | Schema migrations | Query updates | **Neo4j** |

### Functional Parity Achieved
- ✅ **18 test scenarios** implemented and passing
- ✅ **Identical authorization decisions** to SpiceDB/Neo4j
- ✅ **Complex rules supported** (delegation, time bounds, transaction limits)
- ✅ **Educational value** for understanding application-layer patterns

## Canonical Data (Aligned with SpiceDB/GraphDB)

| User | ID | Role | Access |
|------|----|----|--------|
| **David** | `david` | Owner | Full access to res123 |
| **Emma** | `emma` | Delegation | Delegate access to res123 (limit: 5000, expires: 2025-06-30) |
| **Adi** | `adi` | Auditor | Time-bound access to res456 (2025-01-01 to 2025-03-31) |
| **Alice** | `alice` | operations | Organizational access to org-abc resources |
| **Bob** | `bob` | None | No access (negative testing) |
| **Charlie** | `charlie` | None | No access (negative testing) |

### Resource Structure
- **res123**: Owned by David (david)
- **res456**: Owned by Emma (emma)
- **res789-792**: Owned by org-abc (organizational resources)

## Key Features

### Fine-grained Permissions
- **Direct ownership**: Resource owners have full access
- **Delegated authority**: Time-bound and amount-limited delegation
- **Auditor access**: Time-bound auditor access (distinct from delegation)
- **Organizational roles**: Role-based access within organizations
- **Document access**: Resource-level document access permissions

### GraphQL Schema
```graphql
type Query {
  canInitiateTransaction(resourceId: ID!, amount: Float!): Boolean!
  canAccessDocument(resourceId: ID!): Boolean!
  me: User
  resource(id: ID!): Resource
}
```

### Data Model
- **Delegation**: Grants a user (delegate) permission to initiate transactions on a resource, with optional time and amount limits.
- **AuditorAccess**: Grants a user time-bound auditor access to a resource (distinct from delegation), e.g. for Adi on res456.

## 💡 When to Use GraphQL in Authorization Systems

### ✅ Recommended: API Gateway Pattern
```
┌─────────────────────────────────────────────────────────────┐
│                    GraphQL API Gateway                     │
│              (Data Aggregation & Client Interface)         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   SpiceDB   │  │ Business    │  │   Neo4j     │          │
│  │(Authorization)│  │   Data      │  │(Analytics)  │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

**GraphQL works well for:**
- Aggregating authorization decisions with business data
- Providing flexible client interfaces
- Rapid prototyping and demos
- Data fetching layer over specialized systems

### ❌ Not Recommended: Primary Authorization Engine
- Authorization logic scattered across resolvers
- No specialized optimization for permission checks  
- Maintenance burden for complex rules
- Scalability limitations for high-frequency auth decisions

### Migration to Production Patterns
If you started with this approach and want to migrate:
- **To SpiceDB**: Extract resolver logic into Zanzibar schema + caveats
- **To Neo4j**: Convert SQL queries to Cypher with embedded logic
- **Keep GraphQL**: As API gateway over specialized authorization system

## Requirements
- [Docker](https://docs.docker.com/get-docker/) for PostgreSQL
- [Node.js](https://nodejs.org/) (v18+) for GraphQL server
- [Prisma CLI](https://www.prisma.io/docs/concepts/components/prisma-cli) (included in dependencies)

## Extending
- Add new permissions to `schema.graphql`
- Update resolver logic in `src/entitlement-resolvers.ts`
- Add new test cases to `run_permission_checks.sh`
- Update database schema in `prisma/schema.prisma` as needed

## Example: Adding a New Permission
Add to `schema.graphql`:
```graphql
type Query {
  canViewRecords(resourceId: ID!): Boolean!
}
```

Then add resolver logic in `src/entitlement-resolvers.ts`:
```typescript
canViewRecords: async (_, { resourceId }, { user }) => {
  // Permission logic here
}
```

## Example: Adding a New Test
Add to `run_permission_checks.sh`:
```bash
check_permission "$DAVID_ID" "query { canViewRecords(resourceId: \"res123\") }" "true" "David can view records on his resource"
```

## Conclusion

This GraphQL implementation successfully demonstrates application-layer authorization patterns and achieves **100% functional parity** with SpiceDB and Neo4j approaches. However, it serves best as:

- **Educational reference** for understanding authorization anti-patterns
- **Rapid prototyping** tool for exploring authorization concepts  
- **API gateway** candidate over specialized authorization systems
- **Migration starting point** toward production-ready patterns

**For production authorization systems**, choose specialized approaches:
- **[SpiceDB Model](../spicedb-model/)** for external authorization engines
- **[Neo4j Model](../graphdb-model/)** for database-embedded authorization

---
For more details, see the main project README, [SpiceDB documentation](https://authzed.com/docs/spicedb/), or [GraphQL documentation](https://graphql.org/). 