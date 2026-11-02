# Security before adding real data

## Deploying this change

1. Deploy the API first. Migration `000008` creates sessions and adds account
   disabling. The API applies migrations at startup. Previous JWT tokens are no
   longer accepted.
2. In the frontend server environment, set **API_URL** to the API's HTTPS URL,
   ending in `/api`. The previous `NEXT_PUBLIC_API_URL` setting remains supported
   during the transition. The browser now calls only the frontend's own `/api`.
   For origin validation behind reverse proxies, set **APP_ORIGIN** to the public
   frontend origin (HTTPS scheme and hostname, without `/api`). On Netlify, the
   server-provided `URL` variable is used automatically when APP_ORIGIN is unset.
   Configure APP_ORIGIN explicitly when using a different public domain. These
   variables must be available to server functions in the production context.
   Neither request-supplied forwarding headers nor the incoming Origin header
   determine the trusted public origin.
3. Deploy the frontend and sign in again. Check login, search, editing, imports,
   error report downloads, and logout on the final HTTPS domain. Preview deploys
   must use a test API and database without real data.
4. Sessions have an absolute lifetime of two hours. Signing out revokes **all
   sessions for the account in the database** and reports success only after API
   confirmation. New tokens are never exposed to JavaScript. In production, the
   cookie uses HttpOnly, Secure, and SameSite=Lax, without a Domain attribute.
   Mutations require an Origin matching the frontend's origin.
5. The backend requires `APP_ENV=production`, a random `JWT_SECRET` of at least
   32 characters, explicit HTTPS origins, disabled public registration, and TLS
   for PostgreSQL. `JWT_SECRET` now protects session hashes: rotating it
   invalidates all sessions. Use `openssl rand -base64 48` to generate a new secret.
6. Remove `INITIAL_ADMIN_EMAIL` and `INITIAL_ADMIN_PASSWORD` from the hosting
   environment after creating the initial account. Use a strong, unique initial
   password.

The JWT_SECRET variable name was retained for deployment compatibility; JWTs are
no longer used. Member access permissions are unchanged.

## Emergency account disabling

In a protected administrative connection, target the intended account. The
following example uses account ID `1`; replace it with the actual account ID:

```sql
BEGIN;
UPDATE users SET disabled_at = NOW() WHERE id = 1;
DELETE FROM sessions WHERE user_id = 1;
COMMIT;
```

To reactivate the account after investigating the incident, set `disabled_at` to
NULL. The user must sign in again. Changes to `users.updated_at` also invalidate
previous sessions; when changing a password, update this field and delete the
account's sessions. This change does not include an account administration screen.

## Imports and logs

Temporary import results belong to the user who uploaded the CSV. They expire
after 30 minutes and are removed from memory by cleanup running every minute.
Each process retains at most 20 reports; reaching the limit evicts the oldest.
Restarting the API also removes reports. Use one instance while storage remains
local to the process. Members already imported remain in the database.

The API does not automatically log search URLs or return internal or database
errors. Failure logs include the method, route pattern, and error type, without
the error contents. Also check the hosting provider's logs and observability
tools: do not capture request or response bodies, cookies, Authorization headers,
or search parameters containing names and phone numbers. Do not copy real data
into test environments.

## Encrypted backups and recovery

The repository provides `scripts/backup-database.sh`. It **does not schedule
backups** or change hosting services. It requires `pg_dump` compatible with the
server version and [age](https://age-encryption.org/). Configure a libpq service
outside the repository and protect the credentials file with mode `0600`:

```ini
[church_backup]
host=YOUR_HOST
port=5432
dbname=YOUR_DATABASE
user=BACKUP_USER
password=PASSWORD
sslmode=verify-full
```

Use a backup user that can read the required tables without modifying them. Use
separate API and migration credentials when the hosting setup permits it;
currently, the API applies migrations and requires schema modification privileges.

Generate an age key on a protected machine with `age-keygen -o backup-key.txt`.
Keep the private key separate from backups, outside the repository and the
application's hosting environment. Configure **only the public recipient** in
the backup job:

```bash
PGSERVICE=church_backup \
PGSERVICEFILE=/protected/path/pg_service.conf \
BACKUP_RECIPIENT=age1PUBLIC_RECIPIENT \
BACKUP_DIR=/protected/path/backups \
./scripts/backup-database.sh
```

The dump streams directly into encryption without an intermediate plaintext
file. Active sessions are excluded to prevent their reactivation after a restore.
Schedule daily backups in a protected environment and upload the encrypted files
to separate storage with restricted access, a defined retention period, and
failure alerts. Choose a backup interval that matches the acceptable data loss.
Also check the backup and recovery options of the database provider's actual plan.

**Before adding real data**, restore a backup into an empty, isolated database
using the same PostgreSQL major version, without application connections. Point
a libpq service named `church_restore_test` exclusively at that database:

```bash
set -o pipefail
age --decrypt --identity /protected/path/backup-key.txt /backups/FILE.age \
  | pg_restore --dbname=service=church_restore_test --exit-on-error --single-transaction --no-owner --no-acl
```

Verify row counts, known records, constraints, and the absence of sessions.
Delete the restore database after testing. Successfully creating a backup file
alone does not prove that recovery works. Repeat the restore test periodically.

## Hosting checks

- Enable MFA for repository, frontend, API, and database accounts; avoid shared
  accounts.
- Verify HTTPS, database TLS, secret protection, minimal administrative access,
  and the absence of production credentials in previews. Never put secrets in
  `NEXT_PUBLIC_*` variables.
- Check whether provider access logs retain names in query strings. Restrict
  access and retention, and configure parameter removal where available.
- Test the API without authentication: member and dashboard endpoints must
  return 401; public registration must return 404.
- After logout, reusing the previous token against the API must return 401.
- Verify login failures, errors, and alerts without logging personal data.

These external settings must be applied and verified in the hosting accounts.
Local code does not verify MFA, scheduled backups, or the current configuration
of production services.

## Automated verification

CI tests sessions against disposable PostgreSQL, runs CSRF protection tests, and
checks vulnerabilities with `govulncheck` and `npm audit`. Locally, run
`go test ./...`, `go vet ./...`, and, from `web/`, `npm run test:security`,
`npm run lint`, and `npm run build`. Integration tests also require
`TEST_DATABASE_URL` pointing to a migrated **local** database whose name ends in
`_security_test`.

The security update includes Go 1.26.10 and Next.js 16.4.4. The shadcn component
CSS utilities were preserved locally with their license; the generation tool is
no longer an application dependency.

### Local validation results on November 2, 2026

- Migrations were applied to disposable PostgreSQL; the session migration was
  also rolled back and reapplied before testing.
- Login, individual and global logout, expiration, and account disabling were
  tested against real PostgreSQL without using production data.
- An age backup was created by the script and restored into another empty
  database; the fictitious record was recovered and the sessions table was empty,
  as expected.
- HttpOnly/Secure cookie, origin/CSRF protection, allowed route, and upload limit
  tests passed; lint, TypeScript checks, and the production build passed.
- `npm audit --omit=dev`: zero vulnerabilities. The audit including development
  tools still reported 16 moderate alerts related to minimatch in the ESLint
  dependency chain. An incompatible replacement of these tools was not forced.
- `govulncheck`: zero reachable vulnerabilities and zero in imported packages.
  The x/crypto module still had an advisory affecting code not imported by the
  application (GO-2026-5932), with no fix listed at the time of the check. This
  does not guarantee protection against unknown vulnerabilities.

No production service settings were changed during this validation.
