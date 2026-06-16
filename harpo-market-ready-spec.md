# Harpo — Especificação da Versão Pronta para Mercado

**Status:** Draft v0.1  
**Data:** 2026-06-16  
**Tipo:** Produto open source com potencial comercial para segurança de secrets em fluxos de desenvolvimento com agentes de IA  
**Nome:** Harpo  
**Tagline principal:** _A local secret broker for AI coding agents._  
**Categoria proposta:** Agentic Secrets Security / AI Agent Secret Broker

---

## 1. Visão de produto

O **Harpo** é uma ferramenta open source que cria uma camada segura, local e auditável entre **agentes de programação com IA** e **secrets armazenados em cofres existentes**.

O produto nasce para resolver uma dor emergente do mercado: desenvolvedores estão usando agentes como Claude Code, Codex, Cursor, Windsurf, Gemini CLI e outros para executar tarefas reais, mas ainda precisam fornecer credenciais, tokens e chaves de acesso de maneira improvisada.

O Harpo deve se posicionar como:

> **O broker local de secrets para desenvolvimento agentic.**

Ou, em inglês:

> **Harpo is the local secret broker that lets AI coding agents use credentials without seeing your vault, your raw secrets, or your long-lived environment.**

---

## 2. Tese de mercado

A adoção de agentes de IA no desenvolvimento de software cria uma nova superfície de risco:

- agentes executam comandos;
- agentes leem arquivos;
- agentes editam código;
- agentes chamam CLIs;
- agentes acessam APIs;
- agentes podem ser influenciados por prompt injection em README, issues, logs, dependências e arquivos do repositório.

Ao mesmo tempo, o fluxo produtivo exige acesso a secrets:

- repositórios privados;
- registries;
- provedores cloud;
- bancos de desenvolvimento;
- APIs internas;
- ferramentas de observabilidade;
- plataformas de CI/CD.

O mercado já possui cofres de secrets, mas o problema específico é outro:

```text
Como permitir que um agente de programação use credenciais temporárias e limitadas sem dar a ele acesso direto ao cofre e sem obrigar o desenvolvedor a colar tokens no prompt?
```

O Harpo ocupa esse espaço.

---

## 3. Posicionamento

### 3.1 Categoria

```text
Agentic Secrets Security
```

ou

```text
AI Agent Secret Broker
```

### 3.2 Frase curta

```text
Harpo gives AI coding agents scoped, temporary access to secrets from your existing vaults.
```

### 3.3 Frase longa

```text
Harpo is an open source local secret broker for AI coding agents. It lets developers and teams grant temporary, scoped, auditable access to credentials from existing vaults without pasting tokens into prompts, committing .env files, or giving agents direct access to the vault.
```

### 3.4 Frase em português

```text
Harpo é um broker local de secrets para agentes de programação com IA. Ele permite autorizar acesso temporário, limitado e auditável a credenciais de cofres existentes sem colar tokens no prompt, sem commitar .env e sem dar ao agente acesso direto ao vault.
```

---

## 4. Por que open source?

Harpo deve ser open source porque:

- segurança exige confiança e auditabilidade;
- desenvolvedores querem entender como secrets são tratados;
- integração com agentes e CLIs exige comunidade;
- providers de cofre mudam com frequência;
- adoção bottom-up em times de engenharia é mais provável com CLI livre;
- empresas podem adotar o core antes de contratar recursos avançados.

### 4.1 Modelo recomendado

```text
Open source core + commercial extensions opcionais
```

Core open source:

- CLI local;
- providers principais;
- session grants;
- policy local;
- audit local;
- agent setup;
- plugin SDK básico.

Possíveis ofertas comerciais futuras:

- console web para times;
- policy management centralizado;
- audit log centralizado;
- SIEM integration;
- enterprise SSO;
- aprovação remota;
- gestão de rotação;
- compliance reports;
- marketplace de policies;
- suporte empresarial.

---

## 5. Diferenciação

Harpo não deve competir diretamente com Bitwarden, 1Password, HashiCorp Vault, AWS Secrets Manager ou Azure Key Vault.

Harpo deve se posicionar como camada complementar.

### 5.1 Não é um password manager

O Harpo não guarda sua vida digital.

### 5.2 Não é um secrets manager cloud

O Harpo não tenta substituir infra corporativa de secrets.

