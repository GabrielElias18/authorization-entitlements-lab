import neo4j from 'neo4j-driver';

console.log("=== Neo4j Smart Permission Checks ===");

let pass_count = 0;
let fail_count = 0;

const driver = neo4j.driver(
  'neo4j://localhost:7687',
  neo4j.auth.basic('neo4j', 'password')
);

async function checkPaymentPermission(userId: string, accountId: string, amount: number, testDate?: string): Promise<boolean> {
  const session = driver.session();
  try {
    // Smart query with business logic in data store
    const query = `
      MATCH (user:User {id: $userId})
      MATCH (account:Account {id: $accountId})
      WITH user, account
      OPTIONAL MATCH (user)-[owns:OWNS]->(account)
      WITH user, account, owns IS NOT NULL as direct_owner
      OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
      WHERE ${testDate ? `datetime($testDate) >= datetime(poa.starts_at) AND datetime($testDate) <= datetime(poa.expires_at)` : `datetime() >= datetime(poa.starts_at) AND datetime() <= datetime(poa.expires_at)`} AND ($amount <= poa.payment_limit OR poa.payment_limit IS NULL)
      WITH user, account, direct_owner, count(poa) as valid_poa
      OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
      WITH user, account, direct_owner, valid_poa, count(role) as role_access
      RETURN (direct_owner OR valid_poa > 0 OR role_access > 0) as canPay
    `;
    const result = await session.run(query, { userId, accountId, amount, testDate });
    return result.records[0].get('canPay');
  } finally {
    await session.close();
  }
}

async function checkStatementPermission(userId: string, accountId: string, testDate?: string): Promise<boolean> {
  const session = driver.session();
  try {
    // Smart query with business logic in data store
    const query = `
      MATCH (user:User {id: $userId})
      MATCH (account:Account {id: $accountId})
      WITH user, account
      OPTIONAL MATCH (user)-[owns:OWNS]->(account)
      WITH user, account, owns IS NOT NULL as direct_owner
      OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
      WHERE ${testDate ? `datetime($testDate) >= datetime(poa.starts_at) AND datetime($testDate) <= datetime(poa.expires_at)` : `datetime() >= datetime(poa.starts_at) AND datetime() <= datetime(poa.expires_at)`}
      WITH user, account, direct_owner, count(poa) as valid_poa
      OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
      WITH user, account, direct_owner, valid_poa, count(role) as role_access
      OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
      WHERE ${testDate ? `datetime($testDate) >= datetime(acc.starts_at) AND datetime($testDate) <= datetime(acc.expires_at)` : `datetime() >= datetime(acc.starts_at) AND datetime() <= datetime(acc.expires_at)`}
      WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
      RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as canDownload
    `;
    const result = await session.run(query, { userId, accountId, testDate });
    return result.records[0].get('canDownload');
  } finally {
    await session.close();
  }
}

async function checkTransactionPermission(userId: string, accountId: string, testDate?: string): Promise<boolean> {
  const session = driver.session();
  try {
    // Smart query with business logic in data store (same as statement)
    const query = `
      MATCH (user:User {id: $userId})
      MATCH (account:Account {id: $accountId})
      WITH user, account
      OPTIONAL MATCH (user)-[owns:OWNS]->(account)
      WITH user, account, owns IS NOT NULL as direct_owner
      OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
      WHERE ${testDate ? `datetime($testDate) >= datetime(poa.starts_at) AND datetime($testDate) <= datetime(poa.expires_at)` : `datetime() >= datetime(poa.starts_at) AND datetime() <= datetime(poa.expires_at)`}
      WITH user, account, direct_owner, count(poa) as valid_poa
      OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
      WITH user, account, direct_owner, valid_poa, count(role) as role_access
      OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
      WHERE ${testDate ? `datetime($testDate) >= datetime(acc.starts_at) AND datetime($testDate) <= datetime(acc.expires_at)` : `datetime() >= datetime(acc.starts_at) AND datetime() <= datetime(acc.expires_at)`}
      WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
      RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as canView
    `;
    const result = await session.run(query, { userId, accountId, testDate });
    return result.records[0].get('canView');
  } finally {
    await session.close();
  }
}

async function testPermission(userId: string, accountId: string, testType: string, expected: boolean, description: string, testDate?: string, amount?: number) {
  try {
    let result: boolean;
    if (testType === 'payment') {
      result = await checkPaymentPermission(userId, accountId, amount || 1000, testDate);
    } else if (testType === 'statement') {
      result = await checkStatementPermission(userId, accountId, testDate);
    } else if (testType === 'transaction') {
      result = await checkTransactionPermission(userId, accountId, testDate);
    } else {
      console.log(`Unknown test type: ${testType}`);
      return;
    }

    if (result === expected) {
      console.log(`[PASS] ${description}`);
      pass_count++;
    } else {
      console.log(`[FAIL] ${description} (got ${result}, expected ${expected})`);
      fail_count++;
    }
  } catch (error) {
    console.log(`❌ ERROR: ${description} - ${error}`);
  }
}

