# work-stream

`ws` is a shared log for people and agents. Entries have a type, a
short subject, an optional body, and optional key-value metadata.
`ws-server` keeps them in SQLite.

## Build

```
scripts/build.sh
```

This produces `bin/ws` and `bin/ws-server`.

## Setup

The server needs an absolute data directory. Pass it with `--data`.
`WORK_STREAM_DATA` can also supply the directory.

The server listens on port 7139 by default. Change it with
`WORK_STREAM_PORT` or `--port`. Without a secret, the server listens
only on localhost. With a secret, it listens on all interfaces.

Clients connect to localhost on port 7139 by default. Set
`WORK_STREAM_ADDRESS` and `WORK_STREAM_PORT`, or pass the global
`--address` and `--port` flags before the command, to change this.

`WORK_STREAM_TIMEOUT` sets the client or server deadline. The default
is 5 seconds. Override it with the server's `--timeout` flag or the
client's global `--timeout` flag before the command. `ws --version`
and `ws secret` work without a server.

### Secret Authentication

Authentication is optional. Generate one shared secret, then set it
in the server and every client environment:

```
bin/ws secret
export WORK_STREAM_SECRET='<printed value>'
```

`ws secret` does not replace a value already set in the environment.

Connections use plain HTTP with no encryption. work-stream is intended
for trusted private networks, not public internet hosting.

## Run

```
bin/ws-server --data /path/to/work-stream
```

## Use

Run `ws help` for the full command reference.

Types can have 64 characters, subjects 128, bodies 2048, metadata
keys 64, and metadata values 256. Each entry can have 16 metadata
pairs.

```
# Add entries.
ws add todo "Count the new ducklings" --project duck-pond --jira QUACK-1

# Check recent entries and inspect one in full.
ws recent
ws entry e1

# Search entries.
ws search ducklings --type todo --no-subject '*Solved*'
ws search --jira 'QUACK-*'
ws search --origin-host bud110 --origin-claude-session '050f8e2e*'

# Edit entries and metadata.
ws edit e1 "Count the new ducklings (Solved 22/07/2026)"
ws add-meta e1 pr https://github.com/example/pond/pull/123

# Check the server.
ws status
```

### Search entries

Search flags take full-string, ASCII-case-insensitive SQLite GLOB
patterns:

```
--subject  --body  --content  --type  --key  --meta
--origin-user  --origin-host  --origin-dir  --origin-claude-session
--project  --jira  --confluence
```

Prefix any flag with `--no-` to exclude it. `--content` matches the
subject or body. `--meta` takes `KEY=VALUE`; both patterns must match
the same metadata pair. Repeated filters AND together.

A pattern without wildcards is exact. `*` matches any text, `?` one
character, and brackets form character classes. Quote patterns so the
shell does not expand them. In a pattern, use `[*]`, `[?]`, and `[[]`
for literal `*`, `?`, and `[`.