### 5.3 Não é apenas um `.env` manager

O Harpo evita `.env` como padrão e usa `.env` apenas como modo de compatibilidade.

### 5.4 Não é apenas um wrapper do Bitwarden CLI

O Harpo adiciona:

- session grants;
- TTL;
- profiles;
- policy-as-code;
- agent setup;
- audit;
- redaction;
- execução controlada de processo filho;
- provider abstraction;
- padrões específicos para agentes de IA.

---

## 6. Princípio estratégico

O Harpo deve executar o agente sempre que possível.

Fluxo preferido:

```bash
harpo run --profile graces-dev -- claude
```

Fluxo menos seguro, mas suportável:

```bash
claude
# agente chama harpo depois
```

A diferença é importante:

```text
Quando Harpo executa o agente, Harpo controla o ambiente desde o início.
Quando o agente chama Harpo depois, parte do controle já foi perdida.
```

---

## 7. Arquitetura target

```text
┌────────────────────────────────────┐
│ Developer / Team                   │
└─────────────────┬──────────────────┘
                  │
                  ▼
┌────────────────────────────────────┐
│ Harpo CLI                          │
│ - commands                         │
│ - TUI approval                     │
│ - agent setup                      │
└─────────────────┬──────────────────┘
                  │
                  ▼
┌────────────────────────────────────┐
│ Harpo Core                         │
│ - policy engine                    │
│ - session grants                   │
│ - runner                           │
│ - redactor                         │
│ - audit                            │
└─────────────────┬──────────────────┘
                  │
        ┌─────────┴─────────┐
        ▼                   ▼
┌─────────────────┐ ┌──────────────────────┐
│ Provider SDK    │ │ Agent Integration     │
│ - Bitwarden     │ │ - Claude Code         │
│ - 1Password     │ │ - Codex CLI           │
│ - Vault         │ │ - Cursor/Windsurf     │
│ - AWS/GCP/Azure │ │ - MCP server          │
└─────────────────┘ └──────────────────────┘
        │                   │
        ▼                   ▼
┌─────────────────┐ ┌──────────────────────┐
│ Existing Vaults │ │ AI Coding Agents      │
└─────────────────┘ └──────────────────────┘

Optional future:

┌────────────────────────────────────┐
│ Harpo Team Console                 │
│ - central policies                 │
│ - audit aggregation                │
│ - approvals                        │
│ - compliance                       │
└────────────────────────────────────┘
```

---

## 8. Funcionalidades target

## 8.1 Core CLI

Comandos principais:

```bash
harpo init
harpo provider add
harpo provider login
harpo provider status
harpo secret map
harpo secret list
harpo secret test
harpo profile create
harpo profile add-secret
harpo session start
harpo session status
harpo session revoke
harpo run
harpo exec
harpo env render
harpo audit list
harpo policy validate
harpo doctor
harpo agent setup
harpo scan
```

---

## 8.2 Provider SDK

Harpo deve suportar providers por plugin.

### Providers prioritários

| Provider | Prioridade | Observação |
|---|---:|---|
| Bitwarden Password Manager | Alta | ótimo para dev solo |
| Bitwarden Secrets Manager | Alta | melhor para uso profissional e machine accounts |
| 1Password | Alta | forte adoção por devs e times |
| HashiCorp Vault | Alta | padrão em empresas e infra |
| AWS Secrets Manager | Média | relevante para cloud AWS |
| GCP Secret Manager | Média | relevante para cloud GCP |
| Azure Key Vault | Média | relevante para empresas Microsoft |
| Doppler | Média | popular como secrets manager developer-first |
| Infisical | Média | open source e developer-first |
| Local encrypted provider | Baixa/Média | útil para demo, testes e dev offline |

### Provider capabilities

Cada provider deve expor capabilities:

```json
{
  "supportsScopedAccess": true,
  "supportsAudit": true,
  "supportsRotation": false,
  "supportsDynamicSecrets": false,
  "supportsMachineIdentity": true,
  "supportsReadByAlias": true,
  "supportsList": true
}
```

Isso permite ao Harpo alertar o usuário quando a segurança real depende apenas de escopo lógico local.

---

## 8.3 Policy-as-code

Arquivo `harpo.yml` deve evoluir para política completa.

Exemplo:

