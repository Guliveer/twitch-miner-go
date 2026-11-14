# Configuration Reference

Each account gets one YAML file in the `configs/` directory. **The filename (without `.yaml`) is the Twitch username** — no `username` field exists in the schema.

Example: `configs/my_user.yaml` → account for Twitch user `my_user`.

---

## Top-level fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Set to `false` to disable without deleting the file |
| `max_watch_streams` | int | `2` | Maximum concurrent streams to simulate watching |
| `streak_watch_streams` | int | `2` | While any channel still has a watch streak pending, narrow the watch set to this many streams so the streak lands. Twitch only credits about two concurrent streams, so a wider set lets Twitch pick and no streak sticks. `0` disables the narrowing. |
| `watch_streak_minutes` | float | `10` | How long a channel may hold a streak slot before giving way to the next pending channel, whether or not the streak arrived |
| `preferred_streamers` | list | — | Channel logins that the `PREFERRED` priority picks first, in this order |
| `proxy` | string | — | HTTP/SOCKS5 proxy for all Twitch API requests. Format: `socks5://host:port` |

---

## `features`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `claim_drops_startup` | bool | `false` | Immediately sync and claim any pending drops on startup |
| `enable_analytics` | bool | `true` | Enable the built-in analytics HTTP server |

---

## `priority`

Ordered list of rules for selecting which streams to watch when more are live than `max_watch_streams` allows. First matching rule wins.

| Value | Description |
|-------|-------------|
| `STREAK` | Prefer streams where a watch-streak bonus is available |
| `PREFERRED` | Prefer streams listed in `preferred_streamers`, in that order |
| `DROPS` | Prefer streams with an active drop campaign |
| `ORDER` | Use the order of the `streamers` list |
| `SUBSCRIBED` | Prefer streams where you are subscribed |
| `POINTS_ASCENDING` | Prefer streams with fewer channel points (gain more) |
| `POINTS_DESCENDING` | Prefer streams with more channel points |

**Example:**
```yaml
priority:
  - STREAK
  - PREFERRED
  - DROPS
  - ORDER
```

---

## `streamer_defaults`

Applied to every streamer unless overridden per-streamer under `streamers[*].settings`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `make_predictions` | bool | `true` | Place bets in channel point predictions |
| `follow_raid` | bool | `true` | Auto-join raids |
| `claim_drops` | bool | `true` | Claim drop rewards when earned |
| `claim_moments` | bool | `true` | Claim community moments |
| `watch_streak` | bool | `true` | Maintain watch streaks |
| `community_goals` | bool | `false` | Contribute to community goals |
| `drops_only` | bool | `false` | Stop watching this streamer once all active drop campaigns are completed |
| `chat` | enum | `ONLINE` | When to join IRC chat: `ALWAYS` \| `NEVER` \| `ONLINE` \| `OFFLINE` |

### `bet` (nested under `streamer_defaults` and per-streamer `settings`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `strategy` | enum | `SMART` | Betting strategy — see [Prediction Strategies](Prediction-Strategies) |
| `percentage` | int | `5` | Percentage of current channel points to bet (used by PERCENTAGE strategy and as base for others) |
| `percentage_gap` | int | `20` | Minimum percentage gap between outcomes required to place a bet (SMART strategy) |
| `max_points` | int | `50000` | Maximum points to bet in a single prediction |
| `minimum_points` | int | `0` | Minimum channel points balance required before betting |
| `stealth_mode` | bool | `false` | Wait until others have bet before placing (reduces influence on odds) |
| `delay` | int | `6` | Seconds to wait before placing the bet |
| `delay_mode` | enum | `FROM_END` | When the delay is measured from: `FROM_START` \| `FROM_END` \| `PERCENTAGE` |

#### `bet.filter_condition` (optional)

Skip betting if a condition is not met.

| Field | Type | Description |
|-------|------|-------------|
| `by` | enum | Metric to filter on: `total_users` \| `total_points` |
| `where` | enum | Comparison: `GT` \| `GTE` \| `LT` \| `LTE` \| `EQ` |
| `value` | int | Threshold value |

**Example — only bet when 100+ users have voted:**
```yaml
bet:
  strategy: "SMART"
  filter_condition:
    by: "total_users"
    where: "GTE"
    value: 100
```

---

## `streamers`

List of streamers to watch. Per-streamer settings override `streamer_defaults`.

```yaml
streamers:
  - username: "streamer1"                   # uses all defaults
  - username: "streamer2"
    settings:                                # overrides defaults for this streamer
      make_predictions: false
      chat: "NEVER"
  - username: "streamer3"
    settings:
      bet:
        strategy: "HIGH_ODDS"
        max_points: 10000
```

All fields available under `streamer_defaults` are also valid under `settings`.

---

## `blacklist`

Streamers to exclude even if they appear in followers mode or a watcher.

