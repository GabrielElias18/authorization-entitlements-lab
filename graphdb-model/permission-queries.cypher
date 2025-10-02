// Smart Permission Queries - Business Logic in Data Store
// These queries encapsulate all permission logic and can be reused across applications

// Payment Initiation Permission Query
// Returns true if user can initiate payment on account with specified amount
// Business Rules:
// - Owners can always pay
// - POA can pay within limits and time bounds
// - Role-based access can pay
// - Accountants CANNOT pay (only view statements)
MATCH (user:User {id: $userId})
MATCH (account:Account {id: $accountId})
WITH user, account
OPTIONAL MATCH (user)-[owns:OWNS]->(account)
WITH user, account, owns IS NOT NULL as direct_owner
OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
WHERE datetime() <= datetime(poa.expires_at) AND ($amount <= poa.payment_limit OR poa.payment_limit IS NULL)
WITH user, account, direct_owner, count(poa) as valid_poa
OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
WITH user, account, direct_owner, valid_poa, count(role) as role_access
RETURN (direct_owner OR valid_poa > 0 OR role_access > 0) as canInitiatePayment;

// Statement Download Permission Query
// Returns true if user can download statements for account
// Business Rules:
// - Owners can always download
// - POA can download within time bounds
// - Role-based access can download
// - Accountants can download within time bounds
MATCH (user:User {id: $userId})
MATCH (account:Account {id: $accountId})
WITH user, account
OPTIONAL MATCH (user)-[owns:OWNS]->(account)
WITH user, account, owns IS NOT NULL as direct_owner
OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
WHERE datetime() <= datetime(poa.expires_at)
WITH user, account, direct_owner, count(poa) as valid_poa
OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
WITH user, account, direct_owner, valid_poa, count(role) as role_access
OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
WHERE datetime() <= datetime(acc.expires_at)
WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as canDownloadStatement;

// Transaction Viewing Permission Query
// Returns true if user can view transactions for account
// Business Rules: Same as statement download
// - Owners can always view
// - POA can view within time bounds
// - Role-based access can view
// - Accountants can view within time bounds
MATCH (user:User {id: $userId})
MATCH (account:Account {id: $accountId})
WITH user, account
OPTIONAL MATCH (user)-[owns:OWNS]->(account)
WITH user, account, owns IS NOT NULL as direct_owner
OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
WHERE datetime() <= datetime(poa.expires_at)
WITH user, account, direct_owner, count(poa) as valid_poa
OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
WITH user, account, direct_owner, valid_poa, count(role) as role_access
OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
WHERE datetime() <= datetime(acc.expires_at)
WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as canViewTransactions; 