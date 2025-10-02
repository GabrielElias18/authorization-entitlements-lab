// Node constraints
CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE;
CREATE CONSTRAINT account_id IF NOT EXISTS FOR (a:Account) REQUIRE a.id IS UNIQUE;
CREATE CONSTRAINT org_id IF NOT EXISTS FOR (o:Org) REQUIRE o.id IS UNIQUE;
CREATE CONSTRAINT role_id IF NOT EXISTS FOR (r:Role) REQUIRE r.id IS UNIQUE;

// Indexes for performance
CREATE INDEX user_id_index IF NOT EXISTS FOR (u:User) ON (u.id);
CREATE INDEX account_id_index IF NOT EXISTS FOR (a:Account) ON (a.id);
CREATE INDEX org_id_index IF NOT EXISTS FOR (o:Org) ON (o.id);
CREATE INDEX role_id_index IF NOT EXISTS FOR (r:Role) ON (r.id); 