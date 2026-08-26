---
name: ws
description: "Record and search the work stream: a shared chronological log of typed entries such as notes, todos, decisions, ideas, and artifacts. Use it to get broader context on recent work, such as when planning, and to log decisions, useful progress, and artifacts another session may need. Do not use the ws CLI without loading this skill first."
---

# Work Stream

`ws` is a CLI for a shared, chronological log. Durable work lives in
your issue tracker, documentation, and pull requests. When you create
one of those artifacts, log an entry in the work-stream that links
to it.

Run `ws help` for the full reference.

## Commands

### add

```
ws add <type> <subject> [--body <b>] [--meta <key>=<value>]... [--project <p>] [--jira <k>] [--confluence <u>]
```

Adds an entry. Origin (user, host, directory, Claude session) is
captured automatically. The subject is a headline of up to 128
characters; longer detail goes in `--body`, up to 2048.

### recent

```
ws recent [-n <limit>] [--offset <n>] [--order-by-creation | --order-by-modified]
```

The most recently changed entries, default 20. Editing an entry lifts
it back to the top, so a fact you correct resurfaces. Pass
`--order-by-creation` for the order entries were written in. A listing
shows the subject alone, open an entry with `ws entry` to see its body
and metadata.

### search

```
ws search [<text>] [filters] [--id-only]
```

A bare `<text>` is the same as `--content`, so it matches any part of
the subject or body: `ws search duck` and `ws search '*duck*'` find the
same entries. Filters are described below. `--id-only` prints an entry
id per line.

### entry

```
ws entry <id>
```

Shows one entry in full: subject, origin, metadata, and body.

### edit

```
ws edit <id> [<subject>] [--subject <s>] [--body <b>]
```

Changes subject and/or body. `--body ""` clears the body. Metadata is
changed with the meta commands.

### delete

```
ws delete <id>
```

### add-meta / edit-meta / remove-meta

```
ws add-meta <id> <key> <value>
ws edit-meta <id> <key> <value>
ws remove-meta <id> <key>
```

`add-meta` refuses to overwrite an existing key.

### status

```
ws status
```

Server address, connectivity, and secret state.

### Search filters

Every filter takes an ASCII-case-insensitive GLOB pattern.

`--subject`, `--body`, and `--content` (subject or body) search prose,
so they match any part of the value: `--subject duck` finds "Count the
Ducklings".

`--type`, `--key`, `--meta <key>=<value>`, the origin filters
`--origin-user`, `--origin-host`, `--origin-dir`,
`--origin-claude-session`, and the metadata shorthands `--project`,
`--jira`, `--confluence` each name one of a known set, so they match
the whole value: `--type not` does not match `note`, while
`--type 'not*'` does.

Prefix any flag with `--no-` to exclude (e.g., `--no-subject`).
Repeated filters AND together. Patterns use `*`, `?`, and bracket
classes; quote them so the shell does not expand them. Nothing is
escaped for you, so to match a literal `*`, `?`, or `[`, put it in a
bracket class: `--content '*[*]args*'` to partial match `*args`.
Default limit is 20, maximum 500.

To find what your current Claude session logged:
`ws search --origin-claude-session "$CLAUDE_CODE_SESSION_ID"`.

## Conventions

Adapt these to how you work.

Types: `note`, `todo`, `decision`, `idea`, `artifact`, `status`.

`--project`: the repository or project the work belongs to, or a
representative name when there is no repository.

`--jira`: the issue the entry relates to, e.g. `QUACK-1`.

`--confluence`: a related documentation page.

For any key without a shorthand, use `--meta <key>=<value>`. Number
repeats of a key as `jira-1`, `jira-2`.

Append `(Solved <date>)` to a finished todo's subject.
