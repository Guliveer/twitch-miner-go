(function () {
  'use strict';

  let currentAccount = null;
  let accounts = [];
  let schema = null;
  let isDirty = false;

  const NOTIFICATION_PROVIDERS = ['telegram', 'discord', 'webhook', 'matrix', 'pushover', 'gotify'];

  // ─── API ───

  async function api(method, path, body) {
    const opts = { method, headers: { 'Content-Type': 'application/json' } };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch('/api' + path, opts);
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.errors?.join(', ') || 'Request failed');
    return data;
  }

  async function fetchAccounts() { return api('GET', '/accounts'); }
  async function fetchSchema() { return api('GET', '/schema'); }
  async function fetchConfig(name) { return api('GET', '/accounts/' + encodeURIComponent(name)); }
  async function saveConfig(name, config) { return api('PUT', '/accounts/' + encodeURIComponent(name), { config }); }
  async function createConfig(name, config) { return api('POST', '/accounts', { name, config }); }
  async function deleteConfig(name) { return api('DELETE', '/accounts/' + encodeURIComponent(name)); }

  // ─── DOM Helpers ───

  function el(tag, attrs, children) {
    const node = document.createElement(tag);
    if (attrs) {
      for (const [k, v] of Object.entries(attrs)) {
        if (k === 'className') node.className = v;
        else if (k === 'dataset') Object.assign(node.dataset, v);
        else if (k.startsWith('on')) node[k] = v;
        else node.setAttribute(k, v);
      }
    }
    if (children) {
      if (typeof children === 'string') node.textContent = children;
      else if (Array.isArray(children)) children.forEach((c) => { if (c) node.appendChild(c); });
      else node.appendChild(children);
    }
    return node;
  }

  // ─── Toast ───

  function showToast(message, type) {
    const container = document.getElementById('toast-container');
    const toast = el('div', { className: 'toast ' + type }, message);
    container.appendChild(toast);
    setTimeout(() => {
      toast.classList.add('fade-out');
      setTimeout(() => toast.remove(), 300);
    }, 3000);
  }

  // ─── Modal ───

  function showModal(title, bodyContent, buttons) {
    document.getElementById('modal-title').textContent = title;
    const bodyEl = document.getElementById('modal-body');
    bodyEl.textContent = '';
    if (typeof bodyContent === 'string') bodyEl.textContent = bodyContent;
    else bodyEl.appendChild(bodyContent);

    const footer = document.getElementById('modal-footer');
    footer.textContent = '';
    buttons.forEach((b) => {
      const btn = el('button', { className: 'btn ' + (b.class || 'btn-ghost'), onclick: () => { hideModal(); if (b.action) b.action(); } }, b.label);
      footer.appendChild(btn);
    });
    document.getElementById('modal-overlay').classList.remove('hidden');
  }

  function hideModal() {
    document.getElementById('modal-overlay').classList.add('hidden');
  }

  // ─── Sidebar ───

  function renderSidebar() {
    const list = document.getElementById('account-list');
    list.textContent = '';
    if (accounts.length === 0) {
      list.appendChild(el('li', { className: 'loading-placeholder' }, 'No accounts found'));
      return;
    }
    accounts.forEach((a) => {
      const meta = [];
      if (a.streamer_count > 0) meta.push(a.streamer_count + ' streamers');
      if (a.has_category_watcher) meta.push('CW');
      if (a.has_team_watcher) meta.push('TW');
      if (a.has_followers) meta.push('FL');

      const li = el('li', {
        className: 'account-item' + (currentAccount === a.name ? ' active' : ''),
        onclick: () => selectAccount(a.name),
      }, [
        el('span', { className: 'dot ' + (a.enabled ? 'enabled' : 'disabled') }),
        el('span', { className: 'account-name' }, a.name),
        el('span', { className: 'account-meta' }, meta.join(' ')),
      ]);
      list.appendChild(li);
    });
  }

  async function selectAccount(name) {
    if (isDirty && currentAccount !== name) {
      const ok = confirm('You have unsaved changes. Discard?');
      if (!ok) return;
    }
    currentAccount = name;
    isDirty = false;
    renderSidebar();
    try {
      const config = await fetchConfig(name);
      renderEditor(name, config);
    } catch (err) {
      showToast('Failed to load config: ' + err.message, 'error');
    }
  }

  // ─── Editor ───

  function renderEditor(name, config) {
    document.getElementById('editor-placeholder').classList.add('hidden');
    document.getElementById('editor-content').classList.remove('hidden');
    document.getElementById('editor-title').textContent = name;

    const badges = document.getElementById('editor-badges');
    badges.textContent = '';
    if (config.enabled !== false) {
      badges.appendChild(el('span', { className: 'badge badge-enabled' }, 'Enabled'));
    }

    setChecked('cfg-enabled', config.enabled !== false);
    setVal('cfg-max-watch', config.max_watch_streams || schema.defaults.max_watch_streams);
    setVal('cfg-proxy', config.proxy || '');

    renderChipSelect('cfg-priority', schema.priorities, config.priority || schema.defaults.priority);

    setChecked('cfg-claim-drops-startup', config.features?.claim_drops_startup || false);
    setChecked('cfg-enable-analytics', config.features?.enable_analytics || false);

    const cw = config.category_watcher || {};
    setChecked('cfg-cw-enabled', cw.enabled || false);
    setVal('cfg-cw-interval', cw.poll_interval || schema.defaults.category_watcher_poll_interval);
    setChecked('cfg-cw-drops-only', cw.drops_only || false);
    renderTagList('cfg-cw-reminders', cw.campaign_reminders || []);
    renderCategories(cw.categories || []);

    const tw = config.team_watcher || {};
    setChecked('cfg-tw-enabled', tw.enabled || false);
    setVal('cfg-tw-interval', tw.poll_interval || schema.defaults.team_watcher_poll_interval);
    renderTeams(tw.teams || []);

    const sd = config.streamer_defaults || {};
    renderTriToggle('cfg-sd-predictions', sd.make_predictions);
    renderTriToggle('cfg-sd-follow-raid', sd.follow_raid);
    renderTriToggle('cfg-sd-claim-drops', sd.claim_drops);
    renderTriToggle('cfg-sd-claim-moments', sd.claim_moments);
    renderTriToggle('cfg-sd-watch-streak', sd.watch_streak);
    renderTriToggle('cfg-sd-community-goals', sd.community_goals);
    setVal('cfg-sd-chat', sd.chat || '');

    const bet = sd.bet || {};
    setVal('cfg-bet-strategy', bet.strategy || '');
    setVal('cfg-bet-percentage', bet.percentage ?? '');
    setVal('cfg-bet-gap', bet.percentage_gap ?? '');
    setVal('cfg-bet-max', bet.max_points ?? '');
    setVal('cfg-bet-min', bet.minimum_points ?? '');
    renderTriToggle('cfg-bet-stealth', bet.stealth_mode);
    setVal('cfg-bet-delay', bet.delay ?? '');
    setVal('cfg-bet-delay-mode', bet.delay_mode || '');

    const fc = bet.filter_condition || {};
    setVal('cfg-bet-filter-by', fc.by || '');
    setVal('cfg-bet-filter-where', fc.where || '');
    setVal('cfg-bet-filter-value', fc.value ?? '');

    renderStreamers(config.streamers || []);
    document.getElementById('streamers-count').textContent = (config.streamers || []).length;

    renderTagList('cfg-blacklist', config.blacklist || []);
    renderTagList('cfg-cat-blacklist', config.category_blacklist || []);

    const fol = config.followers || {};
    setChecked('cfg-fol-enabled', fol.enabled || false);
    setVal('cfg-fol-order', fol.order || schema.defaults.followers_order);

    const notif = config.notifications || {};
    const batch = notif.batch || {};
    renderTriToggle('cfg-batch-enabled', batch.enabled);
    setVal('cfg-batch-interval', batch.interval || '');
    setVal('cfg-batch-max', batch.max_entries ?? '');
    renderChipSelect('cfg-batch-immediate', schema.notification_events, batch.immediate_events || []);

    renderNotificationProviders(notif);

    isDirty = false;
  }

  // ─── Collect Form Data ───

  function collectConfig() {
    const config = {};

    const enabled = getChecked('cfg-enabled');
    if (!enabled) config.enabled = false;

    const maxWatch = getNum('cfg-max-watch');
    if (maxWatch) config.max_watch_streams = maxWatch;

    const priority = getChipSelectValues('cfg-priority');
    config.priority = priority;

    const proxy = getVal('cfg-proxy');
    if (proxy) config.proxy = proxy;

    config.features = {};
    if (getChecked('cfg-claim-drops-startup')) config.features.claim_drops_startup = true;
    if (getChecked('cfg-enable-analytics')) config.features.enable_analytics = true;
    if (Object.keys(config.features).length === 0) delete config.features;

    if (getChecked('cfg-cw-enabled') || collectCategories().length > 0) {
      config.category_watcher = {};
      if (getChecked('cfg-cw-enabled')) config.category_watcher.enabled = true;
      const cwInt = getVal('cfg-cw-interval');
      if (cwInt && cwInt !== schema.defaults.category_watcher_poll_interval) config.category_watcher.poll_interval = cwInt;
      if (getChecked('cfg-cw-drops-only')) config.category_watcher.drops_only = true;
      const reminders = getTagValues('cfg-cw-reminders');
      if (reminders.length > 0) config.category_watcher.campaign_reminders = reminders;
      const cats = collectCategories();
      if (cats.length > 0) config.category_watcher.categories = cats;
    }

    if (getChecked('cfg-tw-enabled') || collectTeams().length > 0) {
      config.team_watcher = {};
      if (getChecked('cfg-tw-enabled')) config.team_watcher.enabled = true;
      const twInt = getVal('cfg-tw-interval');
      if (twInt && twInt !== schema.defaults.team_watcher_poll_interval) config.team_watcher.poll_interval = twInt;
      const teams = collectTeams();
      if (teams.length > 0) config.team_watcher.teams = teams;
    }

    const sd = collectStreamerDefaults();
    if (Object.keys(sd).length > 0) config.streamer_defaults = sd;

    const streamers = collectStreamers();
    if (streamers.length > 0) config.streamers = streamers;

    const blacklist = getTagValues('cfg-blacklist');
    if (blacklist.length > 0) config.blacklist = blacklist;

    const catBlacklist = getTagValues('cfg-cat-blacklist');
    if (catBlacklist.length > 0) config.category_blacklist = catBlacklist;

    if (getChecked('cfg-fol-enabled')) {
      config.followers = { enabled: true };
      const order = getVal('cfg-fol-order');
      if (order && order !== 'ASC') config.followers.order = order;
    }

    const notif = collectNotifications();
    if (Object.keys(notif).length > 0) config.notifications = notif;

    return config;
  }

  function collectStreamerDefaults() {
    const sd = {};
    assignTriToggle(sd, 'make_predictions', 'cfg-sd-predictions');
    assignTriToggle(sd, 'follow_raid', 'cfg-sd-follow-raid');
    assignTriToggle(sd, 'claim_drops', 'cfg-sd-claim-drops');
    assignTriToggle(sd, 'claim_moments', 'cfg-sd-claim-moments');
    assignTriToggle(sd, 'watch_streak', 'cfg-sd-watch-streak');
    assignTriToggle(sd, 'community_goals', 'cfg-sd-community-goals');
    const chat = getVal('cfg-sd-chat');
    if (chat) sd.chat = chat;

    const bet = {};
    const strategy = getVal('cfg-bet-strategy');
    if (strategy) bet.strategy = strategy;
    assignNum(bet, 'percentage', 'cfg-bet-percentage');
    assignNum(bet, 'percentage_gap', 'cfg-bet-gap');
    assignNum(bet, 'max_points', 'cfg-bet-max');
    assignNum(bet, 'minimum_points', 'cfg-bet-min');
    assignTriToggle(bet, 'stealth_mode', 'cfg-bet-stealth');
    const delay = getNumFloat('cfg-bet-delay');
    if (delay !== null) bet.delay = delay;
    const delayMode = getVal('cfg-bet-delay-mode');
    if (delayMode) bet.delay_mode = delayMode;

    const filterBy = getVal('cfg-bet-filter-by');
    const filterWhere = getVal('cfg-bet-filter-where');
    const filterValue = getNumFloat('cfg-bet-filter-value');
    if (filterBy && filterWhere) {
      bet.filter_condition = { by: filterBy, where: filterWhere, value: filterValue || 0 };
    }

    if (Object.keys(bet).length > 0) sd.bet = bet;
    return sd;
  }

  function collectCategories() {
    const items = document.querySelectorAll('#cfg-cw-categories .dynamic-item');
    return Array.from(items).map((item) => {
      const cat = {};
      cat.slug = item.querySelector('.cat-slug')?.value?.trim() || '';
      const dropsOnly = item.querySelector('.cat-drops-only');
      if (dropsOnly && dropsOnly.value === 'true') cat.drops_only = true;
      else if (dropsOnly && dropsOnly.value === 'false') cat.drops_only = false;
      const reminders = [];
      item.querySelectorAll('.cat-reminders .tag').forEach((t) => reminders.push(t.dataset.value));
      if (reminders.length > 0) cat.campaign_reminders = reminders;
      return cat;
    }).filter((c) => c.slug);
  }

  function collectTeams() {
    const items = document.querySelectorAll('#cfg-tw-teams .dynamic-item');
    return Array.from(items).map((item) => {
      return { name: item.querySelector('.team-name')?.value?.trim() || '' };
    }).filter((t) => t.name);
  }

  function collectStreamers() {
    const items = document.querySelectorAll('#cfg-streamers .dynamic-item');
    return Array.from(items).map((item) => {
      const s = { username: item.querySelector('.streamer-username')?.value?.trim() || '' };
      const details = item.querySelector('.item-details');
      if (details) {
        const settings = {};
        const predEl = details.querySelector('.s-predictions');
        if (predEl) assignTriToggleEl(settings, 'make_predictions', predEl);
        const chatEl = details.querySelector('.s-chat');
        if (chatEl && chatEl.value) settings.chat = chatEl.value;

        const betStrategy = details.querySelector('.s-bet-strategy');
        if (betStrategy && betStrategy.value) {
          if (!settings.bet) settings.bet = {};
          settings.bet.strategy = betStrategy.value;
        }
        const betMax = details.querySelector('.s-bet-max');
        if (betMax && betMax.value) {
          if (!settings.bet) settings.bet = {};
          settings.bet.max_points = parseInt(betMax.value, 10);
        }

        if (Object.keys(settings).length > 0) s.settings = settings;
      }
      return s;
    }).filter((s) => s.username);
  }

  function collectNotifications() {
    const notif = {};
    const batch = {};
    const batchEnabled = getTriToggleValue('cfg-batch-enabled');
    if (batchEnabled !== null) batch.enabled = batchEnabled;
    const batchInt = getVal('cfg-batch-interval');
    if (batchInt) batch.interval = batchInt;
    const batchMax = getNum('cfg-batch-max');
    if (batchMax) batch.max_entries = batchMax;
    const immediate = getChipSelectValues('cfg-batch-immediate');
    if (immediate.length > 0) batch.immediate_events = immediate;
    if (Object.keys(batch).length > 0) notif.batch = batch;

    NOTIFICATION_PROVIDERS.forEach((p) => {
      const section = document.querySelector('[data-provider="' + p + '"]');
      if (!section) return;
      const enabledEl = section.querySelector('.prov-enabled');
      if (!enabledEl) return;
      const pConfig = {};
      if (enabledEl.checked) pConfig.enabled = true;
      const events = [];
      section.querySelectorAll('.prov-events .chip.selected').forEach((c) => events.push(c.dataset.value));
      if (events.length > 0) pConfig.events = events;

      if (p === 'telegram') {
        const disNotif = section.querySelector('.prov-disable-notification');
        if (disNotif && disNotif.checked) pConfig.disable_notification = true;
      }
      if (p === 'webhook') {
        const method = section.querySelector('.prov-method');
        if (method && method.value) pConfig.method = method.value;
      }

      if (Object.keys(pConfig).length > 0) notif[p] = pConfig;
    });

    return notif;
  }

  // ─── Render Helpers ───

  function renderChipSelect(containerId, options, selected) {
    const container = document.getElementById(containerId);
    container.textContent = '';
    options.forEach((opt) => {
      const chip = el('span', {
        className: 'chip' + (selected.includes(opt) ? ' selected' : ''),
        dataset: { value: opt },
        onclick: function () { this.classList.toggle('selected'); markDirty(); },
      }, opt);
      container.appendChild(chip);
    });
  }

  function getChipSelectValues(containerId) {
    return Array.from(document.querySelectorAll('#' + containerId + ' .chip.selected'))
      .map((c) => c.dataset.value);
  }

  function renderTagList(containerId, values) {
    const container = document.getElementById(containerId);
    container.textContent = '';
    values.forEach((v) => addTag(container, v));
  }

  function addTag(container, value) {
    const removeBtn = el('span', {
      className: 'tag-remove',
      onclick: function () { this.parentElement.remove(); markDirty(); },
    }, '\u00d7');
    const tag = el('span', { className: 'tag', dataset: { value: value } }, [
      document.createTextNode(value + ' '),
      removeBtn,
    ]);
    container.appendChild(tag);
    markDirty();
  }

  function getTagValues(containerId) {
    return Array.from(document.querySelectorAll('#' + containerId + ' .tag'))
      .map((t) => t.dataset.value);
  }

  function renderTriToggle(containerId, value) {
    const select = document.getElementById(containerId);
    select.textContent = '';
    const strVal = value === undefined || value === null ? '' : String(value);

    [['', 'Default'], ['true', 'Yes'], ['false', 'No']].forEach(([val, label]) => {
      const opt = el('option', { value: val }, label);
      if (val === strVal) opt.selected = true;
      select.appendChild(opt);
    });

    updateTriSelectStyle(select);
    select.onchange = () => { updateTriSelectStyle(select); markDirty(); };
  }

  function updateTriSelectStyle(select) {
    select.classList.remove('val-true', 'val-false', 'val-null');
    if (select.value === 'true') select.classList.add('val-true');
    else if (select.value === 'false') select.classList.add('val-false');
    else select.classList.add('val-null');
  }

  function getTriToggleValue(containerId) {
    const elem = document.getElementById(containerId);
    if (!elem) return null;
    const v = elem.value;
    if (v === 'true') return true;
    if (v === 'false') return false;
    return null;
  }

  function assignTriToggle(obj, key, containerId) {
    const v = getTriToggleValue(containerId);
    if (v !== null) obj[key] = v;
  }

  function assignTriToggleEl(obj, key, elem) {
    const v = elem.value !== undefined ? elem.value : elem.dataset.value;
    if (v === 'true') obj[key] = true;
    else if (v === 'false') obj[key] = false;
  }

  function assignNum(obj, key, inputId) {
    const v = getNum(inputId);
    if (v !== null) obj[key] = v;
  }

  // ─── Dynamic Lists ───

  function renderCategories(categories) {
    const container = document.getElementById('cfg-cw-categories');
    container.textContent = '';
    categories.forEach((cat) => addCategoryItem(container, cat));
  }

  function addCategoryItem(container, cat) {
    cat = cat || {};
    const slugInput = el('input', {
      type: 'text', className: 'form-input input-sm cat-slug',
      placeholder: 'category-slug', value: cat.slug || '',
      oninput: markDirty,
    });
    const triEl = el('select', { className: 'tri-select cat-drops-only' });
    const label = el('span', { style: 'font-size:0.7rem;color:var(--text-muted)' }, 'drops only');
    const removeBtn = el('button', {
      type: 'button', className: 'item-remove', title: 'Remove',
      onclick: function () { this.closest('.dynamic-item').remove(); markDirty(); },
    }, '\u00d7');

    const fields = el('div', { className: 'item-fields' }, [slugInput, triEl, label]);
    const item = el('div', { className: 'dynamic-item' }, [fields, removeBtn]);
    container.appendChild(item);

    renderTriToggleInline(triEl, cat.drops_only);
  }

  function renderTriToggleInline(container, value) {
    const strVal = value === undefined || value === null ? '' : String(value);
    container.textContent = '';

    [['', 'Default'], ['true', 'Yes'], ['false', 'No']].forEach(([val, label]) => {
      const opt = el('option', { value: val }, label);
      if (val === strVal) opt.selected = true;
      container.appendChild(opt);
    });

    updateTriSelectStyle(container);
    container.onchange = () => { updateTriSelectStyle(container); markDirty(); };
  }

  function renderTeams(teams) {
    const container = document.getElementById('cfg-tw-teams');
    container.textContent = '';
    teams.forEach((t) => addTeamItem(container, t));
  }

  function addTeamItem(container, team) {
    team = team || {};
    const input = el('input', {
      type: 'text', className: 'form-input input-sm team-name',
      placeholder: 'team-name', value: team.name || '', oninput: markDirty,
    });
    const removeBtn = el('button', {
      type: 'button', className: 'item-remove', title: 'Remove',
      onclick: function () { this.closest('.dynamic-item').remove(); markDirty(); },
    }, '\u00d7');
    const fields = el('div', { className: 'item-fields' }, [input]);
    const item = el('div', { className: 'dynamic-item' }, [fields, removeBtn]);
    container.appendChild(item);
  }

  function renderStreamers(streamers) {
    const container = document.getElementById('cfg-streamers');
    container.textContent = '';
    streamers.forEach((s) => addStreamerItem(container, s));
  }

  function addStreamerItem(container, streamer) {
    streamer = streamer || {};
    const settings = streamer.settings || {};

    const expandBtn = el('button', { type: 'button', className: 'item-expand', title: 'Toggle settings' }, '\u25b6');
    const usernameInput = el('input', {
      type: 'text', className: 'form-input input-sm streamer-username',
      placeholder: 'username', value: streamer.username || '', oninput: markDirty,
    });
    const removeBtn = el('button', {
      type: 'button', className: 'item-remove', title: 'Remove',
      onclick: function () { this.closest('.dynamic-item').remove(); updateStreamerCount(); markDirty(); },
    }, '\u00d7');

    const topRow = el('div', { style: 'display:flex;align-items:center;gap:0.5rem;width:100%' }, [expandBtn, usernameInput, removeBtn]);

    // Per-streamer settings
    const predTriToggle = el('select', { className: 'tri-select s-predictions' });
    const chatSelect = el('select', { className: 'form-input input-sm s-chat' });
    [['', '— default —'], ['ALWAYS', 'ALWAYS'], ['NEVER', 'NEVER'], ['ONLINE', 'ONLINE'], ['OFFLINE', 'OFFLINE']].forEach(([val, text]) => {
      const opt = el('option', { value: val }, text);
      if (settings.chat === val) opt.selected = true;
      chatSelect.appendChild(opt);
    });

    const betStrategySelect = el('select', { className: 'form-input input-sm s-bet-strategy' });
    betStrategySelect.appendChild(el('option', { value: '' }, '— default —'));
    schema.strategies.forEach((s) => {
      const opt = el('option', { value: s }, s);
      if (settings.bet?.strategy === s) opt.selected = true;
      betStrategySelect.appendChild(opt);
    });

    const betMaxInput = el('input', {
      type: 'number', className: 'form-input input-sm s-bet-max',
      min: '0', value: settings.bet?.max_points ?? '',
    });

    const detailGrid = el('div', { className: 'form-grid', style: 'margin-top:0.5rem' }, [
      el('div', { className: 'form-field' }, [el('label', { style: 'font-size:0.7rem' }, 'Make Predictions'), predTriToggle]),
      el('div', { className: 'form-field' }, [el('label', { style: 'font-size:0.7rem' }, 'Chat'), chatSelect]),
      el('div', { className: 'form-field' }, [el('label', { style: 'font-size:0.7rem' }, 'Bet Strategy'), betStrategySelect]),
      el('div', { className: 'form-field' }, [el('label', { style: 'font-size:0.7rem' }, 'Bet Max Points'), betMaxInput]),
    ]);

    const details = el('div', { className: 'item-details' }, [detailGrid]);

    expandBtn.onclick = () => {
      expandBtn.classList.toggle('open');
      details.classList.toggle('open');
    };

    const item = el('div', { className: 'dynamic-item', style: 'flex-direction:column' }, [topRow, details]);
    container.appendChild(item);

    renderTriToggleInline(predTriToggle, settings.make_predictions);
  }

  function updateStreamerCount() {
    const count = document.querySelectorAll('#cfg-streamers .dynamic-item').length;
    document.getElementById('streamers-count').textContent = count;
  }

  // ─── Notification Providers ───

  function renderNotificationProviders(notif) {
    const container = document.getElementById('notification-providers');
    container.textContent = '';

    NOTIFICATION_PROVIDERS.forEach((p) => {
      const pConfig = notif[p] || {};

      const enabledId = 'prov-' + p + '-enabled';
      const enabledInput = el('input', {
        type: 'checkbox', className: 'toggle-input prov-enabled', id: enabledId,
      });
      if (pConfig.enabled) enabledInput.checked = true;
      const enabledLabel = el('label', { for: enabledId, className: 'toggle-slider' });

      const title = el('h5', {}, p.charAt(0).toUpperCase() + p.slice(1));
      const header = el('div', { className: 'provider-header' }, [
        title,
        el('div', { className: 'toggle-wrap' }, [enabledInput, enabledLabel]),
      ]);

      const eventsContainer = el('div', { className: 'chip-select prov-events' });
      schema.notification_events.forEach((evt) => {
        const chip = el('span', {
          className: 'chip' + ((pConfig.events || []).includes(evt) ? ' selected' : ''),
          dataset: { value: evt },
          onclick: function () { this.classList.toggle('selected'); markDirty(); },
        }, evt);
        eventsContainer.appendChild(chip);
      });

      const eventsField = el('div', { className: 'form-field' }, [
        el('label', {}, 'Events'),
        eventsContainer,
      ]);

      const section = el('div', { className: 'provider-section', dataset: { provider: p } }, [header, eventsField]);

      if (p === 'telegram') {
        const disNotifId = 'prov-' + p + '-disable-notif';
        const disInput = el('input', { type: 'checkbox', className: 'toggle-input prov-disable-notification', id: disNotifId });
        if (pConfig.disable_notification) disInput.checked = true;
        const disLabel = el('label', { for: disNotifId, className: 'toggle-slider' });
        section.appendChild(el('div', { className: 'form-field', style: 'margin-top:0.5rem' }, [
          el('label', {}, 'Disable Notification Sound'),
          el('div', { className: 'toggle-wrap' }, [disInput, disLabel]),
        ]));
      }

      if (p === 'webhook') {
        const methodSelect = el('select', { className: 'form-input input-sm prov-method', id: 'prov-' + p + '-method' });
        ['POST', 'GET'].forEach((m) => {
          const opt = el('option', { value: m }, m);
          if (pConfig.method === m || (!pConfig.method && m === 'POST')) opt.selected = true;
          methodSelect.appendChild(opt);
        });
        section.appendChild(el('div', { className: 'form-field', style: 'margin-top:0.5rem' }, [
          el('label', {}, 'Method'),
          methodSelect,
        ]));
      }

      container.appendChild(section);
    });
  }

  // ─── Validation ───

  function validateForm(config) {
    const errors = [];
    const maxWatch = config.max_watch_streams;
    if (maxWatch !== undefined && maxWatch < 1) errors.push('Max watch streams must be at least 1');

    const hasStreamers = config.streamers && config.streamers.length > 0;
    const hasFollowers = config.followers && config.followers.enabled;
    const hasCW = config.category_watcher && config.category_watcher.enabled;
    const hasTW = config.team_watcher && config.team_watcher.enabled;
    if (!hasStreamers && !hasFollowers && !hasCW && !hasTW) {
      errors.push('At least one watch source required (streamers, followers, category watcher, or team watcher)');
    }

    if (config.streamers) {
      config.streamers.forEach((s, i) => {
        if (!s.username) errors.push('Streamer #' + (i + 1) + ' has empty username');
      });
    }

    if (config.category_watcher?.enabled && (!config.category_watcher.categories || config.category_watcher.categories.length === 0)) {
      errors.push('Category watcher is enabled but no categories added');
    }

    if (config.team_watcher?.enabled && (!config.team_watcher.teams || config.team_watcher.teams.length === 0)) {
      errors.push('Team watcher is enabled but no teams added');
    }

    if (config.streamer_defaults?.make_predictions === true && !config.streamer_defaults?.bet) {
      errors.push('Predictions enabled but no bet settings configured');
    }

    return errors;
  }

  // ─── Utilities ───

  function setVal(id, val) { const elem = document.getElementById(id); if (elem) elem.value = val; }
  function getVal(id) { const elem = document.getElementById(id); return elem ? elem.value.trim() : ''; }
  function setChecked(id, val) { const elem = document.getElementById(id); if (elem) elem.checked = val; }
  function getChecked(id) { const elem = document.getElementById(id); return elem ? elem.checked : false; }
  function getNum(id) { const v = getVal(id); return v ? parseInt(v, 10) : null; }
  function getNumFloat(id) { const v = getVal(id); return v ? parseFloat(v) : null; }
  function markDirty() { isDirty = true; }

  // ─── Actions ───

  async function handleSave() {
    if (!currentAccount) return;
    const config = collectConfig();
    const errors = validateForm(config);
    if (errors.length > 0) {
      const list = el('ul');
      errors.forEach((e) => list.appendChild(el('li', {}, e)));
      showModal('Validation Errors', list, [
        { label: 'OK', class: 'btn-primary' },
      ]);
      return;
    }
    try {
      await saveConfig(currentAccount, config);
      isDirty = false;
      showToast('Config saved!', 'success');
      accounts = await fetchAccounts();
      renderSidebar();
    } catch (err) {
      showToast('Save failed: ' + err.message, 'error');
    }
  }

  function handleDelete() {
    if (!currentAccount) return;
    const body = el('div', {}, [
      el('p', {}, 'Are you sure you want to delete "' + currentAccount + '"?'),
      el('p', {}, 'This will permanently remove the config file. This cannot be undone.'),
    ]);
    showModal('Delete Account', body, [
      { label: 'Cancel', class: 'btn-ghost' },
      {
        label: 'Delete', class: 'btn-danger', action: async () => {
          try {
            await deleteConfig(currentAccount);
            showToast('Account deleted', 'info');
            currentAccount = null;
            isDirty = false;
            accounts = await fetchAccounts();
            renderSidebar();
            document.getElementById('editor-placeholder').classList.remove('hidden');
            document.getElementById('editor-content').classList.add('hidden');
          } catch (err) {
            showToast('Delete failed: ' + err.message, 'error');
          }
        },
      },
    ]);
  }

  function handleNewAccount() {
    const input = el('input', {
      type: 'text', id: 'new-account-name', className: 'form-input',
      placeholder: 'my_twitch_username',
    });
    const body = el('div', {}, [
      el('p', {}, 'Enter the Twitch username for the new account:'),
      input,
    ]);
    showModal('Create New Account', body, [
      { label: 'Cancel', class: 'btn-ghost' },
      {
        label: 'Create', class: 'btn-primary', action: async () => {
          const name = document.getElementById('new-account-name')?.value?.trim();
          if (!name) { showToast('Name is required', 'error'); return; }
          if (!/^[a-zA-Z0-9_-]+$/.test(name)) { showToast('Invalid name. Use only letters, numbers, _ and -', 'error'); return; }
          try {
            const defaultConfig = {
              streamers: [{ username: 'placeholder' }],
              features: { enable_analytics: true },
            };
            await createConfig(name, defaultConfig);
            showToast('Account created!', 'success');
            accounts = await fetchAccounts();
            renderSidebar();
            await selectAccount(name);
          } catch (err) {
            showToast('Create failed: ' + err.message, 'error');
          }
        },
      },
    ]);
    setTimeout(() => document.getElementById('new-account-name')?.focus(), 100);
  }

  // ─── Populate Select Dropdowns ───

  function populateSelects() {
    const strategySelect = document.getElementById('cfg-bet-strategy');
    strategySelect.textContent = '';
    strategySelect.appendChild(el('option', { value: '' }, '— none —'));
    schema.strategies.forEach((s) => strategySelect.appendChild(el('option', { value: s }, s)));

    const filterBySelect = document.getElementById('cfg-bet-filter-by');
    filterBySelect.textContent = '';
    filterBySelect.appendChild(el('option', { value: '' }, '— none —'));
    schema.filter_by.forEach((s) => filterBySelect.appendChild(el('option', { value: s }, s)));

    const filterWhereSelect = document.getElementById('cfg-bet-filter-where');
    filterWhereSelect.textContent = '';
    filterWhereSelect.appendChild(el('option', { value: '' }, '— none —'));
    schema.filter_where.forEach((s) => filterWhereSelect.appendChild(el('option', { value: s }, s)));
  }

  // ─── Event Listeners ───

  function setupEventListeners() {
    document.getElementById('save-btn').onclick = handleSave;
    document.getElementById('delete-btn').onclick = handleDelete;
    document.getElementById('new-account-btn').onclick = handleNewAccount;

    document.getElementById('modal-overlay').onclick = (e) => {
      if (e.target === e.currentTarget) hideModal();
    };

    document.querySelectorAll('.accordion-header').forEach((header) => {
      header.onclick = () => { header.parentElement.classList.toggle('open'); };
    });

    document.getElementById('add-category-btn').onclick = () => {
      addCategoryItem(document.getElementById('cfg-cw-categories'), {});
      markDirty();
    };

    document.getElementById('add-team-btn').onclick = () => {
      addTeamItem(document.getElementById('cfg-tw-teams'), {});
      markDirty();
    };

    document.getElementById('add-streamer-btn').onclick = () => {
      addStreamerItem(document.getElementById('cfg-streamers'), {});
      updateStreamerCount();
      markDirty();
    };

    document.querySelectorAll('[data-add-tag]').forEach((btn) => {
      const targetId = btn.dataset.addTag;
      const inputId = targetId + '-input';
      btn.onclick = () => {
        const input = document.getElementById(inputId);
        const val = input.value.trim();
        if (!val) return;
        addTag(document.getElementById(targetId), val);
        input.value = '';
      };
      document.getElementById(inputId).onkeydown = (e) => {
        if (e.key === 'Enter') { e.preventDefault(); btn.click(); }
      };
    });

    document.getElementById('config-form').oninput = markDirty;
    document.getElementById('config-form').onchange = markDirty;

    window.onbeforeunload = (e) => {
      if (isDirty) { e.preventDefault(); return ''; }
    };
  }

  // ─── Init ───

  async function init() {
    try {
      schema = await fetchSchema();
      accounts = await fetchAccounts();
      populateSelects();
      renderSidebar();
      setupEventListeners();
    } catch (err) {
      showToast('Failed to initialize: ' + err.message, 'error');
    }
  }

  init();
})();
