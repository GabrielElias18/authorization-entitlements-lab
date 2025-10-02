// Canonical data for entitlement model (aligned with SpiceDB and GraphQL)
// Clear existing data
MATCH (n) DETACH DELETE n;

// Create Organizations
CREATE (org_abc:Org {id: 'abc', name: 'abc'});

// Create Roles
CREATE (finance_ops:Role {id: 'finance_ops', name: 'finance_ops'});

// Create Users (including Bob and Charlie for negative testing)
CREATE (david:User {id: 'david', name: 'david', email: 'david@example.com'});
CREATE (emma:User {id: 'emma', name: 'emma', email: 'emma@example.com'});
CREATE (adi:User {id: 'adi', name: 'adi', email: 'adi@example.com'});
CREATE (alice:User {id: 'alice', name: 'alice', email: 'alice@example.com'});
CREATE (bob:User {id: 'bob', name: 'bob', email: 'bob@example.com'});
CREATE (charlie:User {id: 'charlie', name: 'charlie', email: 'charlie@example.com'});

// Create Accounts
CREATE (acc123:Account {id: 'acc123'});
CREATE (acc456:Account {id: 'acc456'});
CREATE (acc789:Account {id: 'acc789'});
CREATE (acc790:Account {id: 'acc790'});
CREATE (acc791:Account {id: 'acc791'});
CREATE (acc792:Account {id: 'acc792'});

// Role membership
MATCH (alice:User {id: 'alice'}), (finance_ops:Role {id: 'finance_ops'}) CREATE (alice)-[:HAS_ROLE]->(finance_ops);

// Role belongs to organization
MATCH (finance_ops:Role {id: 'finance_ops'}), (org_abc:Org {id: 'abc'}) CREATE (finance_ops)-[:MEMBER_OF]->(org_abc);

// Account ownership
MATCH (david:User {id: 'david'}), (acc123:Account {id: 'acc123'}) CREATE (david)-[:OWNS]->(acc123);
MATCH (emma:User {id: 'emma'}), (acc456:Account {id: 'acc456'}) CREATE (emma)-[:OWNS]->(acc456);

// Organizational account ownership
MATCH (org_abc:Org {id: 'abc'}), (acc789:Account {id: 'acc789'}) CREATE (org_abc)-[:OWNS]->(acc789);
MATCH (org_abc:Org {id: 'abc'}), (acc790:Account {id: 'acc790'}) CREATE (org_abc)-[:OWNS]->(acc790);
MATCH (org_abc:Org {id: 'abc'}), (acc791:Account {id: 'acc791'}) CREATE (org_abc)-[:OWNS]->(acc791);
MATCH (org_abc:Org {id: 'abc'}), (acc792:Account {id: 'acc792'}) CREATE (org_abc)-[:OWNS]->(acc792);

// POA relationships (direct relationships as expected by Go service)
// Emma has POA on acc123 (with payment limits and time restrictions) - updated for current testing
MATCH (emma:User {id: 'emma'}), (acc123:Account {id: 'acc123'}) CREATE (emma)-[:HAS_POA {payment_limit: 5000, starts_at: '2025-01-01', expires_at: '2025-12-31'}]->(acc123);

// Adi has accountant access to acc456 (time-bound, same as POA but different relationship type) - updated for current testing
MATCH (adi:User {id: 'adi'}), (acc456:Account {id: 'acc456'}) CREATE (adi)-[:HAS_ACCOUNTANT_ACCESS {starts_at: '2025-01-01', expires_at: '2025-12-31'}]->(acc456);

// Note: Bob and Charlie have no relationships - used for negative testing 