# Harpo — Especificação da Versão MVP

**Status:** Draft v0.1  
**Data:** 2026-06-16  
**Tipo:** CLI open source local para intermediação segura de secrets entre cofres e agentes de programação com IA  
**Nome:** Harpo  
**Tagline:** _A local secret broker for AI coding agents._

---

## 1. Visão do MVP

O **Harpo** é um CLI local que permite ao usuário autorizar, de forma temporária e limitada, o uso de credenciais armazenadas em um cofre por ferramentas como **Claude Code**, **Codex CLI**, scripts locais, CLIs de terceiros e automações de desenvolvimento.

O objetivo do MVP não é substituir Bitwarden, 1Password, Vault ou outros gerenciadores de secrets. O objetivo é criar uma camada local, simples e segura entre:

```text
Agente de IA / processo local
        ↓
      Harpo
        ↓
Provider de cofre existente
        ↓
Credencial autorizada pelo usuário
```

O princípio central do MVP é:

> **O agente não recebe acesso ao cofre. O agente recebe apenas acesso temporário, limitado e auditável a credenciais específicas que o usuário autorizou.**

---

## 2. Problema que o Harpo resolve

Desenvolvedores que usam agentes de programação com IA frequentemente precisam dar acesso a:

- tokens de GitLab, GitHub ou Azure DevOps;
- credenciais AWS, GCP ou Azure;
- chaves de APIs;
- dados de conexão de bancos de desenvolvimento;
- tokens de ferramentas internas;
- credenciais de ambientes temporários.

O fluxo comum é inseguro ou improdutivo:

```text
1. Usuário cria token.
2. Salva em arquivo texto, .env ou notas soltas.
3. Copia e cola no prompt do agente.
4. O segredo pode cair em histórico, transcript, shell, Git, logs ou screenshots.
```

O Harpo propõe um fluxo melhor:

```text
1. Usuário salva o segredo no cofre.
2. Usuário mapeia um alias local no Harpo.
3. Usuário autoriza uma sessão com TTL e escopo.
4. Harpo inicia o agente ou processo com apenas os secrets permitidos.
5. Harpo audita o uso sem armazenar o valor do secret.
```

---

## 3. Público-alvo do MVP

### Persona principal

**Desenvolvedor solo ou sênior usando agentes locais de programação**, especialmente:

- Claude Code;
- Codex CLI;
- GitHub Copilot CLI/agents;
- Cursor/Windsurf via terminal;
- scripts de automação local;
- CLIs como `glab`, `gh`, `aws`, `az`, `gcloud`.

### Persona secundária

**Pequenos times de desenvolvimento** que querem padronizar o uso seguro de credentials em projetos sem adotar, inicialmente, uma solução corporativa complexa.

---

## 4. Escopo do MVP

O MVP deve entregar um fluxo útil e seguro com poucos recursos, mas muito bem acabados.

### Incluído no MVP

- CLI local `harpo`.
- Provider inicial para **Bitwarden Password Manager via Bitwarden CLI (`bw`)**.
- Arquivo `harpo.yml` versionável, sem valores secretos.
- Mapeamento de aliases locais para itens/campos do cofre.
- Profiles de projeto.
- Sessões temporárias com TTL.
- Execução de processo filho com secrets injetados como variáveis de ambiente.
- Renderização opcional de `.env` temporário dentro de `.harpo/`.
- Setup básico para Claude Code e Codex.
- Auditoria local sem armazenar segredo.
- Redaction básica para comandos executados via Harpo.
- Modo `strict` e `balanced`.

### Fora do MVP

- UI desktop.
- Serviço cloud.
- Compartilhamento de secrets entre usuários.
- Rotação automática de tokens.
- RBAC corporativo.
- SIEM/log centralizado.
- Provider completo para todos os cofres.
- Daemon residente.
- Sincronização remota de políticas.
- MCP server.
- Plugin oficial para IDEs.

---

## 5. Princípios de design

### 5.1 Secure by default

O Harpo deve nascer seguro por padrão. A operação padrão não deve imprimir secrets, não deve criar `.env` persistente e não deve expor a sessão do cofre ao agente.

### 5.2 Convenient by explicit choice

Fluxos mais convenientes, como criar `.env`, revelar secret no terminal ou aumentar TTL, devem existir, mas exigir escolha explícita do usuário.

### 5.3 Agent-safe, not agent-blind

O Harpo deve aceitar que agentes de IA podem executar comandos, interpretar arquivos e sofrer prompt injection. Por isso, o controle não pode depender apenas de instruções em `CLAUDE.md` ou `AGENTS.md`.

