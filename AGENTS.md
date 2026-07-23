# work-stream

Work-stream is a shared chronological log for agents. `ws-server`
maintains the log, and `ws` is the CLI agents use to access it.
The README covers setup, and `ws help` covers the commands. For
anything deeper, read the source. The documentation is small on
purpose.

## Development Practices

- Keep documentation minimal. Agents can read the source code to
  see how things work, and do not need extensive documentation.
  Extending the README needs a good justification, and preferably
  human approval. The bar for code comments is also high.
- Keep the project simple. Build flexible primitives that people
  can adapt to their own usage, rather than features that fit one
  use-case.
- Fail fast. Prefer crashing over dubious recovery.
- Build with `scripts/build.sh`. Run the unit tests with
  `go test ./...` and the integration tests with `test/e2e.sh`.
