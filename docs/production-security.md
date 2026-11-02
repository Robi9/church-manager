# Segurança antes dos dados reais

## Publicação desta mudança

1. Publique primeiro a API. A migração `000008` cria sessões e o bloqueio de conta.
   A API aplica as migrações na inicialização. Tokens JWT anteriores deixam de funcionar.
2. No frontend, configure **API_URL** no ambiente do servidor com a URL HTTPS da API,
   terminando em `/api`. A configuração anterior `NEXT_PUBLIC_API_URL` continua sendo
   aceita como transição. O navegador passa a chamar apenas `/api` do próprio frontend.
3. Publique o frontend e entre novamente. Confira login, busca, edição, importação,
   download de erros e logout pelo domínio HTTPS definitivo. Deploys de preview
   devem usar uma API e banco de teste, sem dados reais.
4. A sessão tem duração absoluta de duas horas. Sair revoga **todas as sessões da
   conta no banco**, e só informa sucesso após confirmação da API. Nenhum token
   novo fica acessível ao JavaScript. O cookie é HttpOnly, Secure e SameSite=Lax
   em produção, sem Domain. Mutações exigem Origin igual à origem do frontend.
5. O backend exige `APP_ENV=production`, um `JWT_SECRET` aleatório de pelo menos
   32 caracteres, origens HTTPS explícitas, cadastro público desativado e TLS no
   PostgreSQL. `JWT_SECRET` agora protege os hashes das sessões: sua rotação
   invalida todas elas. Use `openssl rand -base64 48` para gerar um novo segredo.
6. Remova `INITIAL_ADMIN_EMAIL` e `INITIAL_ADMIN_PASSWORD` da hospedagem depois
   da criação da conta inicial. A senha inicial deve ser exclusiva e forte.

O nome da variável JWT_SECRET foi mantido para compatibilidade de implantação;
JWTs não são mais usados. Não há mudança nos perfis de acesso aos membros.

## Bloqueio emergencial da conta

Em uma conexão administrativa protegida, use parâmetros para selecionar a conta:

```sql
BEGIN;
UPDATE users SET disabled_at = NOW() WHERE id = 1;
DELETE FROM sessions WHERE user_id = 1;
COMMIT;
```

Para reativar após investigar o incidente, remova `disabled_at`. A pessoa precisa
entrar novamente. Alterações em `users.updated_at` também invalidam sessões
anteriores; ao alterar uma senha, atualize esse campo e remova as sessões.
Não há tela de administração de contas nesta mudança.

## Importações e registros

Os resultados temporários de importação pertencem ao usuário que enviou o CSV.
Expiram em 30 minutos e são removidos da memória na limpeza a cada minuto.
O limite é de 20 relatórios por processo; atingir o limite remove o mais antigo.
Reiniciar a API também remove os relatórios; use uma instância enquanto o
armazenamento for local em memória. Os membros já importados continuam no banco.

A API não registra URLs de busca automaticamente e não devolve erros internos
ou do banco. Os registros de falha contêm método, padrão da rota e tipo do erro,
sem seu conteúdo. Confira também os logs e ferramentas de observabilidade da
hospedagem: não capture corpos de requisição/resposta, cookies, Authorization
ou parâmetros de busca com nomes e telefones. Não copie dados reais para testes.

## Backup criptografado e recuperação

