(function () {
  "use strict";

  const REFRESH_INTERVAL = 30000;
  const FILTER_REFRESH_INTERVAL = 60000;

  // Event emoji mapping
  const EVENT_EMOJIS = {
    GAIN_FOR_WATCH: "📺",
    GAIN_FOR_WATCH_STREAK: "📺",
    GAIN_FOR_CLAIM: "🎁",
    GAIN_FOR_RAID: "🎁",
    BONUS_CLAIM: "💰",
    BET_START: "🎰",
    BET_FILTERS: "🎰",
    BET_GENERAL: "🎰",
    BET_FAILED: "🎰",
    BET_WIN: "🏆",
    BET_LOSE: "💸",
    BET_REFUND: "↩️",
    DROP_CLAIM: "🎯",
    DROP_CLAIM_AVAILABLE: "📦",
    DROP_STATUS: "📦",
    STREAMER_ONLINE: "🟢",
    STREAMER_OFFLINE: "⚫",
    JOIN_RAID: "⚔️",
    CHAT_MENTION: "💬",
    MOMENT_CLAIM: "🎉",
  };

  // Category display config
  const CATEGORY_CONFIG = {
    drops: { label: "Drops", emoji: "🎯", color: "#e040fb" },
    points: { label: "Points", emoji: "💰", color: "#ffca28" },
    bets: { label: "Bets", emoji: "🎰", color: "#ff6f00" },
    raids: { label: "Raids", emoji: "⚔️", color: "#f44336" },
    streams: { label: "Streams", emoji: "🟢", color: "#00e676" },
    other: { label: "Other", emoji: "🎉", color: "#29b6f6" },
  };

  // Category to events mapping (must match backend)
  const CATEGORY_EVENTS = {
    drops: ["DROP_CLAIM", "DROP_CLAIM_AVAILABLE", "DROP_STATUS"],
    points: ["GAIN_FOR_WATCH", "GAIN_FOR_WATCH_STREAK", "GAIN_FOR_CLAIM", "GAIN_FOR_RAID", "BONUS_CLAIM"],
    bets: ["BET_START", "BET_WIN", "BET_LOSE", "BET_REFUND", "BET_FILTERS", "BET_GENERAL", "BET_FAILED"],
    raids: ["JOIN_RAID"],
    streams: ["STREAMER_ONLINE", "STREAMER_OFFLINE"],
    other: ["MOMENT_CLAIM", "CHAT_MENTION", "GIFTED_SUB"],
  };

  // ── Utility functions ────────────────────────────────────────────────
  function debounce(fn, ms) {
    let timer;
    return function (...args) {
      clearTimeout(timer);
      timer = setTimeout(() => fn.apply(this, args), ms);
    };
  }

  async function fetchJSON(url) {
    const res = await fetch(url);
    if (!res.ok) throw new Error(res.statusText);
    return res.json();
  }

  function formatPoints(n) {
    return Number(n).toLocaleString("en-US");
  }

  function escapeHTML(str) {
    const d = document.createElement("div");
    d.textContent = str;
    return d.innerHTML;
  }

  // ── Filter helpers ───────────────────────────────────────────────────
  function buildFilterParams() {
    const params = new URLSearchParams();
    const account = document.getElementById("filter-account").value;
    const channel = document.getElementById("filter-channel").value.trim();
    const category = document.getElementById("filter-category").value;
    const event = document.getElementById("filter-event").value;
    if (account) params.set("account", account);
    if (channel) params.set("channel", channel);
    if (category) params.set("category", category);
    if (event) params.set("event", event);
    return params.toString();
  }

  function populateSelect(id, items, defaultLabel, displayMap) {
    const sel = document.getElementById(id);
    const current = sel.value;
    sel.innerHTML = '<option value="">' + defaultLabel + "</option>";
    items.forEach(function (item) {
      const label = displayMap ? displayMap[item] || item : item;
      sel.innerHTML += '<option value="' + escapeHTML(item) + '">' + escapeHTML(label) + "</option>";
    });
    sel.value = current;
  }

  async function loadFilters() {
    try {
      const data = await fetchJSON("/api/event-filters");
      populateSelect("filter-account", data.accounts || [], "All Accounts");
      populateSelect("filter-event", data.events || [], "All Events");
    } catch (e) {
      console.error("Failed to load filters:", e);
    }
  }

  function clearFilters() {
    document.getElementById("filter-account").value = "";
    document.getElementById("filter-channel").value = "";
    document.getElementById("filter-category").value = "";
    document.getElementById("filter-event").value = "";
    refresh();
  }

  // ── Sort state ───────────────────────────────────────────────────────
  let currentSort = { key: "amount", dir: "desc" };

  function sortEntries(entries, key, dir) {
    return entries.slice().sort(function (a, b) {
      var va = a[key],
        vb = b[key];
      if (typeof va === "string") {
        va = va.toLowerCase();
        vb = vb.toLowerCase();
      }
      if (va < vb) return dir === "asc" ? -1 : 1;
      if (va > vb) return dir === "asc" ? 1 : -1;
      return 0;
    });
  }

  function handleSort(key) {
    if (currentSort.key === key) {
      currentSort.dir = currentSort.dir === "asc" ? "desc" : "asc";
    } else {
      currentSort.key = key;
      currentSort.dir = key === "count" || key === "amount" ? "desc" : "asc";
    }
    updateSortIndicators();
    renderTable(lastEntries);
  }

  function updateSortIndicators() {
    document.querySelectorAll("#events-table th[data-sort]").forEach(function (th) {
      th.classList.remove("sort-asc", "sort-desc");
      if (th.dataset.sort === currentSort.key) {
        th.classList.add(currentSort.dir === "asc" ? "sort-asc" : "sort-desc");
      }
    });
  }

  // ── Render category summary cards ────────────────────────────────────
  function renderCategorySummary(entries) {
    const container = document.getElementById("category-summary");

    // Aggregate by category
    const catTotals = {};
    for (const cat of Object.keys(CATEGORY_CONFIG)) {
      catTotals[cat] = { count: 0, amount: 0 };
    }
    for (const entry of entries) {
      for (const [cat, events] of Object.entries(CATEGORY_EVENTS)) {
        if (events.includes(entry.event)) {
          catTotals[cat].count += entry.count;
          catTotals[cat].amount += entry.amount;
          break;
        }
      }
    }

    container.innerHTML = Object.entries(CATEGORY_CONFIG)
      .map(function ([key, cfg]) {
        const totals = catTotals[key];
        return '<div class="category-card" style="border-left-color: ' + cfg.color + '">' + '<div class="category-header">' + '<span class="category-emoji">' + cfg.emoji + "</span>" + '<span class="category-label">' + cfg.label + "</span>" + "</div>" + '<div class="category-stats">' + '<span class="category-count">' + formatPoints(totals.count) + " events</span>" + '<span class="category-points">' + formatPoints(totals.amount) + " pts</span>" + "</div>" + "</div>";
      })
      .join("");
  }

  // ── Render events table ──────────────────────────────────────────────
  let lastEntries = [];

  function renderTable(entries) {
    const sorted = sortEntries(entries, currentSort.key, currentSort.dir);
    const tbody = document.getElementById("events-body");

    if (sorted.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;color:#53535f;padding:2rem;">No events found</td></tr>';
      return;
    }

    tbody.innerHTML = sorted
      .map(function (e) {
        const emoji = EVENT_EMOJIS[e.event] || "❓";
        return "<tr>" + '<td><span class="event-emoji">' + emoji + "</span> " + escapeHTML(e.event) + "</td>" + "<td>" + escapeHTML(e.streamer) + "</td>" + "<td>" + escapeHTML(e.account) + "</td>" + "<td>" + formatPoints(e.count) + "</td>" + '<td class="points">' + formatPoints(e.amount) + "</td>" + "</tr>";
      })
      .join("");
  }

  // ── Main refresh function ────────────────────────────────────────────
  async function refresh() {
    try {
      const qs = buildFilterParams();
      const url = "/api/events" + (qs ? "?" + qs : "");
      const entries = await fetchJSON(url);
      lastEntries = entries;
      renderCategorySummary(entries);
      renderTable(entries);
    } catch (e) {
      console.error("Failed to refresh:", e);
    }
  }

  // ── Bootstrap ────────────────────────────────────────────────────────
  document.addEventListener("DOMContentLoaded", function () {
    // Filter event listeners
    document.getElementById("filter-account").addEventListener("change", refresh);
    document.getElementById("filter-category").addEventListener("change", refresh);
    document.getElementById("filter-event").addEventListener("change", refresh);
    document.getElementById("filter-channel").addEventListener("input", debounce(refresh, 300));
    document.getElementById("clear-filters").addEventListener("click", clearFilters);

    // Sort event listeners
    document.querySelectorAll("#events-table th[data-sort]").forEach(function (th) {
      th.addEventListener("click", function () {
        handleSort(th.dataset.sort);
      });
    });

    // Initial load
    loadFilters();
    refresh();
    updateSortIndicators();

    // Auto-refresh
    setInterval(refresh, REFRESH_INTERVAL);
    setInterval(loadFilters, FILTER_REFRESH_INTERVAL);
  });
})();