async function runChecks() {
  try {
    console.log("== Emma's POA (Time + Limit) checks ==");
    
    // Emma can initiate payment under limit within timeframe (amount=3000, date=2025-03-15)
    await testPermission('emma', 'acc123', 'payment', true, 'Emma can initiate payment under limit within timeframe (amount=3000, date=2025-03-15)', '2025-03-15T00:00:00Z', 3000);
    
    // Emma can initiate payment at limit within timeframe (amount=5000, date=2025-06-15)
    await testPermission('emma', 'acc123', 'payment', true, 'Emma can initiate payment at limit within timeframe (amount=5000, date=2025-06-15)', '2025-06-15T00:00:00Z', 5000);
    
    // Emma cannot initiate payment above limit (amount=5001, date=2025-03-15)
    await testPermission('emma', 'acc123', 'payment', false, 'Emma cannot initiate payment above limit (amount=5001, date=2025-03-15)', '2025-03-15T00:00:00Z', 5001);
    
    // Emma cannot initiate payment before start date (amount=3000, date=2024-12-31)
    await testPermission('emma', 'acc123', 'payment', false, 'Emma cannot initiate payment before start date (amount=3000, date=2024-12-31)', '2024-12-31T23:59:59Z', 3000);
    
    // Emma cannot initiate payment after expiry (amount=3000, date=2025-07-01)
    await testPermission('emma', 'acc123', 'payment', false, 'Emma cannot initiate payment after expiry (amount=3000, date=2025-07-01)', '2025-07-01T00:00:00Z', 3000);

    console.log("== Adi's Accountant Access checks ==");
    
    // Adi can view transactions within timeframe (date=2025-02-15)
    await testPermission('adi', 'acc456', 'transaction', true, 'Adi can view transactions within timeframe (date=2025-02-15)', '2025-02-15T00:00:00Z');
    
    // Adi can download statement within timeframe (date=2025-03-30)
    await testPermission('adi', 'acc456', 'statement', true, 'Adi can download statement within timeframe (date=2025-03-30)', '2025-03-30T00:00:00Z');
    
    // Adi cannot view transactions before start date (date=2024-12-31)
    await testPermission('adi', 'acc456', 'transaction', false, 'Adi cannot view transactions before start date (date=2024-12-31)', '2024-12-31T23:59:59Z');
    
    // Adi cannot view transactions after expiry (date=2025-04-01)
    await testPermission('adi', 'acc456', 'transaction', false, 'Adi cannot view transactions after expiry (date=2025-04-01)', '2025-04-01T00:00:00Z');
    
    // Adi cannot download statement after expiry (date=2025-04-01)
    await testPermission('adi', 'acc456', 'statement', false, 'Adi cannot download statement after expiry (date=2025-04-01)', '2025-04-01T00:00:00Z');
    
    // Adi cannot initiate payment (accountant restriction)
    await testPermission('adi', 'acc456', 'payment', false, 'Adi cannot initiate payment (accountant restriction)', '2025-02-15T00:00:00Z');

    console.log("== POA Delegated Access checks ==");
    
    // Emma can view transactions via POA on acc123 (within timeframe)
    await testPermission('emma', 'acc123', 'transaction', true, 'Emma can view transactions via POA on acc123 (within timeframe)', '2025-03-15T00:00:00Z', 3000);
    
    // Emma can download statement via POA on acc123 (within timeframe)
    await testPermission('emma', 'acc123', 'statement', true, 'Emma can download statement via POA on acc123 (within timeframe)', '2025-03-15T00:00:00Z', 3000);
    
    // Emma cannot view transactions via POA on acc123 (after expiry)
    await testPermission('emma', 'acc123', 'transaction', false, 'Emma cannot view transactions via POA on acc123 (after expiry)', '2025-07-01T00:00:00Z', 3000);

    console.log("== Direct and role-based checks ==");
    
    // David can download statement for acc123 (owner)
    await testPermission('david', 'acc123', 'statement', true, 'David can download statement for acc123 (owner)');
    
    // Emma can download statement for acc456 (owner)
    await testPermission('emma', 'acc456', 'statement', true, 'Emma can download statement for acc456 (owner)');
    
    // Alice can_access org abc via finance_ops role (role membership)
    await testPermission('alice', 'acc789', 'statement', true, 'Alice can_access org abc via finance_ops role (role membership)');
    
    // Bob cannot download statement for acc123 (no relationship)
    await testPermission('bob', 'acc123', 'statement', false, 'Bob cannot download statement for acc123 (no relationship)');

    console.log("== Summary ==");
    console.log(`Passed: ${pass_count}`);
    console.log(`Failed: ${fail_count}`);
    if (fail_count === 0) {
      console.log("✅ All permission checks passed!");
    } else {
      console.log("❌ Some permission checks failed. Review output above.");
    }

  } catch (error) {
    console.error('Error running checks:', error);
  } finally {
    await driver.close();
  }
}

runChecks();