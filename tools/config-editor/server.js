const http = require('http');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');
const { spawn } = require('child_process');

const args = process.argv.slice(2);
let configDir = path.resolve(__dirname, '../../configs');
let port = 3000;

for (let i = 0; i < args.length; i++) {
  if ((args[i] === '--config' || args[i] === '-c') && args[i + 1]) {
    configDir = path.resolve(args[++i]);
  } else if ((args[i] === '--port' || args[i] === '-p') && args[i + 1]) {
    port = parseInt(args[++i], 10);
  }
}

if (process.env.CONFIG_DIR) configDir = path.resolve(process.env.CONFIG_DIR);
if (process.env.PORT) port = parseInt(process.env.PORT, 10);

const MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
};

const VALID_NAME = /^[a-zA-Z0-9_-]+$/;

const SCHEMA = {
  strategies: [
    'MOST_VOTED', 'HIGH_ODDS', 'PERCENTAGE', 'SMART_MONEY', 'SMART',
    'NUMBER_1', 'NUMBER_2', 'NUMBER_3', 'NUMBER_4',
    'NUMBER_5', 'NUMBER_6', 'NUMBER_7', 'NUMBER_8',
  ],
  chat_modes: ['ALWAYS', 'NEVER', 'ONLINE', 'OFFLINE'],
  priorities: ['STREAK', 'DROPS', 'ORDER', 'SUBSCRIBED', 'POINTS_ASCENDING', 'POINTS_DESCENDING'],
  followers_order: ['ASC', 'DESC'],
  delay_modes: ['FROM_START', 'FROM_END', 'PERCENTAGE'],
  filter_where: ['GT', 'LT', 'GTE', 'LTE'],
  filter_by: ['total_users', 'total_points'],
  webhook_methods: ['GET', 'POST'],
  notification_events: [
    'STREAMER_ONLINE', 'STREAMER_OFFLINE',
    'GAIN_FOR_RAID', 'GAIN_FOR_CLAIM', 'GAIN_FOR_WATCH', 'GAIN_FOR_WATCH_STREAK',
    'BET_WIN', 'BET_LOSE', 'BET_REFUND', 'BET_FILTERS', 'BET_GENERAL', 'BET_FAILED', 'BET_START',
    'BONUS_CLAIM', 'MOMENT_CLAIM', 'JOIN_RAID',
    'DROP_CLAIM', 'DROP_STATUS',
    'CAMPAIGN_STARTED', 'CAMPAIGN_REMINDER',
    'CHAT_MENTION', 'GIFTED_SUB',
    'MINER_STARTED', 'MINER_STOPPED', 'MINER_CRASHED',
    'TEST',
  ],
  defaults: {
    max_watch_streams: 2,
    priority: ['STREAK', 'DROPS', 'ORDER'],
    category_watcher_poll_interval: '120s',
    team_watcher_poll_interval: '120s',
    followers_order: 'ASC',
  },
};

function sendJSON(res, status, data) {
  const body = JSON.stringify(data, null, 2);
  res.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' });
  res.end(body);
}

function sendError(res, status, message) {
  sendJSON(res, status, { error: message });
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => {
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString()));
      } catch {
        reject(new Error('Invalid JSON body'));
      }
    });
    req.on('error', reject);
  });
}

function configPath(name) {
  return path.join(configDir, name + '.yaml');
}

function stripSecrets(config) {
  const c = { ...config };
  if (c.notifications) {
    const n = { ...c.notifications };
    const secretFields = {
      telegram: ['token', 'chat_id'],
      discord: ['webhook_url'],
      webhook: ['endpoint'],
      matrix: ['homeserver', 'room_id', 'access_token'],
      pushover: ['api_token', 'user_key'],
      gotify: ['url', 'token'],
    };
    for (const [provider, fields] of Object.entries(secretFields)) {
      if (n[provider]) {
        n[provider] = { ...n[provider] };
        for (const f of fields) delete n[provider][f];
      }
    }
    c.notifications = n;
  }
  return c;
}

function loadConfig(name) {
  const filePath = configPath(name);
  const raw = fs.readFileSync(filePath, 'utf-8');
  return yaml.load(raw) || {};
}

