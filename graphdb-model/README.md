# GraphDB Model: Resource Authorization

This module implements a real-world authorization
system using Neo4j, demonstrating generic
resource access control patterns.

This directory contains the schema, data, and automation for loading and testing entitlement/authorization models in Neo4j.

## Use Cases Covered

- **Transaction Initiation** (with amount limits and time validation)
- **Document Access** (with time validation)
- **Record Viewing** (with time validation)
- **Delegated Authority** (time-bound with transaction limits)
- **Auditor Access** (time-bound, documents only)
- **Role-based Access** (organizational permissions)

## Test Coverage (Aligned with SpiceDB)

- **18 comprehensive test cases** matching SpiceDB exactly
- **Identical test descriptions** for direct comparison
- **Same output format**: [PASS]/[FAIL] with pass/fail counts
- **Transaction limit testing**: Under limit (3000), at limit (5000), above limit (5001)
- **Time boundary testing**: Before start, within range, after expiry
- **Delegation scenarios**: Emma's access to David's resource via delegation
- **Auditor restrictions**: Adi can view/download but not initiate transactions
- **Role-based access**: Alice's organizational permissions
- **Negative testing**: Bob with no access relationships

## Contents
- `schema.cypher` — Neo4j schema (constraints and indexes)
- `sample-data.cypher` — Canonical data with relationships and properties
- `check-permission.ts` — Smart permission tests using data store logic
- `permission-queries.cypher` — Reusable smart permission queries
- `neo4j-config/` — Configuration for running Neo4j locally (e.g., Docker Compose)

## Workflow

### 1. Start Neo4j
```sh
cd neo4j-config
docker-compose up -d
```

### 2. Load Schema and Data
```sh
# Load schema (constraints and indexes)
cypher-shell -u neo4j -p password < schema.cypher

# Load canonical data
cypher-shell -u neo4j -p password < sample-data.cypher
```

### 3. Run Permission Tests
```sh
npm install
npx ts-node check-permission.ts
```
This will check all key permissions and scenarios, printing pass/fail for each test.

### 4. Access Neo4j Browser
Open [http://localhost:7474](http://localhost:7474) in your browser.

## Sample Data

| User | ID | Role | Access |
|------|----|----|--------|
| **David** | `david` | Owner | Full access to res123 |
| **Emma** | `emma` | Owner + Delegation | Owns res456 + delegation on res123 (limit: 5000, expires: 2025-06-30) |
| **Adi** | `adi` | Auditor | Document access to res456 (expires: 2025-03-31) |
| **Alice** | `alice` | operations | Organizational access to org-abc resources |
| **Bob** | `bob` | None | No access (negative testing) |
| **Charlie** | `charlie` | None | No access (negative testing) |

### Resource Structure
- **res123**: Owned by David
- **res456**: Owned by Emma
- **res789-792**: Owned by org-abc (organizational resources)

## Smart Permission Model

### Business Logic in Data Store
The permission logic is encapsulated in smart Cypher queries that handle:

- **Time validation**: All relationships respect `expires_at` properties
- **Transaction limits**: Delegation relationships enforce `transaction_limit` constraints
- **Role hierarchies**: Organizational access through role membership
- **Access restrictions**: Auditors can only view documents, not initiate transactions

### Permission Types

#### Transaction Initiation
```cypher
// Owners, delegated authority (within limits), and role-based access can initiate transactions
// Auditors CANNOT initiate transactions
```

#### Document Access & Record Viewing
```cypher
// Owners, delegated authority, role-based access, and auditors can view/download
// All subject to time validation
```

### Graph Schema
```cypher
// Core nodes
(:User {id: string, name: string, email: string})
(:Resource {id: string})
(:Org {id: string, name: string})
(:Role {id: string, name: string})

// Key relationships with properties
(:User)-[:OWNS]->(:Resource)
(:User)-[:HAS_DELEGATION {transaction_limit: number, expires_at: string}]->(:Resource)
(:User)-[:HAS_AUDITOR_ACCESS {expires_at: string}]->(:Resource)
(:User)-[:HAS_ROLE]->(:Role)
(:Role)-[:MEMBER_OF]->(:Org)
(:Org)-[:OWNS]->(:Resource)
```

## Key Features

### Smart Data Store Benefits
- **Business logic encapsulated** in Cypher queries
- **Fast permission checks** with proper indexing
- **Consistent rules** applied across all applications
- **Time validation** handled automatically
- **Transaction limits** enforced at data store level

### Permission Hierarchy
1. **Direct Ownership**: Full access to owned resources
2. **Delegated Authority**: Full access with time + transaction limits
3. **Role-based Access**: Organizational resource access
4. **Auditor Access**: Limited access (documents only) with time bounds

### Time-Bound Access
- **Emma's Delegation**: Expires 2025-06-30 (aligned with SpiceDB for testing)
- **Adi's Auditor Access**: Expires 2025-03-31 (aligned with SpiceDB for testing)

## Extending
- Add new relationships to `sample-data.cypher`
- Add new test cases to `check-permission.ts`
- Update smart queries in `permission-queries.cypher`
- Update schema constraints in `schema.cypher`

## Requirements
- [Neo4j](https://neo4j.com/docs/) running locally (see `neo4j-config/`)
- [Node.js](https://nodejs.org/) (v18+) for test scripts
- [TypeScript](https://www.typescriptlang.org/) for test compilation

## Example: Adding a New Permission Type
Add to `check-permission.ts`:
```typescript
async function checkNewPermission(userId: string, accountId: string): Promise<boolean> {
  const query = `
    // Smart query with business logic
    MATCH (user:User {id: $userId})
    MATCH (account:Account {id: $accountId})
    // ... permission logic ...
    RETURN canAccess as result
  `;
  // Implementation
}
```

---
For more details, see the main project README or the [Neo4j documentation](https://neo4j.com/docs/). 