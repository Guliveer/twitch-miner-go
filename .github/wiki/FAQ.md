# FAQ

Frequently asked questions from users of the miner. This is the expanded version
of the quick list in the [README FAQ](https://github.com/Guliveer/twitch-miner-go#116-faq).
Answers reference the code's actual behaviour, so they reflect what the program
really does — not just what the documentation hopes it does.

If your question is not answered here, see [Troubleshooting](Troubleshooting) for
errors and their causes, or ask in
[Discussions](https://github.com/Guliveer/twitch-miner-go/discussions).

---

## Running

<details>
<summary>Can I use this without installing Go?</summary>

Yes. To *run* the miner you only need the prebuilt binary for your operating
system from [GitHub Releases](https://github.com/Guliveer/twitch-miner-go/releases)
— Windows, Linux and macOS are all built in CI. Go is only required when you
build from source (`_run.sh` on Unix, `_run.bat` on Windows) or build the
standalone config editor (`tools/edit-config.sh` / `.bat`).

</details>

<details>
<summary>What exactly does the startup banner animation do?</summary>

`playStartupAnimation` prints a short spinner, then the ASCII banner and the
version line, purely for looks. It does not affect functionality. You can
suppress it with `-no-banner` or `NO_BANNER=true` (for example on headless
servers where coloured progress is noise). The `-log-no-time` flag exists for
platforms that add their own timestamps to logs — notably Fly.io.

</details>

<details>
<summary>The program asks me to log in in the terminal. Do I have to every time?</summary>

Only the first time. That is the **device-code flow**: the program prints a code
you activate at `twitch.tv/activate`, then validates the token belongs to the
channel whose config filename you created. The token is saved to a cookie file
and reused on later starts, so the interactive step is a one-time thing.

For machines with no interactive terminal at all (Docker, Fly.io, a VPS), set
`TWITCH_AUTH_TOKEN_<USERNAME>` instead so the miner never prompts. See
[Authentication](Authentication) for all five methods and when to use each.

</details>

---

## Accounts & configuration

<details>
<summary>How do I add another account?</summary>

Create a file named `configs/<your_twitch_username>.yaml`. **The filename minus
the extension *is* the username** — there is no `username` field to fill in
inside the YAML. The [file watcher](Architecture) notices the new file and
starts a miner for it automatically; there is no restart needed. Existing YAML
files are also **hot-reloaded**: when you save a change, the watcher restarts
just that account's miner.

</details>

<details>
<summary>What is the deal with `configs/guliveer_.yaml` and `guliveer_2.yaml`?</summary>

Those are the repository owner's own account configs, checked in as working
examples. They are **skipped by default** and will not run on your machine unless
you explicitly set `RUN_OWNER_ACCOUNTS=true`. You can leave them in place or
delete them; neither affects your own accounts.

</details>

<details>
<summary>I changed a YAML and nothing happened. Why?</summary>

Hot-reload is on by default, but three conditions must hold for a change to take
effect:

1. The file must actually be saved. On change the log prints
   `config changed, restarting miner`.
2. The config must be **valid**. An invalid config is skipped with
   `invalid config, skipping` in the log and the miner keeps running on the last
   good configuration.
3. The account must be **enabled** (`enabled: true`). A disabled account's miner
   is stopped, with `account disabled, stopping miner` in the log.

If the directory contains no YAML files at all, the watcher treats it as empty
and stops all miners. If you see `failed to load configs, keeping current miners`,
a transient error (permissions, I/O) occurred and your current miners are
deliberately left running.

</details>

<details>
<summary>How do I disable one account without deleting its config?</summary>

Set `enabled: false` in that account's YAML. The watcher stops that miner and it
will not be started again, but the file stays on disk so you can change it back
to `true` later.

</details>

<details>
<summary>Can I check which accounts are running?</summary>

Yes. The analytics server is exposed on port `8080` by default (`-port`). The
dashboard shows the running accounts, and the log at startup lists
`streamers`, `pubsub_topics` and the startup duration per miner. In DB mode the
REST API (`GET /api/accounts`) lists them too.

</details>

---

## Config editor & tray

<details>
<summary>How do I open the config editor?</summary>

While the miner is running, the embedded editor is always available at
**http://localhost:8070** (bound to `127.0.0.1` only — nothing on your network
can reach it). On desktop systems, right-click the **system tray icon** and pick
**Config Editor**. There is also a standalone editor (`tools/edit-config.sh` /
`.bat`) that runs a web UI on port `3000` or a terminal TUI with `--tui`. These
are the two easiest ways to manage accounts.

</details>

<details>
<summary>Why is the tray icon not showing?</summary>

Check, in order:

- **Windows service (session 0).** A service runs in a non-interactive session
  with no desktop, so the tray cannot appear. This is normal; the miner still
  runs and serves the dashboard/editor.
- **Headless/container.** On Docker or Fly.io there is no GUI. Set `NO_TRAY=true`
  (or pass `-no-tray`) to silence the related log lines and skip the check.
- **macOS cgo.** On macOS the tray library needs cgo (Objective-C/Cocoa).
  Release binaries are built on a macOS runner with cgo, so releases are fine;
  if you build the binary yourself you must build with cgo enabled
  (`CGO_ENABLED=1 go build`).
- **Explicit disable.** `-no-tray` / `NO_TRAY=true` turns it off deliberately.

When conditions are met, "System tray icon available" is logged. On
Windows/Linux/macOS desktops the icon's left-click opens the dashboard; the
right-click menu offers **Dashboard**, **Config Editor** and **Exit** (the last
one triggers a graceful shutdown).

</details>

<details>
<summary>Can I change the editor port?</summary>

Yes — `-config-editor-port <port>` (default `8070`). There is no environment
variable for it; it is flag-only. The tray's "Config Editor" entry uses the same
port automatically.

</details>

<details>
<summary>Why is the editor bound to localhost only?</summary>

Security. The editor can read and write your account configs, so it must never
be reachable from the network. It binds to `127.0.0.1` and only devices on the
same machine can use it. For remote management use the dashboard's auth
(`DASHBOARD_USER` / `DASHBOARD_PASSWORD_SHA256`) or the DB-mode REST API behind
your own auth.

</details>

---

## Telemetry & data

<details>
<summary>What is `.instance_id` and why do I not see it?</summary>

`.instance_id` is the anonymous telemetry instance identifier — a random UUID v4
generated once on first run and reused on later starts so the same install keeps
a stable ID across restarts. It is created in `DATA_DIR` (default `.`, i.e. the
directory the process runs from). The file is created **hidden and read-only**:
on Unix the leading dot already hides it, and on Windows the *Hidden* + *Read-only*
attributes are set explicitly, which is why it may be missing from a file
browser. It is never overwritten while the process runs.

</details>

<details>
<summary>Does the program send my data anywhere?</summary>

Anonymous telemetry is on by default. The heartbeat sends only the instance ID,
version, OS, architecture and the number of running accounts — **no usernames,
channel names, IP addresses, or tokens**. The server is open source
([twitch-miner-go-telemetry](https://github.com/Guliveer/twitch-miner-go-telemetry)).
To disable it entirely, set `TELEMETRY_AGREE=false`. If you want to start with a
fresh instance ID, stop the miner and remove `.instance_id` (note the hidden/read-only
attributes) from `DATA_DIR`.

</details>

<details>
<summary>Can I delete or edit `.instance_id`?</summary>

You can, but it is unnecessary. It is write-protected so the miner (which only
reads it after creation) never overwrites it. Deleting it resets the instance ID
on the next run. If you run a fork with your own telemetry server
(`TELEMETRY_URL`), keep it stable so you can correlate installs.

</details>

---

## Docker & Fly.io

<details>
<summary>I lose my configs/accounts when the container restarts. Why?</summary>

Your configs (`/configs`) and data (`/data`, where cookies, state and the
`DATA_DIR`-resident instance ID live) must be on **persistent volumes**. When
you run with `docker run`, mount them explicitly:

```bash
-v miner_data:/data \
-v $(pwd)/configs:/configs:ro
```

Or use the provided `docker-compose.yml`, which already wires the volumes and
`.env`. On Fly.io a volume (`fly volumes create miner_data`) backs the mounted
`/data`. Without persistence, everything written to `/data` (including the
instance ID) is wiped on each deploy.

</details>

<details>
<summary>Why is there a read-only configs mount in the Docker examples?</summary>

`/configs:ro` mounts your account YAMLs read-only so the container cannot
modify them. The miner only reads configs from there; any writes (cookies, saved
state) go to the writable `/data` volume. This is a deliberate separation of
immutable configs from mutable data.

</details>

<details>
<summary>Do I need `-no-tray` in Docker?</summary>

Not strictly — the tray check fails gracefully without a desktop. But setting
`NO_TRAY=true` avoids a few extraneous log lines and is the intended use for
containers. Likewise `LOG_NO_TIME=true` (or `-log-no-time`) is handy on Fly.io,
which duplicates timestamps itself.

</details>

---

## Behaviour & resources

<details>
<summary>Why is it so much lighter than the Python miner?</summary>

Go is a compiled, memory-efficient language. The miner runs on a handful of OS
threads (~4–5) using goroutines, versus 60+ threads in the Python version, and
the Docker image is a statically linked binary of a few MB rather than a Python
environment. See the comparison in the [README](https://github.com/Guliveer/twitch-miner-go#13-resource-comparison).

</details>

<details>
<summary>Must the program run 24/7 to be useful?</summary>

Yes — it only earns channel points / claims drops while it is running. For
always-on usage you have first-class options: a systemd/OpenRC unit
([Linux service](https://github.com/Guliveer/twitch-miner-go#110-linux-service-systemd--openrc)),
a Windows service via NSSM
([Windows service](https://github.com/Guliveer/twitch-miner-go#111-windows-service)),
or Docker/Fly.io. The embedded config editor and tray are conveniences for
desktop use; they do not require a GUI on servers.

</details>

<details>
<summary>What happens when a miner crashes?</summary>

The manager restarts it automatically with exponential backoff (starting at 10 s
and doubling up to 5 min). A `MINER_CRASHED` lifecycle notification is sent if
you have notifications configured. This repeats until the process runs again or
the account is removed/disabled. This recovery is why a single binary can stay
up long-term even after transient Twitch errors.

</details>