```yaml
blacklist:
  - "streamer_to_skip"
```

---

## `category_blacklist`

Game category slugs to exclude from category watcher discovery.

```yaml
category_blacklist:
  - "just-chatting"
  - "pools-hot-tubs-and-beaches"
```

---

## `followers`

Watch all channels you follow, in addition to the explicit `streamers` list.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable followers mode |
| `order` | enum | `ASC` | Order to process followed channels: `ASC` (oldest follow first) \| `DESC` |

---

## Watcher options

### `category_watcher`

Auto-discover live streams by game category.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable category watcher |
| `poll_interval` | duration | `120s` | How often to re-query Twitch for live streams in the category |
| `drops_only` | bool | `false` | Global default: only watch streams with active drop campaigns |
| `categories` | list | — | List of category objects |

Each category entry:

| Field | Type | Description |
|-------|------|-------------|
| `slug` | string | Category URL slug (e.g. `league-of-legends`, `just-chatting`) |
| `drops_only` | bool | Per-category override for `drops_only` |

```yaml
category_watcher:
  enabled: true
  poll_interval: 120s
  drops_only: false
  categories:
    - slug: "just-chatting"
    - slug: "league-of-legends"
      drops_only: true
```

### `URGENT_REWARDS`

Opt in by adding `URGENT_REWARDS` to the `priority` list. Put it first to let a
time-sensitive reward take precedence over normal badge, streak, and Drops
selection. Keep the existing priorities for the remaining slots and for when no
reward is urgent:

```yaml
priority: [URGENT_REWARDS, BADGES, STREAK, DROPS, ORDER]
```

Only one slot may be reserved by this priority. For each eligible campaign the
miner considers the nearest unfinished watch-time reward and subtracts the
remaining credited watch time from its deadline. It prefers the smallest
remaining slack, between five minutes (a safety margin) and two hours. Within
five minutes of the tightest slack, it prefers a channel already being watched,
then a badge. Expired, unattainable, paid, and already claimable rewards do not
get the urgent slot. Regular priorities may still select those channels.

This only ranks already tracked, live, eligible channels. Discovery, the badge
watcher limit, existing watch-slot limits, and claim settings stay unchanged.
Progress comes from Twitch inventory, not from locally sent watch events. Badge
rewards without matched inventory progress can use the full watch duration from
the badge catalog as a conservative estimate. Ordinary Drops without confirmed
inventory progress keep their normal priority until their eligibility and
remaining time can be verified. When the Drops dashboard is empty, in-progress
inventory campaigns can supply metadata for channels whose Twitch campaign IDs
match. This also makes those campaigns visible to the existing campaign-based
priorities.

### `badge_watcher`

Automatically discovers active watch-time Drops whose rewards match global
Twitch chat badges, then keeps an eligible live channel in the watch set. The
feature is opt-in and uses community-maintained Drops and badge catalogs; both
catalogs are cached for one hour. Inventory and earned-badge checks still run
on every poll.

`BADGES` must be present in `priority` when this watcher is enabled, otherwise
the discovered channel is not guaranteed a watch slot.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable badge campaign discovery |
| `poll_interval` | duration | `5m` | How often to check earned badges, inventory, and eligible live streams |
| `streamer_limit` | int | `1` | Concurrent badge-discovered channels; valid range is 1–2 |
| `drops_catalog_url` | string | v3 community catalog | Optional Drops catalog override |
| `badges_catalog_url` | string | v3 community catalog | Optional global badge catalog override |

```yaml
priority:
  - BADGES
  - STREAK
  - DROPS
  - ORDER

badge_watcher:
  enabled: true
  poll_interval: 5m
  streamer_limit: 1
```

### `team_watcher`

Auto-discover live streams from a Twitch team. Picks the stream with the most viewers per team.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable team watcher |
| `poll_interval` | duration | `180s` | How often to re-query |
| `teams` | list | — | List of team objects |

Each team entry:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Team slug from the URL: `https://www.twitch.tv/team/{name}` |

---

## `notifications`

See the full [Notifications](Notifications) page for provider setup guides.

### `batch` (global batching defaults)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable notification batching |
| `interval` | duration | `15m` | How often to flush buffered notifications |
| `max_entries` | int | `15` | Max lines per message before splitting (0 = unlimited) |
| `immediate_events` | list | `[]` | Events that bypass batching and are sent instantly |

Lifecycle events (`MINER_STARTED`, `MINER_STOPPED`, `MINER_CRASHED`) always bypass batching regardless of this setting.

### Provider keys

`telegram`, `discord`, `webhook`, `matrix`, `pushover`, `gotify` — see [Notifications](Notifications) for their fields.

---

## Full example

See [`configs/example.yaml.example`](https://github.com/Guliveer/twitch-miner-go/blob/main/configs/example.yaml.example) in the repository for the fully annotated reference config.
