import { PrismaClient } from '@prisma/client';

interface Context {
  prisma: PrismaClient;
  userId?: string;
}

async function userHasOrgAccess(prisma: PrismaClient, userId: string, orgId: string): Promise<boolean> {
  const org = await prisma.organization.findUnique({
    where: { id: orgId },
    include: { members: true, roles: { include: { members: true } } },
  });
  if (!org) return false;
  if (org.members.some(u => u.id === userId)) return true;
  for (const role of org.roles) {
    if (role.members.some(u => u.id === userId)) return true;
  }
  return false;
}

async function userHasPOAAccess(account: any, userId: string, amount?: number): Promise<boolean> {
  const now = new Date();
  for (const poa of account.poas ?? []) {
    if (poa.delegateId === userId) {
      if (poa.validFrom && poa.validTo && (now < poa.validFrom || now > poa.validTo)) continue;
      if (poa.maxAmount != null && amount != null && amount > poa.maxAmount) continue;
      return true;
    }
  }
  return false;
}

async function userHasAccountantAccess(prisma: PrismaClient, accountId: string, userId: string): Promise<boolean> {
  const now = new Date();
  const access = await prisma.accountantAccess.findFirst({
    where: {
      accountId,
      userId,
      validFrom: { lte: now },
      validTo: { gte: now },
    },
  });
  return !!access;
}

