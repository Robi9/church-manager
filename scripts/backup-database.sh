#!/usr/bin/env bash
set -euo pipefail
umask 077

# Configure a libpq service outside the repository, with a read-only backup role.
# No password or full connection string is put in a command argument or log.
: "${PGSERVICE:?Set PGSERVICE to the libpq service for the source database}"
: "${BACKUP_RECIPIENT:?Set BACKUP_RECIPIENT to an age public recipient}"
: "${BACKUP_DIR:?Set BACKUP_DIR to a protected backup directory}"
command -v pg_dump >/dev/null
command -v age >/dev/null
mkdir -p -- "$BACKUP_DIR"
backup_tmp=$(mktemp "$BACKUP_DIR/.church-backup.XXXXXXXX.age")
trap 'rm -f -- "$backup_tmp"' EXIT
backup_final="$BACKUP_DIR/church-$(date -u +%Y%m%dT%H%M%SZ)-${backup_tmp##*.church-backup.}"
# Stream directly into authenticated encryption; never write a plaintext dump.
# Exclude active sessions so restoring a backup cannot revive stolen sessions.
pg_dump --format=custom --no-owner --no-acl --exclude-table-data=public.sessions \
  | age --recipient "$BACKUP_RECIPIENT" --output "$backup_tmp"
mv -- "$backup_tmp" "$backup_final"
printf 'Encrypted backup created: %s\n' "$backup_final"