```yaml
version: 1

project:
  name: graces-api
  environments: [dev, staging]
  protected_environments: [production]

providers:
  bw-personal:
    type: bitwarden-password-manager

secrets:
  gitlab.ad.read:
    provider: bw-personal
    ref: "gitlab.com | ad | PAT | claude-code | read_api"
    field: password
    classification: internal
    allowed_agents: [claude, codex]
    allowed_destinations:
      - env:GITLAB_TOKEN
    max_ttl: 2h

profiles:
  graces-dev:
    mode: strict
    environment: dev
    agents: [claude, codex]
    ttl: 2h
    secrets:
      - gitlab.ad.read -> env:GITLAB_TOKEN

policies:
  require_ttl: true
  max_ttl: 8h
  allow_reveal: false
  allow_dotenv: false
  require_audit: true
  deny_production_by_default: true
  deny_direct_vault_cli_for_agents: true
  deny_env_printing: true
```

---

## 8.4 Session grants

Session grant é o conceito de produto mais importante.

Um grant deve responder:

```text
Quem pode usar?
Qual agente/processo?
Qual projeto?
Qual secret?
Qual destino?
Por quanto tempo?
Com qual modo de segurança?
O uso foi auditado?
```

Exemplo:

```json
{
  "id": "sess_20260616_abc123",
  "subject": "local-user",
  "agent": "claude",
  "project": "graces-api",
  "project_path": "C:/dev/graces-api",
  "profile": "graces-dev",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "grants": [
    {
      "secret": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN",
      "mode": "run-only"
    }
  ]
}
```

---

## 8.5 Delivery modes

### 8.5.1 `run` — modo recomendado

```bash
harpo run --profile graces-dev -- claude
```

Harpo inicia o processo filho e controla o ambiente.

### 8.5.2 `exec` — comando pontual

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

### 8.5.3 `env render` — compatibilidade

```bash
harpo env render --profile graces-dev --out .harpo/.env.session --ttl 30m
```

### 8.5.4 `proxy` — versão avançada

Em vez de entregar o token ao processo, Harpo pode agir como proxy local:

```bash
harpo proxy start --profile gitlab-dev
```

O agente chama:

```bash
harpo gitlab projects list
```

Ou acessa um endpoint local autorizado:

```text
http://127.0.0.1:<port>/gitlab/projects
```

Neste modo, o token nunca entra no ambiente do agente.

### 8.5.5 `fd/stdin` — versão avançada

Para ferramentas compatíveis:

```bash
harpo exec --secret gitlab.ad.read --stdin -- some-command
```

Ou uso de file descriptor temporário para evitar env vars.

---

## 8.6 Agent integrations

### Claude Code

Harpo deve gerar:

- `CLAUDE.md`;
- `.claude/settings.local.json`, quando aplicável;
- regras de deny para `.env`, `.harpo/`, `bw`, `op`, `vault`, `printenv`, `env`, `set`;
- instruções de uso de `harpo`;
- comandos recomendados.

Claude Code possui mecanismos de permissões e sandbox que podem bloquear acesso a recursos restritos e limitar comandos. Harpo deve se integrar a esses mecanismos sempre que possível.

### Codex CLI

Harpo deve gerar:

- `AGENTS.md`;
- recomendações de sandbox;
- recomendações de approval policy;
- regras locais quando disponíveis;
- comandos recomendados como:

```bash
harpo run --profile graces-dev -- codex --sandbox workspace-write --ask-for-approval on-request
```

Codex separa tecnicamente sandbox e approval policy: sandbox define limites técnicos; approval policy define quando o agente precisa pedir confirmação.

### Cursor, Windsurf e outros

Suporte inicial via documentação e `.env` temporário; suporte avançado via plugins ou MCP.

---

## 8.7 MCP server local

Versão de mercado deve incluir um MCP server local opcional:

```bash
harpo mcp start --profile graces-dev
```

O MCP não deve expor secrets crus por padrão.

Ferramentas MCP seguras:

```text
harpo.session.status
harpo.secret.available
harpo.exec.approved
harpo.gitlab.projects.list
harpo.aws.identity.get
```

Ferramentas perigosas, como `harpo.secret.reveal`, devem ser desabilitadas por padrão.

---

## 8.8 TUI de autorização

Uma interface terminal interativa deve permitir:

