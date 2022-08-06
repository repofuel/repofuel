const {createProxyMiddleware} = require('http-proxy-middleware');

module.exports = function (app) {
  app.use(
    '/ingest/',
    createProxyMiddleware({
      target: 'http://localhost:3002',
      // changeOrigin: true,
      ws: true,
    })
  );
  app.use(
    '/ingest/',
    createProxyMiddleware({
      target: 'http://localhost:3002',
      changeOrigin: true,
    })
  );
  app.use(
    '/accounts/',
    createProxyMiddleware({
      target: 'http://localhost:3003',
      changeOrigin: true,
    })
  );
  app.use(
    '/ml/',
    createProxyMiddleware({
      target: 'http://localhost:3004',
      changeOrigin: true,
    })
  );
};
