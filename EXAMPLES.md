# Authorization Architecture Use Cases: SpiceDB vs Neo4j

This document presents comprehensive resource authorization use cases implemented and validated across both SpiceDB (external authorization engine) and Neo4j (database-embedded authorization) with **100% functional parity** achieved.

## Authorization Patterns Implemented

### **Core Functions**
- **Document Access**: Document retrieval with time validation
- **Transaction Initiation**: Transaction creation with amount limits
- **Record Viewing**: Read access with delegation support

### **Implementation Approaches**
- **SpiceDB**: External authorization engine with caveat-based rules
- **Neo4j**: Database-embedded authorization with Cypher query logic

### **Validation Results**
- **18 comprehensive test scenarios** validated across both systems
- **100% functional parity** - identical authorization decisions
- **Complex business rules** successfully implemented in both patterns

## Valid Access Paths

### Direct Ownership
A person who directly owns a resource can:
- Access documents
- View records
- Initiate transactions

### Delegation by Owner
An owner may grant access to another person for a specific resource:
- Via delegated authority for a defined time period
- The delegate can perform functions such as:
  - Access documents
  - View records
  - Initiate transactions (if permissioned and within limits)

### Time-Bound Delegation to Auditors
Resource owners may delegate temporary access to auditors:
- Auditor access is typically scoped to document access
- Valid for a specific time window (e.g., 2 months)

### Organizational Delegation via Role
Operations staff working for an organization that manages multiple resources:
- May act on behalf of the organization
- Can perform functions including:
  - Accessing documents
  - Viewing records
  - Initiating transactions (if permitted by their role)

## Users and Context

### **Test Users**
- **David**: Direct owner of resource res123
- **Emma**:
  - Delegated authority for res123 (transaction limit: $5,000, expires: 2025-06-30)
  - Direct owner of res456
- **Adi**: Temporary auditor access to res456 (2025-01-01 to 2025-03-31)
- **Alice**: Operations staff at org ABC (manages res789–res792)
- **Bob**: No access relationships (negative testing)

### **Time-Bound Access Details**
- **Emma's Delegation**: Valid from 2025-01-01 to 2025-06-30 with $5,000 transaction limit
- **Adi's Auditor Access**: Valid from 2025-01-01 to 2025-03-31 (view/access only)

## Comprehensive Test Scenarios (18 Cases)

### **Emma's Delegation (Time + Limit) Tests**

| # | User | Resource | Function | Context | SpiceDB | Neo4j | Expected |
|---|------|----------|----------|---------|---------|-------|----------|
| 1 | Emma | res123 | can_initiate_transaction | amount=3000, date=2025-03-15 | ✅ | ✅ | **Allowed** |
| 2 | Emma | res123 | can_initiate_transaction | amount=5000, date=2025-06-15 | ✅ | ✅ | **Allowed** |
| 3 | Emma | res123 | can_initiate_transaction | amount=5001, date=2025-03-15 | ❌ | ❌ | **Denied** |
| 4 | Emma | res123 | can_initiate_transaction | amount=3000, date=2024-12-31 | ❌ | ❌ | **Denied** |
| 5 | Emma | res123 | can_initiate_transaction | amount=3000, date=2025-07-01 | ❌ | ❌ | **Denied** |

### **Adi's Auditor Access Tests**

| # | User | Resource | Function | Context | SpiceDB | Neo4j | Expected |
|---|------|----------|----------|---------|---------|-------|----------|
| 6 | Adi | res456 | can_view_records | date=2025-02-15 | ✅ | ✅ | **Allowed** |
| 7 | Adi | res456 | can_access_document | date=2025-03-30 | ✅ | ✅ | **Allowed** |
| 8 | Adi | res456 | can_view_records | date=2024-12-31 | ❌ | ❌ | **Denied** |
| 9 | Adi | res456 | can_view_records | date=2025-04-01 | ❌ | ❌ | **Denied** |
| 10 | Adi | res456 | can_access_document | date=2025-04-01 | ❌ | ❌ | **Denied** |
| 11 | Adi | res456 | can_initiate_transaction | date=2025-02-15 | ❌ | ❌ | **Denied** |

### **Delegated Access Tests**