### 5.4 Least privilege

Cada sessão deve liberar apenas os secrets autorizados, no destino autorizado e pelo tempo autorizado.

### 5.5 No vault access for agents

O agente não deve receber `BW_SESSION`, token do cofre, master password, API key do provider ou qualquer capacidade de listar o cofre.

### 5.6 Auditable without leaking

O Harpo deve registrar o que foi usado, quando, por qual profile e em qual projeto, sem registrar o valor do secret.

---

## 6. Referências de segurança usadas no design

O design do MVP deve se alinhar a boas práticas reconhecidas:

- OWASP Secrets Management Cheat Sheet: centralização, rotação, auditoria e menor privilégio para secrets.
- Bitwarden CLI: o estado `unlocked` depende de uma chave ativa em `BW_SESSION`; o estado `locked` indica ausência dessa sessão ativa.
- Claude Code: permissões e sandbox podem bloquear acesso a recursos restritos; deny rules ajudam a impedir tentativas de acesso a arquivos e comandos sensíveis.
- Codex CLI: sandbox e approval policy são controles diferentes e complementares; o modo local pode operar com `workspace-write` e aprovação `on-request`.

Links:

- https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html
- https://bitwarden.com/help/cli/
- https://code.claude.com/docs/en/permissions
- https://code.claude.com/docs/en/sandboxing
- https://developers.openai.com/codex/agent-approvals-security
- https://developers.openai.com/codex/concepts/sandboxing
- https://developers.openai.com/codex/cli/reference

---

## 7. Threat model do MVP

### 7.1 Ameaças consideradas

| Ameaça | Mitigação no MVP |
|---|---|
| Agente tenta listar o cofre | Harpo não expõe `BW_SESSION` ao processo filho; setup recomenda negar `bw`, `op`, `vault` |
| Secret cai em prompt/transcript | Harpo não imprime secret por padrão |
| Secret cai no histórico do shell | Harpo evita secrets em argumentos de comando e stdout |
| Secret é commitado em `.env` | `.env` temporário fica em `.harpo/`; Harpo adiciona `.harpo/` ao `.gitignore` |
| Prompt injection manda ler `.env` | Setup gera regras/instruções para bloquear leitura; Harpo recomenda sandbox/deny rules |
| Sessão longa demais | TTL obrigatório para profiles de agente |
| Reuso indevido em outro projeto | Sessão pode ser amarrada ao path do projeto |
| Usuário autoriza secret errado | Fluxo interativo mostra alias, provider, campo, escopo e destino antes de confirmar |
| Output de comando vaza secret | Redaction básica para comandos executados via `harpo exec` |

### 7.2 Ameaças não resolvidas completamente no MVP

| Limitação | Observação |
|---|---|
| Variáveis de ambiente podem ser lidas pelo próprio processo filho | Harpo reduz escopo, mas env vars continuam sendo plaintext dentro do processo |
| Agente com acesso shell pode tentar imprimir variáveis que recebeu | Mitigar com permissões, instruções, deny rules e modo `strict`; não há garantia absoluta |
| `.env` temporário é plaintext em disco | Permitido apenas em modo explícito, com aviso, TTL e path seguro |
| Provider Bitwarden Password Manager não fornece escopo fino por item no próprio CLI | Harpo aplica escopo lógico; para escopo forte, usar futuramente Bitwarden Secrets Manager |
| Redaction em TUI interativo pode não ser confiável | MVP deve evitar prometer redaction total no modo `harpo run -- claude` |

---

## 8. Modos de segurança

### 8.1 `strict`

Modo recomendado para agentes de IA.

Regras:

- TTL obrigatório.
- `harpo run` como modo principal.
- Sem `reveal`.
- Sem `.env` por padrão.
- Sem wildcard de secrets.
- Sem herdar `BW_SESSION` para processo filho.
- Auditoria obrigatória.
- Confirmação interativa para secrets novos.

### 8.2 `balanced`

Modo recomendado para desenvolvedor solo.

Regras:

- TTL recomendado, mas configurável.
- `.env` temporário permitido dentro de `.harpo/`.
- `reveal` permitido com confirmação forte.
- Auditoria obrigatória.
- Avisos claros para ações perigosas.

### 8.3 Fora do MVP: `convenient`

Modo de conveniência extrema, a ser avaliado depois.

---

## 9. Arquitetura do MVP