O repositório fornece `scripts/backup-database.sh`. Ele **não agenda backups**
nem altera serviços de hospedagem. Requer `pg_dump` compatível com a versão do
servidor e [age](https://age-encryption.org/). Configure um serviço libpq fora
do repositório e proteja o arquivo de credenciais com permissão `0600`:

```ini
[church_backup]
host=SEU_HOST
port=5432
dbname=SEU_BANCO
user=USUARIO_DE_BACKUP
password=SENHA
sslmode=verify-full
```

Use um usuário de backup que possa ler as tabelas necessárias, sem permissão de
alteração. Use credenciais diferentes na API e nas migrações quando a hospedagem
permitir; atualmente a API aplica migrações e exige permissões de alteração de schema.

Gere uma chave age em uma máquina protegida com `age-keygen -o backup-key.txt`.
Guarde a chave privada separada dos backups, fora do repositório e da hospedagem
da aplicação. Configure **somente a chave pública** no job de backup:

```bash
PGSERVICE=church_backup \
PGSERVICEFILE=/caminho/protegido/pg_service.conf \
BACKUP_RECIPIENT=age1CHAVE_PUBLICA \
BACKUP_DIR=/caminho/protegido/backups \
./scripts/backup-database.sh
```

O dump passa diretamente para a criptografia, sem arquivo intermediário aberto.
Sessões ativas são excluídas do dump para evitar sua reativação após restauração.
Agende a execução diária em um ambiente protegido e envie os arquivos criptografados
para armazenamento separado, com acesso restrito, retenção definida e alerta de falha.
O intervalo deve corresponder à perda de dados aceitável. Confira também as opções
de backup e recuperação do plano efetivamente contratado no provedor do banco.

**Antes de inserir dados reais**, restaure um backup em um banco vazio e isolado
com a mesma versão principal do PostgreSQL, sem conexões da aplicação. Use um
serviço libpq `church_restore_test` apontando exclusivamente para esse banco:

```bash
set -o pipefail
age --decrypt --identity /caminho/protegido/backup-key.txt /backups/ARQUIVO.age \
  | pg_restore --dbname=service=church_restore_test --exit-on-error --single-transaction --no-owner --no-acl
```

Verifique contagens e registros conhecidos, constraints e ausência de sessões.
Apague o banco de restauração após o teste. Um arquivo criado com sucesso, sozinho,
não comprova que a recuperação funciona. Repita o teste periodicamente.

## Conferências na hospedagem

- Ative MFA nas contas do repositório, frontend, API e banco; evite contas compartilhadas.
- Confirme HTTPS, TLS do banco, proteção dos segredos, acesso administrativo mínimo
  e ausência de credenciais de produção em previews. Não coloque segredos em `NEXT_PUBLIC_*`.
- Confira se logs de acesso do provedor preservam nomes em query strings; restrinja
  acesso e retenção e configure a remoção desses parâmetros quando disponível.
- Teste a API sem autenticação: membros e dashboard devem retornar 401; cadastro público, 404.
- Após logout, reutilizar o token anterior na API deve retornar 401.
- Confirme falhas de login, erros e alertas sem registrar dados pessoais.

Estas configurações externas precisam ser aplicadas e verificadas nas contas de
hospedagem. O código local não comprova MFA, backups agendados ou a configuração
atual dos serviços em produção.

## Verificação automatizada

O CI testa sessões contra PostgreSQL descartável, executa os testes de proteção
contra CSRF e verifica vulnerabilidades com `govulncheck` e `npm audit`.
Localmente: `go test ./...`, `go vet ./...` e, em `web/`,
`npm run test:security`, `npm run lint` e `npm run build`.
O teste de integração também exige `TEST_DATABASE_URL` apontando para um banco
**local** migrado cujo nome termine em `_security_test`.

A atualização de segurança inclui Go 1.26.10 e Next.js 16.4.4. Os utilitários
CSS usados dos componentes shadcn foram preservados localmente com sua licença;
a ferramenta de geração deixou de ser uma dependência da aplicação.

### Resultado da validação local em 02/11/2026

- Migrações aplicadas em PostgreSQL descartável; migração de sessões também
  revertida e reaplicada antes do teste.
- Login, logout individual e global, expiração e bloqueio de conta testados
  contra PostgreSQL real, sem usar dados de produção.
- Backup age criado pelo script e restaurado em outro banco vazio; registro
  fictício recuperado e tabela de sessões vazia, como esperado.
- Testes de cookie HttpOnly/Secure, proteção de origem/CSRF, rotas permitidas
  e limite de upload aprovados; lint, TypeScript e build de produção aprovados.
- `npm audit --omit=dev`: zero vulnerabilidades. A auditoria incluindo ferramentas
  de desenvolvimento ainda informa 16 alertas moderados relacionados a minimatch
  na cadeia do ESLint. Não foi forçada uma troca incompatível dessas ferramentas.
- `govulncheck`: zero vulnerabilidades alcançáveis e zero nos pacotes importados.
  O módulo x/crypto ainda tem um aviso em código não importado pela aplicação
  (GO-2026-5932), sem correção indicada na consulta. Isso não equivale a uma
  garantia contra falhas desconhecidas.

Nenhuma configuração dos serviços de produção foi alterada nesta execução.