```text
Select secrets for this agent session:

[ ] aws.prod.root
[x] aws.dev.goritek
[x] gitlab.ad.read
[ ] database.prod.graces
[x] database.dev.graces

TTL: 2h
Mode: strict
Agent: claude
Project: graces-api

Confirm session grant? [y/N]
```

---

## 8.9 Redaction avançada

Harpo deve conseguir mascarar:

- secrets exatos;
- prefixos comuns;
- tokens conhecidos por formato;
- env vars sensíveis;
- outputs de `harpo exec`;
- logs internos;
- mensagens de erro.

Tipos conhecidos:

```text
GITLAB_TOKEN
GITHUB_TOKEN
gh[pousr]_
AKIA...
AWS_SECRET_ACCESS_KEY
GOOGLE_APPLICATION_CREDENTIALS
AZURE_CLIENT_SECRET
DATABASE_URL
JWT
OPENAI_API_KEY
ANTHROPIC_API_KEY
```

---

## 8.10 Secret scanning

Harpo deve incluir scanner local simples:

```bash
harpo scan
harpo scan --staged
harpo scan --history
```

Objetivo:

- alertar `.env` trackeado;
- detectar tokens em arquivos;
- detectar secrets em `harpo.yml`;
- instalar pre-commit hook opcional.

---

## 8.11 Rotation assistant

Harpo pode não rotacionar todos os secrets automaticamente, mas pode ajudar:

```bash
harpo rotate plan gitlab.ad.read
```

Saída:

```text
Secret: gitlab.ad.read
Provider: bitwarden
Target: GitLab PAT
Suggested rotation:
1. Create new GitLab PAT with same/minimal scopes.
2. Update Bitwarden item.
3. Run harpo secret test gitlab.ad.read.
4. Revoke old token in GitLab.
5. Confirm rotation complete.
```

Para providers com API de rotação, implementar:

```bash
harpo rotate gitlab.ad.read
```

---

## 8.12 Audit e compliance

Audit local:

```json
{
  "time": "2026-06-16T14:22:31-03:00",
  "event": "secret.injected",
  "user": "local-user",
  "agent": "claude",
  "project": "graces-api",
  "secret": "gitlab.ad.read",
  "destination": "env:GITLAB_TOKEN",
  "ttl": "2h",
  "policy": "strict",
  "result": "success"
}
```

Versão enterprise pode exportar para:

- JSONL;
- OpenTelemetry;
- SIEM;
- Datadog;
- Splunk;
- CloudWatch;
- Elastic;
- Grafana Loki.

---

## 9. Segurança target

### 9.1 Requisitos obrigatórios

- Secrets nunca devem ser gravados em logs.
- Secrets nunca devem ser impressos por padrão.
- Configuração versionável nunca deve conter secrets.
- Sessões devem ter TTL.
- Profiles de agente devem negar produção por padrão.
- Harpo deve remover tokens do provider do ambiente do agente.
- `.env` deve ser exceção explícita, não padrão.
- Redaction deve proteger stdout/stderr quando Harpo controla o processo.
- Harpo deve ser compatível com deny/sandbox dos agentes.
- Harpo deve ter testes de regressão contra vazamento.

### 9.2 Segurança por nível

#### Level 1 — Solo Developer

- provider pessoal;
- policy local;
- audit local;
- `run`, `exec`, `.env` temporário.

#### Level 2 — Team

- profiles compartilháveis;
- policy-as-code;
- providers de secrets manager;
- audit exportável;
- pre-commit scanning;
- aprovação por ambiente.

#### Level 3 — Enterprise

- SSO;
- RBAC;
- policy centralizada;
- SIEM;
- secrets manager corporativo;
- approval workflow;
- compliance reporting;
- dynamic secrets;
- central revocation.

---

## 10. Experiência de instalação

### 10.1 Instalação desejada

```bash
# macOS/Linux
curl -fsSL https://harpo.dev/install.sh | sh

# Windows PowerShell
iwr https://harpo.dev/install.ps1 -useb | iex

# Homebrew
brew install harpo

# Winget
winget install harpo

# Chocolatey
choco install harpo

# Docker, para automações
 docker run --rm harpo/harpo version
```

### 10.2 Primeiro uso ideal

