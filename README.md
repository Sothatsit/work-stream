# work-stream

Work-stream is a shared chronological log for agents. Agents log
decisions, progress, and artifacts as they work, so every session
can see what the others have been up to. Entries have a type, a
short subject, an optional body, and optional key-value metadata.
`ws-server` maintains the log, and `ws` is the CLI agents use to
access it.

## Build

```
scripts/build.sh
```

This produces `bin/ws` and `bin/ws-server`.

## Setup

Set these in your bashrc, on the server host and every client host:

```
export WORK_STREAM_DATA=/absolute/path/for/server/data
export WORK_STREAM_ADDRESS=localhost   # the server's host
export WORK_STREAM_PORT=7139
```

To let clients on other hosts connect, generate one shared secret
and add it to every environment:

```
bin/ws secret
export WORK_STREAM_SECRET='<printed value>'
```

Connections use plain HTTP with no encryption. Keep work-stream on
a trusted private network, never the public internet.

## Run the Server

```
bin/ws-server
```

The server stores entries in SQLite under `WORK_STREAM_DATA` and
listens on `WORK_STREAM_PORT`. Without a secret it listens on
loopback only. With one, it listens on all interfaces.

## Use the CLI

Run `ws help` for the full command reference.

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

### Search

Search flags take full-string, ASCII-case-insensitive SQLite GLOB
patterns:

```
--subject  --body  --content  --type  --key  --meta
--origin-user  --origin-host  --origin-dir  --origin-claude-session
--project  --jira  --confluence
```

Prefix any flag with `--no-` to exclude it (e.g., `--no-subject`).
`--content` matches the subject or body. `--meta` takes `KEY=VALUE`;
both patterns must match the same metadata pair. Repeated filters
AND together.

A pattern without wildcards is exact. `*` matches any text, `?` one
character, and brackets form character classes. Quote patterns so the
shell does not expand them. In a pattern, use `[*]`, `[?]`, and `[[]`
for literal `*`, `?`, and `[`.

## Agent Skill

`skills/ws/SKILL.md` is an example agent skill for the `ws` CLI.
Copy it into your skills directory and edit its conventions to fit
how you work.
