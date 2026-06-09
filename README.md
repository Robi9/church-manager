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
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=church_manager

JWT_SECRET=your-secret-key
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
NEXT_PUBLIC_API_URL=http://localhost:8080
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

## License

This project is currently under development.