```bash
harpo init
harpo provider add bitwarden
harpo secret import --interactive
harpo profile create dev-agent
harpo agent setup claude
harpo run --profile dev-agent -- claude
```

---

## 11. Developer experience

O Harpo deve ser:

- previsível;
- fácil de explicar;
- seguro por padrão;
- rápido;
- cross-platform;
- amigável para agentes;
- amigável para humanos;
- scriptável;
- audível.

### 11.1 Boas mensagens de erro

Ruim:

```text
Error: provider failed
```

Bom:

```text
Bitwarden vault is locked.
Run:
  bw unlock
Then retry:
  harpo run --profile graces-dev -- claude

Harpo will not pass BW_SESSION to the agent process.
```

---

## 12. Command surface target

```bash
harpo init
harpo doctor
harpo version

harpo provider add <name> --type <type>
harpo provider login <name>
harpo provider status [name]
harpo provider list
harpo provider remove <name>

harpo secret map <alias>
harpo secret import --interactive
harpo secret list
harpo secret show <alias> --metadata
harpo secret test <alias>
harpo secret remove <alias>

harpo profile create <name>
harpo profile list
harpo profile show <name>
harpo profile add-secret <profile> <secret>
harpo profile remove-secret <profile> <secret>

harpo session start --profile <name>
harpo session status
harpo session revoke [id|current]
harpo session list

harpo run --profile <name> -- <command>
harpo exec --with <secret:ENV> -- <command>
harpo env render --profile <name> --out <path>
harpo reveal <secret>

harpo audit list
harpo audit export

harpo policy validate
harpo policy explain

harpo scan
harpo scan --staged
harpo scan --history

harpo agent setup claude
harpo agent setup codex
harpo agent setup cursor
harpo agent setup windsurf

harpo mcp start --profile <name>
```

---

## 13. Open source governance

### 13.1 Repositório

Sugestão:

```text
github.com/harpo-sh/harpo
```

Alternativas:

```text
github.com/harpo-dev/harpo
github.com/harpocrates-dev/harpo
github.com/adnilson/harpo
```

### 13.2 Estrutura de repositório

```text
harpo/
  README.md
  LICENSE
  SECURITY.md
  CONTRIBUTING.md
  CODE_OF_CONDUCT.md
  docs/
    getting-started.md
    security-model.md
    providers.md
    agents/
      claude-code.md
      codex.md
    policies.md
    threat-model.md
  src/
  tests/
  examples/
    bitwarden-claude/
    bitwarden-codex/
    onepassword/
  scripts/
```

### 13.3 Documentos obrigatórios

- `README.md`
- `SECURITY.md`
- `CONTRIBUTING.md`
- `docs/security-model.md`
- `docs/threat-model.md`
- `docs/providers.md`
- `docs/agents/claude-code.md`
- `docs/agents/codex.md`

### 13.4 Licença

Recomendação: **Apache-2.0**.

Motivo:

- baixa fricção para adoção corporativa;
- boa aceitação em infraestrutura;
- linguagem explícita sobre patentes;
- permite monetização de serviços e extensões.

Alternativa estratégica:

- **MPL-2.0** se quiser proteger melhor o core contra forks fechados diretos.

Evitar inicialmente:

- AGPL, se o objetivo for adoção ampla em empresas que evitam licenças copyleft fortes.

---

## 14. Estratégia de diferenciação para mercado

### 14.1 Tema central

```text
AI agents should not need vault access to get work done.
```

### 14.2 Narrativa

Antes:

```text
Developers pasted secrets into prompts or stored them in .env files.
```

Depois:

```text
Developers grant scoped, temporary, auditable access to secrets through Harpo.
```

### 14.3 Principais mensagens

- “Stop pasting tokens into AI coding agents.”
- “Use your existing vault. Add agent-safe access.”
- “Temporary secrets for agentic development.”
- “Give agents what they need, not your whole vault.”
- “Policy-as-code for secrets used by coding agents.”

---

## 15. Possíveis páginas de documentação/marketing

### Home

Título:

```text
Harpo — Agent-safe secrets for AI coding workflows
```

Subtítulo:

```text
Grant Claude Code, Codex and other coding agents temporary access to credentials from your existing vaults without exposing your vault, pasting tokens or committing .env files.
```

### Security model

Explicar claramente:

