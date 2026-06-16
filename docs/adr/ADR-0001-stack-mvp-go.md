# ADR-0001 — Linguagem do MVP: Go

**Status:** Aceito
**Data:** 2026-06-16
**Decisor:** Ad Soares
**Contexto do produto:** Harpo — broker local de secrets para agentes de programação com IA (ver `harpo-mvp-spec.md`)

---

## Contexto

O MVP spec (§14.2) recomendava .NET 10/8 e listava Go e Rust como alternativas fortes, sem travar a escolha. Como a linguagem é uma decisão de difícil reversão (porta quase-de-mão-única) e o Runner — núcleo do produto — depende fortemente de spawnar processo filho com ambiente controlado, a escolha precisava ser fechada antes de qualquer scaffolding.

A discussão central foi **Go vs Rust**, dado que o Harpo atua no domínio de segurança.

## Decisão

**O MVP do Harpo será escrito em Go.**

## Alternativas consideradas

### Rust
- **A favor:** memória sem GC permite zeroização determinística de secrets (`zeroize`, `secrecy`), tipos que impedem `Debug`/log acidental do valor, e forte percepção de mercado como ferramenta de segurança.
- **Contra:** curva de aprendizado real (não faz parte do stack atual do fundador), maior verbosidade na parte de `exec`/env, SDKs de cloud menos maduros para providers futuros, entrega mais lenta.

### Go (escolhida)
- **A favor:** `os/exec` + manipulação de env idiomáticos — exatamente o coração do Runner (resolver secrets, filtrar ambiente, remover `BW_SESSION`, injetar só o autorizado); cross-compile e single binary triviais; SDKs de cloud maduros para os providers da fase 3; já faz parte do stack do fundador; entrega rápida do MVP.
- **Contra:** GC torna a zeroização de secret em memória frágil; percepção de "ferramenta de segurança" um pouco abaixo de Rust.

### .NET (recomendação original do spec)
- Boa produtividade para o fundador, mas distribuição single-binary e percepção de mercado para CLI de segurança ficam atrás de Go/Rust. Preterida.

## Justificativa

1. **O valor do MVP está na corretude do broker e na DX**, não em hardening criptográfico de memória: policy engine, session grants, strip de `BW_SESSION`, audit sem valor, redaction best-effort. Go entrega isso bem e rápido.
2. **A superfície de vazamento dominante é a env var entregue ao processo filho** (MVP spec §7.2), que é plaintext dentro do filho independentemente da linguagem. A principal vantagem de Rust (zeroização in-process) protege um elo que não é o mais fraco neste momento.
3. **Idiomático no ponto mais central:** spawnar processo com ambiente filtrado é exatamente o forte de Go.
4. **Redução de fricção de entrega:** stack já conhecido pelo fundador; um produto entregue e útil vale mais que a stack ideal (MVP spec §24).

## Consequências

- Scaffolding do MVP em Go (CLI framework e libs definidos no início da implementação; ver `CLAUDE.md`).
- Mesmo com GC, secrets devem ser tratados como `[]byte` de vida curta, nunca colocados em tipos com `String()`/`Error()` logáveis, e descartados assim que possível.
- Distribuição via single binary + cross-compile (Windows/Linux/macOS).

## Reavaliação futura

Rust **não está descartado**. Se o posicionamento de segurança se tornar o eixo competitivo do produto (página "Security model", auditoria pública, público enterprise), uma reescrita do core em Rust passa a ser justificável — com `secrecy`/`zeroize` virando parte da narrativa de produto. Gatilho: tração real + segurança como diferenciador exigido pelo mercado. Não antes de existir usuário.
