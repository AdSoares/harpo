# ADR-0001 — MVP language: Go

**Status:** Accepted
**Date:** 2026-06-16
**Decider:** Ad Soares
**Product context:** Harpo — a local secret broker for AI coding agents (see `harpo-mvp-spec.md`)

---

## Context

The MVP spec (§14.2) recommended .NET 10/8 and listed Go and Rust as strong alternatives, without locking the choice. Because the language is a hard-to-reverse decision (a near one-way door) and the Runner — the core of the product — relies heavily on spawning a child process with a controlled environment, the choice had to be settled before any scaffolding.

The central discussion was **Go vs Rust**, given that Harpo operates in the security domain.

## Decision

**The Harpo MVP will be written in Go.**

## Alternatives considered

### Rust
- **For:** GC-free memory enables deterministic zeroization of secrets (`zeroize`, `secrecy`), types that prevent accidental `Debug`/log exposure of values, and strong market perception as a security tool.
- **Against:** real learning curve (not part of the founder's current stack), more verbose `exec`/env handling, less mature cloud SDKs for future providers, slower delivery.

### Go (chosen)
- **For:** idiomatic `os/exec` + environment handling — exactly the heart of the Runner (resolve secrets, filter the environment, strip `BW_SESSION`, inject only what is authorized); trivial cross-compilation and single binary; mature cloud SDKs for the phase-3 providers; already part of the founder's stack; fast MVP delivery.
- **Against:** GC makes in-memory secret zeroization fragile; "security tool" perception slightly below Rust.

### .NET (the spec's original recommendation)
- Good productivity for the founder, but single-binary distribution and market perception for a security CLI lag behind Go/Rust. Rejected.

## Rationale

1. **The MVP's value is in broker correctness and DX**, not in cryptographic memory hardening: policy engine, session grants, stripping `BW_SESSION`, audit-without-value, best-effort redaction. Go delivers this well and fast.
2. **The dominant leak surface is the env var handed to the child process** (MVP spec §7.2), which is plaintext inside the child regardless of language. Rust's main advantage (in-process zeroization) protects a link that is not the weakest one at this stage.
3. **Idiomatic at the most central point:** spawning a process with a filtered environment is exactly Go's strength.
4. **Reduced delivery friction:** stack already known to the founder; a shipped, useful product matters more than the ideal stack (MVP spec §24).

## Consequences

- MVP scaffolding in Go (CLI framework and libraries defined at the start of implementation; see `CLAUDE.md`).
- Even with a GC, secrets must be treated as short-lived `[]byte`, never placed in types with a loggable `String()`/`Error()`, and discarded as soon as possible.
- Distribution via single binary + cross-compilation (Windows/Linux/macOS).

## Future reevaluation

Rust is **not rejected**. If security positioning becomes the product's competitive axis ("Security model" page, public auditing, enterprise audience), rewriting the core in Rust becomes justifiable — with `secrecy`/`zeroize` becoming part of the product narrative. Trigger: real traction + security as a market-required differentiator. Not before there is a user.