const resolvers = {
  Query: {
    me: async (parent: any, args: any, context: Context) => {
      if (!context.userId) return null;
      return context.prisma.user.findUnique({ where: { id: context.userId } });
    },
    account: async (parent: any, args: { id: string }, context: Context) => {
      return context.prisma.account.findUnique({ where: { id: args.id } });
    },
    statement: async (parent: any, args: { id: string }, context: Context) => {
      return context.prisma.statement.findUnique({ where: { id: args.id }, include: { account: true } });
    },
    canInitiatePayment: async (parent: any, args: { accountId: string, amount: number }, context: Context) => {
      if (!context.userId) return false;
      const { prisma, userId } = context;
      const { accountId, amount } = args;
      const account = await prisma.account.findUnique({
        where: { id: accountId },
        include: { ownerUser: true, ownerOrg: true, delegates: true, poas: true },
      });
      if (!account) return false;
      if (account.ownerUser && account.ownerUser.id === userId) return true;
      if (account.ownerOrg && await userHasOrgAccess(prisma, userId, account.ownerOrg.id)) return true;
      if (account.delegates && Array.isArray(account.delegates) && account.delegates.some((u: any) => u.id === userId)) return true;
      if (await userHasPOAAccess(account, userId, amount)) return true;
      // Accountants cannot initiate payments - only view/download statements
      return false;
    },
    canDownloadStatement: async (parent: any, args: { accountId: string }, context: Context) => {
      if (!context.userId) return false;
      const { prisma, userId } = context;
      const account = await prisma.account.findUnique({
        where: { id: args.accountId },
        include: { ownerUser: true, ownerOrg: true, delegates: true, poas: true },
      });
      if (!account) return false;
      if (account.ownerUser && account.ownerUser.id === userId) return true;
      if (account.ownerOrg && await userHasOrgAccess(prisma, userId, account.ownerOrg.id)) return true;
      if (account.delegates && Array.isArray(account.delegates) && account.delegates.some((u: any) => u.id === userId)) return true;
      if (await userHasPOAAccess(account, userId)) return true;
      if (await userHasAccountantAccess(prisma, args.accountId, userId)) return true;
      return false;
    },
    canViewTransactions: async (parent: any, args: { accountId: string }, context: Context) => {
      if (!context.userId) return false;
      const { prisma, userId } = context;
      const account = await prisma.account.findUnique({
        where: { id: args.accountId },
        include: { ownerUser: true, ownerOrg: true, delegates: true, poas: true },
      });
      if (!account) return false;
      if (account.ownerUser && account.ownerUser.id === userId) return true;
      if (account.ownerOrg && await userHasOrgAccess(prisma, userId, account.ownerOrg.id)) return true;
      if (account.delegates && Array.isArray(account.delegates) && account.delegates.some((u: any) => u.id === userId)) return true;
      if (await userHasPOAAccess(account, userId)) return true;
      if (await userHasAccountantAccess(prisma, args.accountId, userId)) return true;
      return false;
    },
    canAccess: async (parent: any, args: { accountId: string }, context: Context) => {
      if (!context.userId) return false;
      const { prisma, userId } = context;
      const account = await prisma.account.findUnique({
        where: { id: args.accountId },
        include: { ownerUser: true, ownerOrg: true, delegates: true, poas: true },
      });
      if (!account) return false;
      if (account.ownerUser && account.ownerUser.id === userId) return true;
      if (account.ownerOrg && await userHasOrgAccess(prisma, userId, account.ownerOrg.id)) return true;
      return false;
    },
  },
  User: {
    roles: async (parent: any, args: any, context: Context) => {
      const user = await context.prisma.user.findUnique({ where: { id: parent.id }, include: { roles: true } });
      return user && Array.isArray(user.roles) ? user.roles : [];
    },
    orgs: async (parent: any, args: any, context: Context) => {
      const user = await context.prisma.user.findUnique({ where: { id: parent.id }, include: { orgs: true } });
      return user && Array.isArray(user.orgs) ? user.orgs : [];
    },
  },
  Organization: {
    members: async (parent: any, args: any, context: Context) => {
      const org = await context.prisma.organization.findUnique({ where: { id: parent.id }, include: { members: true } });
      return org && Array.isArray(org.members) ? org.members : [];
    },
    roles: async (parent: any, args: any, context: Context) => {
      const org = await context.prisma.organization.findUnique({ where: { id: parent.id }, include: { roles: true } });
      return org && Array.isArray(org.roles) ? org.roles : [];
    },
    accounts: async (parent: any, args: any, context: Context) => {
      const org = await context.prisma.organization.findUnique({ where: { id: parent.id }, include: { accounts: true } });
      return org && Array.isArray(org.accounts) ? org.accounts : [];
    },
  },
  Role: {
    members: async (parent: any, args: any, context: Context) => {
      const role = await context.prisma.role.findUnique({ where: { id: parent.id }, include: { members: true } });
      return role && Array.isArray(role.members) ? role.members : [];
    },
    org: async (parent: any, args: any, context: Context) => {
      const role = await context.prisma.role.findUnique({ where: { id: parent.id }, include: { org: true } });
      return role && role.org ? role.org : null;
    },
  },
  Account: {
    ownerUser: async (parent: any, args: any, context: Context) => {
      const account = await context.prisma.account.findUnique({ where: { id: parent.id }, include: { ownerUser: true } });
      return account && account.ownerUser ? account.ownerUser : null;
    },
    ownerOrg: async (parent: any, args: any, context: Context) => {
      const account = await context.prisma.account.findUnique({ where: { id: parent.id }, include: { ownerOrg: true } });
      return account && account.ownerOrg ? account.ownerOrg : null;
    },
    delegates: async (parent: any, args: any, context: Context) => {
      const account = await context.prisma.account.findUnique({ where: { id: parent.id }, include: { delegates: true } });
      return account && Array.isArray(account.delegates) ? account.delegates : [];
    },
    poas: async (parent: any, args: any, context: Context) => {
      const account = await context.prisma.account.findUnique({ where: { id: parent.id }, include: { poas: true } });
      return account && Array.isArray(account.poas) ? account.poas : [];
    },
  },
  POA: {
    account: async (parent: any, args: any, context: Context) => {
      const poa = await context.prisma.pOA.findUnique({ where: { id: parent.id }, include: { account: true } });
      return poa && poa.account ? poa.account : null;
    },
    delegate: async (parent: any, args: any, context: Context) => {
      const poa = await context.prisma.pOA.findUnique({ where: { id: parent.id }, include: { delegate: true } });
      return poa && poa.delegate ? poa.delegate : null;
    },
  },
};

export default resolvers; 