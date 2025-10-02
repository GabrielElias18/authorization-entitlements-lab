-- Canonical SQL seed for entitlement model (aligned with SpiceDB and GraphDB)
-- Drop tables if they exist (for idempotency in dev)
TRUNCATE "AccountantAccess", "POA", "Payment", "Statement", "Account", "User", "Role", "Organization" RESTART IDENTITY CASCADE;

-- Organizations (aligned with canonical data)
INSERT INTO "Organization" (id, name, "createdAt") VALUES
  ('abc', 'abc', NOW());

-- Roles (aligned with canonical data)
INSERT INTO "Role" (id, name, "orgId", "createdAt") VALUES
  ('finance_ops', 'finance_ops', 'abc', NOW());

-- Users (canonical test users)
INSERT INTO "User" (id, name, email, "createdAt") VALUES
  ('david', 'david', 'david@example.com', NOW()),
  ('emma', 'emma', 'emma@example.com', NOW()),
  ('adi', 'adi', 'adi@example.com', NOW()),
  ('alice', 'alice', 'alice@example.com', NOW()),
  ('bob', 'bob', 'bob@example.com', NOW()),
  ('charlie', 'charlie', 'charlie@example.com', NOW());

-- Role membership (Alice has finance_ops role for org access)
INSERT INTO "_RoleMembers" ("A", "B") VALUES
  ('finance_ops', 'alice');

-- Accounts (canonical ownership structure)
INSERT INTO "Account" (id, "ownerUserId", "ownerOrgId", balance, "createdAt") VALUES
  ('acc123', 'david', NULL, 10000, NOW()),          -- David owns acc123
  ('acc456', 'emma', NULL, 20000, NOW()),           -- Emma owns acc456
  ('acc789', NULL, 'abc', 30000, NOW()),            -- Org abc owns acc789
  ('acc790', NULL, 'abc', 40000, NOW()),            -- Org abc owns acc790
  ('acc791', NULL, 'abc', 50000, NOW()),            -- Org abc owns acc791
  ('acc792', NULL, 'abc', 60000, NOW());            -- Org abc owns acc792

-- POA: Emma has POA on David's account (acc123) with payment limit
INSERT INTO "POA" (id, "accountId", "delegateId", "validFrom", "validTo", "maxAmount", "createdAt") VALUES
  ('poa1', 'acc123', 'emma', '2025-01-01', '2025-12-31', 5000, NOW());

-- Accountant access: Adi has time-bound access to Emma's account (acc456)
INSERT INTO "AccountantAccess" (id, "accountId", "userId", "validFrom", "validTo", "createdAt") VALUES
  ('acct1', 'acc456', 'adi', '2025-01-01', '2025-12-31', NOW());

-- Note: Bob and Charlie have no relationships - used for negative testing 