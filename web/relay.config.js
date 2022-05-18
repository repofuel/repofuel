module.exports = {
  src: './src',
  schema: './src/graphql/schema.graphql',
  language: 'typescript',
  exclude: ['**/node_modules/**', '**/__mocks__/**', '**/__generated__/**'],
  customScalars: {
    DateTime: 'String',
  },
};
