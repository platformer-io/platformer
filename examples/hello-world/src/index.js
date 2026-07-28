'use strict';

const START_TIME = Date.now();

// ── Route handlers ────────────────────────────────────────────────────────────

function home(event) {
  const fn  = process.env.AWS_LAMBDA_FUNCTION_NAME || '—';
  const mem = process.env.AWS_LAMBDA_FUNCTION_MEMORY_SIZE || '—';
  const rgn = process.env.AWS_REGION || '—';
  const env = process.env.APP_ENV || 'development';
  const ver = process.env.APP_VERSION || '1.0.0';
  const up  = ((Date.now() - START_TIME) / 1000).toFixed(1);

  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>PlatFormer Demo</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body   { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
             background: #0f172a; color: #e2e8f0; min-height: 100vh; }
    header { background: #1e293b; border-bottom: 1px solid #334155;
             padding: 1rem 2rem; display: flex; align-items: center; gap: .75rem; }
    h1     { font-size: 1.1rem; font-weight: 600; letter-spacing: -.01em; }
    .pill  { background: #3b82f6; color: #fff; font-size: .65rem;
             padding: 2px 8px; border-radius: 999px; font-weight: 600; }
    main   { max-width: 760px; margin: 2.5rem auto; padding: 0 1.5rem; display: grid; gap: 1.25rem; }
    .card  { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 1.25rem 1.5rem; }
    h2     { font-size: .7rem; text-transform: uppercase; letter-spacing: .08em;
             color: #64748b; margin-bottom: .9rem; }
    dl     { display: grid; grid-template-columns: max-content 1fr; gap: .35rem 1.5rem; }
    dt     { color: #94a3b8; font-size: .85rem; }
    dd     { font-size: .85rem; font-family: 'SF Mono', ui-monospace, monospace; }
    .badge { display: inline-block; padding: 1px 8px; border-radius: 6px;
             font-size: .75rem; font-weight: 600; background: #166534; color: #86efac; }
    nav    { display: flex; flex-wrap: wrap; gap: .75rem; }
    nav a  { background: #334155; color: #cbd5e1; text-decoration: none;
             padding: .45rem 1rem; border-radius: 8px; font-size: .85rem;
             display: flex; flex-direction: column; gap: .15rem; }
    nav a code { font-size: .75rem; color: #94a3b8; }
    nav a:hover { background: #475569; color: #e2e8f0; }
    footer { text-align: center; color: #475569; font-size: .75rem; margin-top: 2rem; padding-bottom: 2rem; }
  </style>
</head>
<body>
  <header>
    <h1>PlatFormer Demo</h1>
    <span class="pill">serverless</span>
    <span class="pill" style="background:#0d9488">${env}</span>
  </header>
  <main>
    <div class="card">
      <h2>Runtime</h2>
      <dl>
        <dt>Function</dt>         <dd>${fn}</dd>
        <dt>Region</dt>           <dd>${rgn}</dd>
        <dt>Memory</dt>           <dd>${mem} MB</dd>
        <dt>Version</dt>          <dd><span class="badge">${ver}</span></dd>
        <dt>Container uptime</dt> <dd>${up}s</dd>
      </dl>
    </div>
    <div class="card">
      <h2>Request</h2>
      <dl>
        <dt>Path</dt>   <dd>${event.rawPath || '/'}</dd>
        <dt>Method</dt> <dd>${event.requestContext?.http?.method || 'GET'}</dd>
        <dt>Time</dt>   <dd>${new Date().toISOString()}</dd>
      </dl>
    </div>
    <div class="card">
      <h2>Endpoints</h2>
      <nav>
        <a href="/"><code>GET /</code>Home</a>
        <a href="/health"><code>GET /health</code>Health check</a>
        <a href="/api/info"><code>GET /api/info</code>Runtime info (JSON)</a>
      </nav>
    </div>
  </main>
  <footer>Deployed with <strong>PlatFormer</strong></footer>
</body>
</html>`;

  return {
    statusCode: 200,
    headers: { 'Content-Type': 'text/html; charset=utf-8' },
    body: html,
  };
}

function health() {
  return {
    statusCode: 200,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      status:    'ok',
      uptimeMs:  Date.now() - START_TIME,
      timestamp: new Date().toISOString(),
    }),
  };
}

function apiInfo(event) {
  return {
    statusCode: 200,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      function: {
        name:     process.env.AWS_LAMBDA_FUNCTION_NAME,
        region:   process.env.AWS_REGION,
        memoryMB: Number(process.env.AWS_LAMBDA_FUNCTION_MEMORY_SIZE),
      },
      app: {
        env:     process.env.APP_ENV,
        version: process.env.APP_VERSION,
      },
      request: {
        path:      event.rawPath,
        method:    event.requestContext?.http?.method,
        userAgent: event.headers?.['user-agent'] ?? null,
      },
      timestamp: new Date().toISOString(),
    }, null, 2),
  };
}

function notFound(path) {
  return {
    statusCode: 404,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ error: 'Not found', path }),
  };
}

// ── Entry point ───────────────────────────────────────────────────────────────

exports.handler = async (event) => {
  const path   = event.rawPath || '/';
  const method = event.requestContext?.http?.method || 'GET';
  console.log(`${method} ${path}`);

  if (method === 'GET') {
    switch (path) {
      case '/':         return home(event);
      case '/health':   return health();
      case '/api/info': return apiInfo(event);
    }
  }

  return notFound(path);
};
