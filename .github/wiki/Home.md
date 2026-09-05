# Twitch Channel Points Miner — Go Edition

A high-performance Go rewrite of [Twitch-Channel-Points-Miner-v2](https://github.com/rdavydov/Twitch-Channel-Points-Miner-v2). Whether you have never written a line of code or you are evaluating the codebase for a production rollout, this wiki walks you through the whole project — starting from the plainest explanations and working up to the internal architecture.

## What this program does (in plain English)

Twitch gives viewers channel points over time while they watch a channel, and runs Drops and prediction events during many streams. This program automates the boring part: it runs in the background, watches chosen channels for you, claims the channel points you earn, grabs Twitch Drops when they become claimable, joins raids, and places prediction bets according to rules you set. You set it up once and it runs quietly in a corner of your computer — or on a server (Docker, Fly.io, a VPS) so you do not even need a machine of your own running 24/7.

No coding is required to use it: the included visual config editor runs in your browser at `http://localhost:8070` while the miner is running, and on desktop systems a system-tray icon offers quick links to the dashboard, the config editor, and a way to stop the miner.

| | Python | Go |
|-|--------|----|
| Memory | ~80–120 MB | ~5–15 MB |
| Docker image | ~800 MB | ~10–15 MB |
| Startup | ~5–10 s | < 100 ms |
| Threads | 60+ | ~10–20 goroutines |

The rest of this page and the linked pages progress from the simplest concepts to the most technical, so you can read from the top to the bottom for a complete tour, or jump straight to the section that matches your comfort level.

## Where do I start?

- **Never configured a bot before?** → [Getting Started](Getting-Started). It takes you from a fresh clone to a running miner in under 5 minutes and explains the concepts as it goes.
- **Just want to adjust an account?** → While the miner is running, open `http://localhost:8070` (the embedded config editor) or, on Windows/Linux/macOS desktops, right-click the tray icon and pick **Config Editor**. Changes hot-reload — no restart needed.
- **Ready to dig into options?** → [Configuration Reference](Configuration-Reference) documents every YAML option.
- **Evaluating architecture or contributing code?** → [Architecture](Architecture) maps the packages and data flow.

## Wiki pages

| Page | What it covers |
|------|----------------|
| [Getting Started](Getting-Started) | Concepts + from zero to running in under 5 minutes; easy managing via tray and embedded editor |
| [Configuration Reference](Configuration-Reference) | Every YAML option with type, default and description |
| [Authentication](Authentication) | All 5 auth methods — when and how to use each |
| [Prediction Strategies](Prediction-Strategies) | All 8 strategies explained with examples and tips |
| [Notifications](Notifications) | Setup guide for all 6 providers + batching deep-dive |
| [Troubleshooting](Troubleshooting) | Common errors, their causes and fixes — including editor/tray problems |
| [FAQ](FAQ) | Frequently asked questions and their answers — the expanded version of the README FAQ |
| [Advanced Guide](Advanced-Guide) | How the features behave under the hood — strategies, limits, background loops, recovery & performance |
| [Architecture](Architecture) | Package map and data flow for contributors and evaluators |

## Key resources

- [README](https://github.com/Guliveer/twitch-miner-go#readme) — installation, Docker, Fly.io, service setup
- [Telemetry dashboard](https://github.com/Guliveer/twitch-miner-go-telemetry) — anonymous usage data server (open source)
- [configs/example.yaml.example](https://github.com/Guliveer/twitch-miner-go/blob/main/configs/example.yaml.example) — fully annotated config template
- [CONTRIBUTING.md](https://github.com/Guliveer/twitch-miner-go/blob/main/CONTRIBUTING.md) — commit convention, git hooks, automated versioning
- [Releases](https://github.com/Guliveer/twitch-miner-go/releases) — changelog and downloads
- [Discussions](https://github.com/Guliveer/twitch-miner-go/discussions) — questions and ideas
