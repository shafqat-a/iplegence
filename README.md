# iplegence

Private project: merge the best free/open IP intelligence sources into one
MaxMind-compatible MMDB (`Superior-IP.mmdb`), ship it daily, then add a lookup API.

This repo was bootstrapped on 2026-08-14. **Implementation has not started.**
The spec and the Phase 0+1 plan are in the tree. Continue from those.

## Start here on a new machine

```bash
git clone git@github.com:shafqat-a/iplegence.git
cd iplegence
```

Read, in this order:

1. [`AGENTS.md`](AGENTS.md) — handoff for the next engineer/agent
2. [`doc/spec/superior-open-ip-intelligence.md`](doc/spec/superior-open-ip-intelligence.md) — transcribed spec
3. [`docs/superpowers/plans/2026-08-14-phase0-1-mvp-pipeline.md`](docs/superpowers/plans/2026-08-14-phase0-1-mvp-pipeline.md) — executable implementation plan
4. [`doc/session/chat.md`](doc/session/chat.md) — this session’s conversation

Original image spec: [`doc/spec/grok_report.pdf`](doc/spec/grok_report.pdf) (8 pages, image-only).

## Status

| Item | State |
|------|--------|
| GitHub repo | Private: https://github.com/shafqat-a/iplegence |
| Code | None yet (plan only) |
| Next work | Execute Phase 0 + Phase 1 plan |
| Secrets needed for a live build | `IPINFO_TOKEN` (required), `MAXMIND_LICENSE_KEY` (optional) |

## License

Code will be Apache-2.0. Released data stays under each source license
(see the plan and future `ATTRIBUTION.md`). Do not claim commercial-grade accuracy.