| # | User | Resource | Function | Context | SpiceDB | Neo4j | Expected |
|---|------|----------|----------|---------|---------|-------|----------|
| 12 | Emma | res123 | can_view_records | via delegation, date=2025-03-15 | ✅ | ✅ | **Allowed** |
| 13 | Emma | res123 | can_access_document | via delegation, date=2025-03-15 | ✅ | ✅ | **Allowed** |
| 14 | Emma | res123 | can_view_records | via delegation, date=2025-07-01 | ❌ | ❌ | **Denied** |

### **Direct and Role-based Tests**

| # | User | Resource | Function | Context | SpiceDB | Neo4j | Expected |
|---|------|----------|----------|---------|---------|-------|----------|
| 15 | David | res123 | can_access_document | Owner | ✅ | ✅ | **Allowed** |
| 16 | Emma | res456 | can_access_document | Owner | ✅ | ✅ | **Allowed** |
| 17 | Alice | org:abc | can_access | Role membership | ✅ | ✅ | **Allowed** |
| 18 | Bob | res123 | can_access_document | No relationship | ❌ | ❌ | **Denied** |

### **Test Results Summary**
- **Total Scenarios**: 18 comprehensive test cases
- **SpiceDB Results**: 18/18 passed (100%)  
- **Neo4j Results**: 18/18 passed (100%)
- **Functional Parity**: ✅ **VALIDATED** - Identical authorization decisions

## Implementation Examples

### **SpiceDB Approach (External Authorization Engine)**

```yaml
# Schema definition with caveats
definition resource {
  relation owner: user
  relation delegated_access: delegation
  relation auditor_access: user with within_active_range

  permission can_initiate_transaction = owner + delegated_access->delegate_with_time_and_limit
}

caveat within_time_and_limit(amount int, max_amount int, start timestamp, end timestamp, now timestamp) {
  amount <= max_amount && now >= start && now <= end
}
```

**Usage:**
```bash
# Test Case 1: Emma transaction under limit
zed permission check delegation:del1 can_initiate_transaction user:emma \
  --caveat-context '{"amount":3000,"now":"2025-03-15T00:00:00Z"}'
# Result: true
```

### **Neo4j Approach (Database-Embedded Authorization)**

```cypher
// Permission check query with business logic
MATCH (user:User {id: $userId})-[del:HAS_DELEGATION]->(resource:Resource {id: $resourceId})
WHERE datetime($testDate) >= datetime(del.starts_at)
  AND datetime($testDate) <= datetime(del.expires_at)
  AND ($amount <= del.transaction_limit OR del.transaction_limit IS NULL)
RETURN count(del) > 0 as canInitiateTransaction
```

**Usage:**
```typescript
// Test Case 1: Emma transaction under limit
const result = await checkTransactionPermission('emma', 'res123', 3000, '2025-03-15T00:00:00Z');
// Result: true
```

## Architectural Comparison

| **Aspect** | **SpiceDB** | **Neo4j** |
|------------|-------------|-----------|
| **Rule Definition** | Declarative schema + caveats | Cypher query logic |
| **Business Logic** | External authorization engine | Database-embedded queries |
| **Time Validation** | Dynamic caveat evaluation | DateTime comparisons |
| **Complex Rules** | Caveat expressions | Query conditions |
| **Development Model** | Schema-first approach | Query-first approach |

## Production Considerations

### **Choose SpiceDB When:**
- Building dedicated authorization infrastructure
- Need separation between auth and business logic
- Want standardized Zanzibar patterns
- Scaling authorization across multiple services

### **Choose Neo4j When:**
- Prefer database-driven authorization
- Want visual representation of access patterns
- Need rapid prototyping of permission models
- Comfortable with graph query languages

## Conclusion

Both SpiceDB and Neo4j successfully implement these complex resource authorization use cases with **identical functional outcomes**. The validation of 18 comprehensive test scenarios demonstrates that both architectural patterns can handle:

- **Complex delegation patterns** (delegated authority)
- **Time-bound access controls** (temporary access windows)
- **Amount-based restrictions** (transaction limits)
- **Multi-hop relationships** (organizational roles)
- **Negative access scenarios** (boundary testing)

The choice between approaches depends on architectural preferences, team expertise, and infrastructure requirements rather than functional capabilities.