function listAccounts() {
  if (!fs.existsSync(configDir)) return [];
  return fs.readdirSync(configDir)
    .filter((f) => f.endsWith('.yaml') || f.endsWith('.yml'))
    .filter((f) => !f.endsWith('.example'))
    .map((f) => {
      const name = f.replace(/\.ya?ml$/, '');
      try {
        const config = loadConfig(name);
        return {
          name,
          enabled: config.enabled !== undefined ? config.enabled : true,
          streamer_count: (config.streamers || []).length,
          has_category_watcher: !!(config.category_watcher && config.category_watcher.enabled),
          has_team_watcher: !!(config.team_watcher && config.team_watcher.enabled),
          has_followers: !!(config.followers && config.followers.enabled),
          has_notifications: !!(config.notifications && Object.keys(config.notifications).some(
            (k) => config.notifications[k] && config.notifications[k].enabled
          )),
        };
      } catch {
        return { name, enabled: false, error: true };
      }
    });
}

function validateConfig(config) {
  const errors = [];

  if (config.max_watch_streams !== undefined && config.max_watch_streams < 1) {
    errors.push('max_watch_streams must be at least 1');
  }

  const hasStreamers = config.streamers && config.streamers.length > 0;
  const hasFollowers = config.followers && config.followers.enabled;
  const hasCatWatcher = config.category_watcher && config.category_watcher.enabled;
  const hasTeamWatcher = config.team_watcher && config.team_watcher.enabled;

  if (!hasStreamers && !hasFollowers && !hasCatWatcher && !hasTeamWatcher) {
    errors.push('At least one of streamers, followers, category_watcher, or team_watcher must be configured');
  }

  if (config.streamers) {
    config.streamers.forEach((s, i) => {
      if (!s.username || !s.username.trim()) {
        errors.push(`Streamer at index ${i} has empty username`);
      }
    });
  }

  if (config.category_watcher && config.category_watcher.enabled) {
    if (!config.category_watcher.categories || config.category_watcher.categories.length === 0) {
      errors.push('category_watcher is enabled but no categories are configured');
    }
  }

  if (config.team_watcher && config.team_watcher.enabled) {
    if (!config.team_watcher.teams || config.team_watcher.teams.length === 0) {
      errors.push('team_watcher is enabled but no teams are configured');
    }
  }

  const defaults = config.streamer_defaults;
  if (defaults && defaults.make_predictions === true && !defaults.bet) {
    errors.push('make_predictions is enabled in streamer_defaults but no bet config is set');
  }

  return errors;
}

function cleanConfig(config) {
  const removeEmpty = (obj) => {
    if (obj === null || obj === undefined) return undefined;
    if (typeof obj !== 'object') return obj;
    if (Array.isArray(obj)) {
      const arr = obj.map(removeEmpty).filter((v) => v !== undefined);
      return arr.length > 0 ? arr : undefined;
    }
    const cleaned = {};
    for (const [k, v] of Object.entries(obj)) {
      if (v === null || v === undefined || v === '') continue;
      const val = removeEmpty(v);
      if (val !== undefined) cleaned[k] = val;
    }
    return Object.keys(cleaned).length > 0 ? cleaned : undefined;
  };
  return removeEmpty(config) || {};
}

function atomicWrite(filePath, content) {
  const tmp = filePath + '.tmp';
  fs.writeFileSync(tmp, content, 'utf-8');
  fs.renameSync(tmp, filePath);
}

function mergeSecretsBack(newConfig, existingConfig) {
  if (!existingConfig.notifications) return newConfig;
  if (!newConfig.notifications) return newConfig;

  const secretFields = {
    telegram: ['token', 'chat_id'],
    discord: ['webhook_url'],
    webhook: ['endpoint'],
    matrix: ['homeserver', 'room_id', 'access_token'],
    pushover: ['api_token', 'user_key'],
    gotify: ['url', 'token'],
  };

  const merged = { ...newConfig, notifications: { ...newConfig.notifications } };
  for (const [provider, fields] of Object.entries(secretFields)) {
    if (merged.notifications[provider] && existingConfig.notifications[provider]) {
      merged.notifications[provider] = { ...merged.notifications[provider] };
      for (const f of fields) {
        if (existingConfig.notifications[provider][f] !== undefined) {
          merged.notifications[provider][f] = existingConfig.notifications[provider][f];
        }
      }
    }
  }
  return merged;
}

