import { ApolloServer } from 'apollo-server';
import { readFileSync } from 'fs';
import { join } from 'path';
import { PrismaClient } from '@prisma/client';
import resolvers from './entitlement-resolvers';

const prisma = new PrismaClient();

const typeDefs = readFileSync(join(__dirname, '../schema.graphql'), 'utf8');

const server = new ApolloServer({
  typeDefs,
  resolvers,
  context: ({ req }) => {
    // For demo: extract userId from headers (simulate auth)
    const userId = req.headers['x-user-id'] as string | undefined;
    return { prisma, userId };
  },
});

server.listen({ port: 4000 }).then(({ url }) => {
  console.log(`🚀 GraphQL Entitlement server ready at ${url}`);
}); 