- o que Harpo protege;
- o que Harpo não protege;
- riscos de env vars;
- riscos de `.env`;
- provider capabilities;
- threat model.

### Providers

Documentar:

- Bitwarden Password Manager;
- Bitwarden Secrets Manager;
- 1Password;
- Vault;
- AWS;
- GCP;
- Azure.

### Agents

Documentar:

- Claude Code;
- Codex;
- Cursor;
- Windsurf;
- GitHub Copilot;
- Gemini CLI.

---

## 16. Métricas de sucesso

### Open source

- GitHub stars.
- Instalações mensais.
- Issues abertas por usuários reais.
- Contribuidores externos.
- Providers contribuídos pela comunidade.
- Mentions em blogs, newsletters e comunidades de AI coding.

### Produto

- Número de profiles criados.
- Número de sessões iniciadas.
- Número de secrets acessados via grants.
- Uso de `harpo run` vs `.env`.
- Adoção por times.
- Redução de `.env` versionados.
- Integrações com agentes.

### Segurança

- Zero secrets em logs internos.
- Testes automatizados de leak prevention.
- Issues de segurança respondidas com SLA.
- Auditoria revisável.
- Threat model público.

---

## 17. Roadmap de mercado

### Fase 1 — MVP open source

Objetivo: provar valor para dev solo.

- CLI core.
- Bitwarden Password Manager provider.
- `harpo run`.
- `harpo exec`.
- `harpo env render`.
- `harpo agent setup claude/codex`.
- audit local.
- docs fortes.

### Fase 2 — Developer adoption

Objetivo: adoção por comunidade.

- Homebrew/Winget/Chocolatey.
- Provider 1Password.
- Provider Bitwarden Secrets Manager.
- `harpo doctor`.
- scanner básico.
- exemplos reais.
- website simples.
- vídeos curtos.

### Fase 3 — Team-ready

Objetivo: pequenos times.

- policy-as-code robusto;
- profiles compartilhados;
- audit exportável;
- provider Vault;
- AWS/GCP/Azure;
- hooks de pre-commit;
- modelos de `harpo.yml` por stack.

### Fase 4 — Agent-native platform

Objetivo: relevância em segurança para agentes.

- MCP server;
- TUI de autorização;
- proxy mode;
- dynamic secrets;
- command broker;
- integração com Claude/Codex avançada;
- plugin SDK.

### Fase 5 — Enterprise/commercial

Objetivo: monetização.

- console web;
- central policy;
- SSO;
- RBAC;
- remote approvals;
- audit aggregation;
- SIEM integrations;
- compliance reports;
- suporte empresarial.

---

## 18. Riscos de produto

| Risco | Mitigação |
|---|---|
| Ser visto como “wrapper do bw” | Focar em session grants, agent integrations, policy e audit |
| Segurança prometer mais do que entrega | Documentar limitações claramente |
| `.env` temporário virar padrão inseguro | `run` como default; `.env` com alerta e modo explícito |
| Muitos providers antes de produto maduro | Começar com 1-2 providers muito bons |
| Agentes mudarem APIs/permissions | Arquitetura modular por agent integration |
| Empresas desconfiarem de OSS novo | Threat model público, testes, security policy, releases assinadas |
| Dificuldade de monetizar | Core OSS + team/enterprise features |

---

## 19. Requisitos para parecer sério no mercado

Antes de divulgar amplamente, o Harpo deve ter:

- README excelente com problema, demo e threat model.
- Instalação simples.
- Demo GIF ou vídeo curto.
- `SECURITY.md` claro.
- Documentação de limitações.
- Testes automatizados.
- Releases assinadas ou checksums.
- Exemplos com Claude Code e Codex.
- Integração real com Bitwarden.
- Página “Why Harpo?” comparando com alternativas.
- Página “Security model”.
- Página “What Harpo does not protect against”.

---

## 20. README inicial sugerido

```md
# Harpo

Harpo is a local secret broker for AI coding agents.

It lets developers grant Claude Code, Codex and other coding agents temporary, scoped access to credentials from existing vaults without pasting tokens into prompts, committing .env files, or giving agents direct access to the vault.

## Why

AI coding agents can run commands, edit files and call developer tools. But real work often requires credentials. Pasting secrets into prompts or storing them in plaintext .env files is risky.

Harpo keeps your vault as the source of truth and gives agents only the secrets you explicitly authorize for the current project and session.

## Example

```bash
bw unlock
harpo run --profile my-project-dev -- claude
```

## Security model

- Harpo does not replace your vault.
- Harpo does not expose your vault session to agents.
- Harpo does not print secrets by default.
- Harpo uses temporary session grants with TTL.
- Harpo writes audit logs without secret values.
```

