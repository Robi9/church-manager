# ⛪ Church Manager

Church Manager is a web application designed to help churches manage members, registrations, and administrative tasks.

## Features

* Member registration
* Member editing
* Member import via spreadsheet
* Authentication
* Responsive web interface
* REST API

## Tech Stack

### Backend

* Go
* PostgreSQL
* JWT Authentication

### Frontend

* Next.js
* React
* TypeScript
* Tailwind CSS
* shadcn/ui

## Project Structure

```text
church-manager/
├── cmd/            # Application entrypoints
├── internal/       # Business logic
├── migrations/     # Database migrations
└── web/            # Next.js frontend
```

## Prerequisites

Make sure you have the following installed:

* Go 1.24+
* Node.js 22+
* PostgreSQL

## Environment Configuration

### Backend

Create a `.env` file in the project root:

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/church_manager?sslmode=disable
JWT_SECRET=replace-with-at-least-32-random-characters
ALLOWED_ORIGINS=http://localhost:3000
APP_ENV=development
```

### Database

Create a PostgreSQL database:

```sql
CREATE DATABASE church_manager;
```

Run the SQL migrations located in:

```text
migrations/
```

## Running the Backend

From the project root:

```bash
go mod download
godotenv go run cmd/api/main.go
```

The API will be available at:

```text
http://localhost:8080
```

## Running the Frontend

Navigate to the frontend directory:

```bash
cd web
```

Install dependencies:

```bash
npm install
```

Create a `.env.local` file:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

Start the development server:

```bash
npm run dev
```

The frontend will be available at:

```text
http://localhost:3000
```

## Development Workflow

Start the backend:

```bash
godotenv go run cmd/api/main.go
```

Start the frontend:

```bash
cd web
npm run dev
```

## Current Features

* User authentication
* Member management
* Member import
* Member details page
* Member editing
* Responsive UI

## Future Improvements

* Dashboard
* Attendance tracking
* Ministry management
* Financial management
* Reports and analytics

## Production deployment

The lowest-cost supported setup is:

* Netlify for the Next.js frontend
* Koyeb for the Go API
* Neon for PostgreSQL

The frontend and backend are separate deployments connected to the same Git
repository. The committed `netlify.toml` builds the frontend from `web/`, and
the root `Dockerfile` builds the API.

### 1. Create the Neon database

Create a PostgreSQL project and copy its pooled connection string. It must use
TLS (`sslmode=require`). The API applies the SQL migrations during startup.

### 2. Deploy the API on Koyeb

Create a Web Service from this repository and select the root `Dockerfile`.
Configure the health check path as `/health`, port `8080`, and add:

```env
APP_ENV=production
GIN_MODE=release
DATABASE_URL=postgresql://...neon.tech/...?sslmode=require
JWT_SECRET=<at-least-32-random-characters>
ALLOWED_ORIGINS=https://your-site.netlify.app
ALLOW_REGISTRATION=false
INITIAL_ADMIN_EMAIL=admin@example.com
INITIAL_ADMIN_PASSWORD=<at-least-12-characters>
```

Generate `JWT_SECRET` locally with `openssl rand -base64 48`. Never commit the
real values above.

The initial admin is created only when the users table is empty. After the
first successful login, remove `INITIAL_ADMIN_EMAIL` and
`INITIAL_ADMIN_PASSWORD` from Koyeb and redeploy. Public registration stays
disabled unless `ALLOW_REGISTRATION=true` is explicitly configured.

Verify the API before deploying the frontend:

```bash
curl https://your-api.koyeb.app/health
curl https://your-api.koyeb.app/ready
```

Both endpoints should return `{"status":"ok"}`.

### 3. Deploy the frontend on Netlify

Import the same repository. Netlify reads `netlify.toml`. Add this environment
variable before the first production build:

```env
NEXT_PUBLIC_API_URL=https://your-api.koyeb.app/api
```

After Netlify provides the final site URL, make sure that exact origin appears
in the API's `ALLOWED_ORIGINS` value and redeploy the API. Multiple origins are
comma-separated, for example:

```env
ALLOWED_ORIGINS=https://your-site.netlify.app,https://sistema.example.org
```

### 4. Production checks

1. Open the frontend and sign in with the initial admin.
2. Remove the two `INITIAL_ADMIN_*` secrets and redeploy the API.
3. Confirm that `POST /api/auth/register` returns 404.
4. Test member creation, editing, deactivation, CSV import and export.
5. Export an encrypted PostgreSQL backup and test restoring it before storing
   real member data.

The free tiers can sleep or pause at their usage limits and do not provide a
production SLA. Upgrade the database first when reliable backups are needed,
then the API when cold starts become disruptive. Because the app uses standard
PostgreSQL, Docker and environment variables, each component can be moved or
upgraded independently.

## License

This project is currently under development.