```text
┌─────────────────────────┐
│ Usuário                  │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Harpo CLI                │
│ - comandos               │
│ - UX interativa          │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Policy Engine            │
│ - modo                   │
│ - TTL                    │
│ - profile                │
│ - path autorizado        │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Session Manager          │
│ - grants                 │
│ - expiração              │
│ - metadata               │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Provider Adapter         │
│ - Bitwarden CLI          │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│ Vault existente          │
│ - Bitwarden              │
└─────────────────────────┘

┌─────────────────────────┐
│ Runner                  │
│ - inicia Claude/Codex    │
│ - injeta env filtrado    │
│ - remove BW_SESSION      │
└─────────────────────────┘

┌─────────────────────────┐
│ Audit Logger             │
│ - registra acesso        │
│ - nunca registra valor   │
└─────────────────────────┘
```

---

## 10. Componentes

### 10.1 CLI Core

Responsável por parsing de comandos, UX, validações, mensagens de erro e integração entre módulos.

### 10.2 Provider Adapter

Interface para integração com cofres.

MVP:

```text
bitwarden-password-manager
```

Implementação inicial pode chamar o binário `bw` instalado localmente.

Responsabilidades:

- verificar se `bw` está instalado;
- verificar `bw status`;
- acionar `bw sync` quando solicitado;
- buscar campo específico do item autorizado;
- nunca listar o cofre para o agente;
- nunca retornar dados para stdout diretamente.

### 10.3 Policy Engine

Valida:

- profile;
- modo de segurança;
- TTL;
- path do projeto;
- destino permitido;
- se `.env` é permitido;
- se `reveal` é permitido;
- se o secret solicitado está autorizado.

### 10.4 Session Manager

Cria e gerencia grants temporários.

A sessão armazena apenas metadata:

```json
{
  "id": "sess_abc123",
  "created_at": "2026-06-16T10:00:00-03:00",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "project_path": "C:/dev/graces-api",
  "agent": "claude",
  "profile": "graces-dev",
  "allowed_secrets": [
    {
      "alias": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN"
    }
  ]
}
```

Não deve armazenar valores secretos.

### 10.5 Runner

Executa processo filho com ambiente controlado:

```text
harpo run --profile graces-dev -- claude
```

Responsabilidades:

- resolver secrets autorizados;
- montar environment do processo filho;
- remover variáveis perigosas do ambiente herdado, especialmente `BW_SESSION`;
- injetar somente variáveis autorizadas;
- registrar auditoria;
- encerrar sessão ao fim se configurado.

### 10.6 Env Renderer

Renderiza `.env` temporário.

Regras:

- path padrão: `.harpo/.env.session`;
- exigir confirmação;
- criar `.harpo/` se não existir;
- adicionar `.harpo/` ao `.gitignore`;
- mostrar TTL;
- apagar arquivo ao revogar sessão quando possível;
- registrar auditoria.

### 10.7 Audit Logger

Registra eventos sem secrets.

Exemplo:

```json
{
  "time": "2026-06-16T14:22:31-03:00",
  "event": "secret.injected",
  "profile": "graces-dev",
  "agent": "claude",
  "project": "graces-api",
  "secret_alias": "gitlab.ad.read",
  "destination": "env:GITLAB_TOKEN",
  "mode": "run",
  "ttl_seconds": 7200,
  "result": "success"
}
```

### 10.8 Redactor

MVP:

- redaction para comandos executados via `harpo exec`;
- redaction best effort para erros do Harpo;
- não prometer redaction completa para TUI interativo.

---

## 11. Estrutura de arquivos

### 11.1 `harpo.yml`

Arquivo versionável, sem secrets.

```yaml
version: 1

project:
  name: graces-api
  allowed_paths:
    - .

mode: strict

providers:
  bitwarden-personal:
    type: bitwarden-password-manager

secrets:
  gitlab.ad.read:
    provider: bitwarden-personal
    ref: "gitlab.com | ad | PAT | claude-code | read_api"
    field: password
    default_env: GITLAB_TOKEN
    tags: [gitlab, dev, agent-safe]

profiles:
  graces-dev:
    ttl: 2h
    agent: claude
    secrets:
      - secret: gitlab.ad.read
        env: GITLAB_TOKEN

policies:
  allow_dotenv: false
  allow_reveal: false
  default_ttl: 2h
  max_ttl: 8h
```

### 11.2 `.harpo/`

Diretório local não versionado.

```text
.harpo/
  sessions/
  audit.log.jsonl
  .env.session
```

### 11.3 `.gitignore`

O Harpo deve garantir:

```gitignore
.harpo/
.env
.env.*
!.env.example
```

---

## 12. Comandos do MVP

### 12.1 `harpo init`

Inicializa o projeto.

```bash
harpo init
```

Cria:

- `harpo.yml`;
- `.harpo/`;
- entrada no `.gitignore`;
- instruções iniciais.

Opções:

```bash
harpo init --mode strict
harpo init --agent claude
harpo init --agent codex
```

---

### 12.2 `harpo provider add`

Adiciona provider.

```bash
harpo provider add bitwarden-personal --type bitwarden-password-manager
```

Valida:

- se `bw` existe no PATH;
- se `bw status` está acessível;
- se o cofre está `unlocked` ou `locked`.

---

### 12.3 `harpo provider status`

```bash
harpo provider status
```

Exemplo de saída:

```text
Provider: bitwarden-personal
Type: bitwarden-password-manager
CLI: bw found
Vault status: unlocked
Safe for agent inheritance: no
```

---

### 12.4 `harpo secret map`

Mapeia alias local para item do cofre.

```bash
harpo secret map gitlab.ad.read \
  --provider bitwarden-personal \
  --ref "gitlab.com | ad | PAT | claude-code | read_api" \
  --field password \
  --env GITLAB_TOKEN
```

Deve mostrar confirmação:

```text
Alias: gitlab.ad.read
Provider: bitwarden-personal
Vault ref: gitlab.com | ad | PAT | claude-code | read_api
Field: password
Default destination: env:GITLAB_TOKEN

No secret value will be stored in harpo.yml.
Confirm? [y/N]
```

---

### 12.5 `harpo secret list`

Lista aliases, nunca valores.

```bash
harpo secret list
```

---

### 12.6 `harpo secret test`

Testa se o Harpo consegue resolver o secret, sem imprimir valor.

```bash
harpo secret test gitlab.ad.read
```

Saída:

```text
Secret resolved successfully.
Length: 20 chars
Fingerprint: sha256:ab12...9f
Value: [redacted]
```

---

### 12.7 `harpo profile create`

Cria profile.

```bash
harpo profile create graces-dev --ttl 2h --agent claude
```

---

### 12.8 `harpo profile add-secret`

```bash
harpo profile add-secret graces-dev gitlab.ad.read --env GITLAB_TOKEN
```

---

### 12.9 `harpo session start`

Cria sessão explícita.

```bash
harpo session start --profile graces-dev --ttl 2h
```

---

### 12.10 `harpo session status`

```bash
harpo session status
```

Saída:

```text
Session: sess_abc123
Profile: graces-dev
Agent: claude
Project: C:/dev/graces-api
Expires in: 01:42:10
Secrets: 1
- gitlab.ad.read -> env:GITLAB_TOKEN
```

---

### 12.11 `harpo session revoke`

```bash
harpo session revoke current
```

Remove metadata local e apaga `.env` temporário associado, se existir.

---

### 12.12 `harpo run`

Modo principal.

```bash
harpo run --profile graces-dev -- claude
```

```bash
harpo run --profile graces-dev -- codex --sandbox workspace-write --ask-for-approval on-request
```

Regras:

- Harpo resolve secrets autorizados.
- Harpo remove `BW_SESSION` do ambiente herdado pelo processo filho.
- Harpo injeta apenas secrets do profile.
- Harpo registra auditoria.
- Harpo não imprime valores.

---

### 12.13 `harpo exec`

Executa comando específico com secret temporário.

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

Recomendado para comandos pontuais.

---

### 12.14 `harpo env render`

Renderiza `.env` temporário.

```bash
harpo env render --profile graces-dev --out .harpo/.env.session --ttl 30m
```

Deve exibir alerta:

```text
This will write plaintext secrets to disk.
Path: .harpo/.env.session
This path is ignored by Git.
Expires in: 30m
Confirm? [y/N]
```

---

### 12.15 `harpo audit list`

```bash
harpo audit list
```

Não mostra secrets.

---

### 12.16 `harpo agent setup claude`

Gera arquivos de orientação para Claude Code.

```bash
harpo agent setup claude
```

Pode criar/atualizar:

- `CLAUDE.md`;
- `.claude/settings.local.json` quando possível;
- `.gitignore`.

Conteúdo sugerido para `CLAUDE.md`:

```md
# Secrets policy

Use Harpo para secrets.

Permitido:
- harpo run
- harpo exec
- harpo session status

Proibido:
- bw
- op
- vault
- env
- printenv
- set
- Get-ChildItem Env:
- cat .env
- type .env
- git add .env
- git add .harpo
- harpo reveal

Nunca imprima secrets.
Nunca peça tokens no prompt.
Nunca grave secrets em arquivos versionados.
```