function serveStatic(req, res) {
  let filePath;
  const urlPath = req.url.split('?')[0];
  if (urlPath === '/' || urlPath === '/index.html') {
    filePath = path.join(__dirname, 'public', 'index.html');
  } else {
    const safePath = path.normalize(urlPath).replace(/^(\.\.[/\\])+/, '');
    filePath = path.join(__dirname, 'public', safePath);
    if (!filePath.startsWith(path.join(__dirname, 'public'))) {
      res.writeHead(403);
      res.end('Forbidden');
      return;
    }
  }

  const ext = path.extname(filePath);
  const mime = MIME_TYPES[ext] || 'application/octet-stream';

  try {
    const content = fs.readFileSync(filePath);
    res.writeHead(200, { 'Content-Type': mime });
    res.end(content);
  } catch {
    res.writeHead(404);
    res.end('Not found');
  }
}

async function handleRequest(req, res) {
  const url = new URL(req.url, `http://${req.headers.host}`);
  const pathname = url.pathname;

  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  if (req.method === 'OPTIONS') { res.writeHead(204); res.end(); return; }

  try {
    if (pathname === '/api/schema' && req.method === 'GET') {
      sendJSON(res, 200, SCHEMA);
      return;
    }

    if (pathname === '/api/accounts' && req.method === 'GET') {
      sendJSON(res, 200, listAccounts());
      return;
    }

    if (pathname === '/api/accounts' && req.method === 'POST') {
      const body = await readBody(req);
      const name = body.name;
      if (!name || !VALID_NAME.test(name)) {
        sendError(res, 400, 'Invalid account name. Use only letters, numbers, underscores, and hyphens.');
        return;
      }
      if (fs.existsSync(configPath(name))) {
        sendError(res, 409, `Account "${name}" already exists`);
        return;
      }
      const config = body.config || {};
      const errors = validateConfig(config);
      if (errors.length > 0) {
        sendJSON(res, 422, { errors });
        return;
      }
      const cleaned = cleanConfig(config);
      const yamlStr = yaml.dump(cleaned, { indent: 2, lineWidth: -1, noRefs: true, sortKeys: false });
      atomicWrite(configPath(name), yamlStr);
      sendJSON(res, 201, { name, message: 'Account created' });
      return;
    }

    const accountMatch = pathname.match(/^\/api\/accounts\/([^/]+)$/);
    if (accountMatch) {
      const name = decodeURIComponent(accountMatch[1]);
      if (!VALID_NAME.test(name)) {
        sendError(res, 400, 'Invalid account name');
        return;
      }

      if (req.method === 'GET') {
        if (!fs.existsSync(configPath(name))) {
          sendError(res, 404, `Account "${name}" not found`);
          return;
        }
        const config = loadConfig(name);
        sendJSON(res, 200, stripSecrets(config));
        return;
      }

      if (req.method === 'PUT') {
        if (!fs.existsSync(configPath(name))) {
          sendError(res, 404, `Account "${name}" not found`);
          return;
        }
        const body = await readBody(req);
        const config = body.config || body;
        const errors = validateConfig(config);
        if (errors.length > 0) {
          sendJSON(res, 422, { errors });
          return;
        }

        const existing = loadConfig(name);
        const merged = mergeSecretsBack(config, existing);
        const cleaned = cleanConfig(merged);
        const yamlStr = yaml.dump(cleaned, { indent: 2, lineWidth: -1, noRefs: true, sortKeys: false });
        atomicWrite(configPath(name), yamlStr);
        sendJSON(res, 200, { name, message: 'Account updated' });
        return;
      }

      if (req.method === 'DELETE') {
        if (!fs.existsSync(configPath(name))) {
          sendError(res, 404, `Account "${name}" not found`);
          return;
        }
        fs.unlinkSync(configPath(name));
        sendJSON(res, 200, { name, message: 'Account deleted' });
        return;
      }
    }

    serveStatic(req, res);
  } catch (err) {
    console.error('Request error:', err);
    sendError(res, 500, err.message || 'Internal server error');
  }
}

function openBrowser(url) {
  const cmd = process.platform === 'win32' ? 'cmd'
    : process.platform === 'darwin' ? 'open' : 'xdg-open';
  const cmdArgs = process.platform === 'win32' ? ['/c', 'start', url] : [url];
  try {
    spawn(cmd, cmdArgs, { stdio: 'ignore', detached: true }).unref();
  } catch {}
}

const server = http.createServer(handleRequest);
server.listen(port, () => {
  const url = `http://localhost:${port}`;
  console.log(`\n  Config Editor running at ${url}`);
  console.log(`  Config directory: ${configDir}\n`);
  openBrowser(url);
});