---

## 21. Estratégia de conteúdo

Temas de artigos:

1. “Stop pasting secrets into AI coding agents.”
2. “Why AI coding agents need scoped secret grants.”
3. “The new security problem created by agentic development.”
4. “How to use Bitwarden safely with Claude Code.”
5. “Why `.env` is not enough for AI-native development.”
6. “Designing a secret broker for coding agents.”
7. “Agentic Secrets Security: a new category for software teams.”

---

## 22. Decisões importantes de produto

### 22.1 `run` deve ser o caminho principal

Porque permite controlar o ambiente do processo desde o início.

### 22.2 `.env` deve ser compatibilidade, não default

Porque `.env` é plaintext em disco.

### 22.3 Provider Password Manager deve mostrar aviso

Porque cofre pessoal desbloqueado pode ter acesso amplo.

### 22.4 Secrets Manager deve ser recomendado para times

Porque permite melhor separação entre secrets pessoais e secrets de máquina/projeto.

### 22.5 O threat model deve ser honesto

Harpo reduz risco, mas não transforma agente local em ambiente perfeitamente isolado.

---

## 23. Modelo comercial possível

### Open Source Core

Grátis:

- CLI;
- providers básicos;
- local policies;
- audit local;
- agent setup;
- plugin SDK.

### Harpo Team

Pago:

- policy bundles;
- audit export;
- team profile templates;
- shared configuration validation;
- provider enterprise packs;
- support.

### Harpo Enterprise

Pago:

- console central;
- SSO/SAML/OIDC;
- RBAC;
- remote approvals;
- SIEM;
- compliance;
- dynamic secrets;
- customer support;
- private plugins.

### Serviços

- consultoria de implantação;
- hardening de workflows com Claude/Codex;
- criação de policies por empresa;
- treinamento de AI-native secure development.

---

## 24. Escolha de stack para mercado

### Opção A — Go

Vantagens:

- single binary simples;
- ótimo para CLI;
- fácil distribuição;
- boa performance;
- adoção forte em DevOps.

### Opção B — Rust

Vantagens:

- reputação forte em segurança;
- binário performático;
- controle fino;
- boa imagem para ferramenta de segurança.

### Opção C — .NET

Vantagens:

- excelente produtividade para criador com background .NET;
- cross-platform;
- single-file publish;
- bom ecossistema;
- Spectre.Console oferece ótima UX.

Recomendação pragmática:

```text
Começar em .NET se isso acelerar muito a entrega.
Reavaliar Go/Rust apenas se distribuição, performance ou percepção de mercado virarem gargalo.
```

Um produto entregue e útil é mais importante que a stack ideal.

---

## 25. Versão 1.0: definição de pronto para mercado

Harpo 1.0 deve ter:

- CLI estável.
- Instalação simples em Windows, macOS e Linux.
- Bitwarden Password Manager provider.
- Bitwarden Secrets Manager provider.
- 1Password provider.
- `harpo run` maduro.
- `harpo exec` maduro.
- `harpo env render` seguro.
- `harpo agent setup claude`.
- `harpo agent setup codex`.
- policy-as-code validável.
- audit local e exportável.
- scanner básico.
- docs de segurança.
- threat model público.
- releases assinadas/checksums.
- testes automatizados de leak prevention.

---

## 26. North Star

A métrica conceitual do Harpo é:

```text
Quantas vezes um desenvolvedor deixou de colar um secret no prompt ou salvar um .env inseguro porque o Harpo tornou o caminho seguro mais fácil?
```

O objetivo final não é apenas proteger secrets.

O objetivo é tornar o desenvolvimento com agentes de IA **mais produtivo, mais auditável e mais seguro por padrão**.

---

## 27. Frase final de posicionamento

```text
Harpo is the missing security layer between AI coding agents and developer secrets.
```

Em português:

```text
Harpo é a camada de segurança que faltava entre agentes de programação com IA e os secrets dos desenvolvedores.
```