---

### 12.17 `harpo agent setup codex`

Gera `AGENTS.md` com política de secrets.

```bash
harpo agent setup codex
```

Conteúdo sugerido:

```md
# Secrets policy

Use Harpo para credenciais.

Quando precisar de credenciais:
- harpo session status
- harpo exec com comandos aprovados

Não execute diretamente:
- bw
- op
- vault
- printenv
- env
- cat .env
- type .env

Modo recomendado para Codex:
- workspace-write
- approval on-request
```

---

## 13. UX principal do MVP

### 13.1 Primeiro uso

```bash
harpo init --mode strict --agent claude
harpo provider add bitwarden-personal --type bitwarden-password-manager
harpo secret map gitlab.ad.read --provider bitwarden-personal --ref "gitlab.com | ad | PAT | claude-code | read_api" --field password --env GITLAB_TOKEN
harpo profile create graces-dev --ttl 2h --agent claude
harpo profile add-secret graces-dev gitlab.ad.read --env GITLAB_TOKEN
harpo agent setup claude
harpo run --profile graces-dev -- claude
```

### 13.2 Uso diário

```bash
bw unlock
harpo run --profile graces-dev -- claude
```

### 13.3 Comando pontual

```bash
harpo exec --with gitlab.ad.read:GITLAB_TOKEN -- glab repo list
```

---

## 14. Requisitos técnicos

### 14.1 Cross-platform

O MVP deve funcionar em:

- Windows 10/11 com PowerShell;
- Linux;
- macOS.

### 14.2 Stack recomendada

Opção recomendada para o criador do projeto:

- **.NET 10 Console** ou **.NET 8 LTS** se quiser estabilidade de LTS;
- `System.CommandLine` para CLI;
- `Spectre.Console` para UX terminal;
- `YamlDotNet` para `harpo.yml`;
- testes com xUnit;
- empacotamento como single-file publish.

Alternativas fortes para mercado:

- Go: distribuição simples, single binary, ótimo para CLI.
- Rust: excelente para segurança e distribuição, mas maior curva.

Como MVP, a melhor escolha é a que reduz fricção de desenvolvimento e permite entrega rápida.

---

## 15. Requisitos de segurança do MVP

### Obrigatórios

- Nunca imprimir secret por padrão.
- Nunca registrar secret em audit log.
- Nunca salvar valor do secret em `harpo.yml`.
- Nunca passar `BW_SESSION` para processo filho iniciado por `harpo run`.
- TTL obrigatório para profiles de agente.
- `.harpo/` deve ser ignorado pelo Git.
- `.env` temporário deve exigir confirmação.
- `reveal` deve ser desabilitado no modo `strict`.
- `secret test` deve mostrar apenas fingerprint parcial e metadata segura.
- Erros devem redigir valores sensíveis quando detectados.

### Recomendados

- Detectar se `.env` está trackeado no Git e alertar.
- Detectar se `harpo.yml` contém padrões parecidos com tokens.
- Alertar se TTL for maior que 8h.
- Alertar se secret alias parecer produção, exemplo: `prod`, `production`, `root`, `admin`.
- Alertar se usuário tentar usar `.env` fora de `.harpo/`.

---

## 16. Modelo de dados interno

### Secret alias

```json
{
  "alias": "gitlab.ad.read",
  "provider": "bitwarden-personal",
  "ref": "gitlab.com | ad | PAT | claude-code | read_api",
  "field": "password",
  "default_env": "GITLAB_TOKEN",
  "tags": ["gitlab", "dev", "agent-safe"]
}
```

### Profile

```json
{
  "name": "graces-dev",
  "agent": "claude",
  "ttl": "2h",
  "secrets": [
    {
      "alias": "gitlab.ad.read",
      "destination": "env:GITLAB_TOKEN"
    }
  ]
}
```

### Session grant

```json
{
  "id": "sess_abc123",
  "profile": "graces-dev",
  "agent": "claude",
  "project_path": "C:/dev/graces-api",
  "created_at": "2026-06-16T10:00:00-03:00",
  "expires_at": "2026-06-16T12:00:00-03:00",
  "allowed_secrets": ["gitlab.ad.read"]
}
```

---

## 17. Provider interface

Interface conceitual:

```ts
interface SecretProvider {
  id: string;
  type: string;

  status(): Promise<ProviderStatus>;
  authenticate(): Promise<AuthResult>;
  resolveSecret(ref: SecretRef): Promise<SecretValue>;
  testSecret(ref: SecretRef): Promise<SecretTestResult>;
  capabilities(): ProviderCapabilities;
}
```

Capabilities:

```json
{
  "canList": true,
  "canReadByRef": true,
  "supportsScopedAccess": false,
  "supportsAudit": false,
  "supportsRotation": false,
  "supportsDynamicSecrets": false
}
```

Para Bitwarden Password Manager via CLI, `supportsScopedAccess` deve ser `false` porque o Harpo aplica escopo lógico, mas o cofre desbloqueado pode ter acesso amplo no contexto do usuário.

---

## 18. Critérios de aceite do MVP

O MVP é considerado funcional quando:

- Um usuário consegue inicializar um projeto com `harpo init`.
- Um provider Bitwarden CLI pode ser configurado.
- Um alias de secret pode ser mapeado sem salvar valor em disco.
- Um profile pode ser criado com TTL.
- `harpo run --profile X -- claude` injeta `GITLAB_TOKEN` no processo filho.
- `BW_SESSION` não é herdado pelo processo filho.
- `harpo exec` funciona para um comando simples.
- `harpo env render` cria `.harpo/.env.session` com confirmação.
- `harpo audit list` mostra eventos sem valores sensíveis.
- `harpo agent setup claude` cria instruções úteis.
- `harpo agent setup codex` cria instruções úteis.
- Testes automatizados cobrem resolução de secret, política, sessão, runner e audit log.

---

## 19. Testes essenciais

### Unitários

- Parse de TTL.
- Validação de `harpo.yml`.
- Validação de profile.
- Bloqueio de `reveal` no modo `strict`.
- Redaction de valores conhecidos.
- Geração de audit log sem segredo.
- Montagem de environment sem `BW_SESSION`.

### Integração

- Mock de provider Bitwarden.
- `harpo run` com processo filho que imprime presença/ausência de variável.
- `harpo exec` com comando simples.
- `harpo env render` criando arquivo e `.gitignore`.

### Segurança

- Garantir que secret não aparece em stdout.
- Garantir que secret não aparece em stderr.
- Garantir que secret não aparece em audit log.
- Garantir que `harpo.yml` não recebe valor secreto.
- Garantir que `BW_SESSION` não é passado ao processo filho.

---

## 20. Roadmap pós-MVP

### v0.2

- Provider Bitwarden Secrets Manager.
- Provider 1Password.
- Melhor redaction.
- Templates de policy por agente.
- `harpo doctor`.

### v0.3

- Provider AWS Secrets Manager.
- Provider HashiCorp Vault.
- Scanner básico de secrets no repositório.
- Suporte a `.env.example` gerado a partir do profile.

### v0.4

- MCP server local.
- TUI de autorização.
- OS keychain para cache temporário.
- Políticas por organização/time.

---

## 21. Posicionamento inicial do projeto open source

Harpo deve ser apresentado como:

```text
A local secret broker for AI coding agents.
```

Descrição curta:

```text
Harpo lets developers grant temporary, scoped access to secrets from their existing vaults without pasting credentials into prompts, committing .env files, or giving AI coding agents direct access to the vault.
```

Descrição em português:

```text
Harpo permite que desenvolvedores autorizem acesso temporário e limitado a secrets do cofre existente sem colar credenciais no prompt, sem commitar .env e sem dar ao agente de IA acesso direto ao vault.
```

---

## 22. Licença recomendada para o MVP

Para maximizar adoção no mercado de software:

- **Apache-2.0**: boa opção para ferramenta de infraestrutura, permite uso comercial e contribuições amplas.
- **MIT**: simples e permissiva, mas com menos linguagem explícita de patente.
- **MPL-2.0**: opção intermediária caso queira garantir abertura de alterações no core sem impedir uso comercial.

Recomendação inicial:

```text
Apache-2.0
```

Motivo: reduz fricção para empresas, permite adoção ampla e é familiar para ferramentas de infraestrutura.

---

## 23. Definição de sucesso do MVP

O MVP será bem-sucedido se permitir este fluxo sem fricção relevante:

```bash
bw unlock
harpo run --profile my-project-dev -- claude
```

E se o usuário conseguir dizer:

> “Agora meus agentes conseguem trabalhar com credenciais de desenvolvimento sem eu colar token no chat, sem espalhar `.env` e sem dar acesso ao meu cofre inteiro.”
