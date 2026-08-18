# Meal Tracker

A production-grade meal tracking application built with Go, HTML/CSS/JavaScript, and MariaDB. Track your daily meals with a clean, minimalist interface. Designed for mutual aid efforts and self-hosted deployment on Coolify.

## Features

- **User Management**: Single-user by default (first user is admin), with multi-user ready architecture
- **Meal Logging**: Log breakfast, lunch, and dinner with time, portion size, and notes
- **Calendar View**: Today's meals, weekly overview, and monthly tracking
- **Admin Panel**: Manage users, configure site settings, and handle SMTP for password resets
- **Flexible Configuration**: All settings managed through web UI, minimal config files
- **Public Signup**: Toggle public sign-ups on/off (disabled by default)
- **Timezone Support**: Per-user timezone configuration
- **Production Ready**: Secure authentication (JWT + bcrypt), error handling, SQL injection prevention

## Tech Stack

- **Backend**: Go (Golang) with Gorilla Mux router
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Database**: MariaDB
- **Deployment**: Docker + Docker Compose (Coolify compatible)

## Quick Start (Local Development)

### Prerequisites
- Go 1.22+
- MariaDB 10.5+
- Node.js (optional, for frontend tooling)

### Setup

1. Clone the repository
```bash
git clone <repo> meal-tracker
cd meal-tracker
```

2. Create `.env` from template
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. Install Go dependencies
```bash
cd backend
go mod download
```

4. Start MariaDB locally (or use existing instance)
```bash
# If using Docker just for DB
docker run -d \
  -e MARIADB_ROOT_PASSWORD=root \
  -e MARIADB_DATABASE=meal_tracker \
  -e MARIADB_USER=meal_tracker \
  -e MARIADB_PASSWORD=password \
  -p 3306:3306 \
  mariadb:latest
```

5. Run the application
```bash
cd backend
go run main.go
```

6. Open http://localhost:8080 in your browser

## Deployment on Coolify

Coolify is a self-hosted PaaS that handles reverse proxying, SSL, and orchestration automatically. This application is optimised for Coolify.

### Prerequisites
- Coolify instance running
- A domain name (e.g., `meals.obulou.org`)
- Git repository access (GitHub, GitLab, Gitea, etc.)

### Deployment Steps

1. **Create a new Coolify application**:
   - Go to Coolify dashboard
   - Click "Add Service" → "Docker Compose"
   - Select your git provider and repository
   - Choose the `meal-tracker` branch

2. **Configure environment variables in Coolify UI**:
   - `DB_ROOT_PASSWORD`: Strong password for MariaDB root
   - `DB_USER`: `meal_tracker`
   - `DB_PASSWORD`: Strong password for app DB user
   - `DB_NAME`: `meal_tracker`
   - `JWT_SECRET`: Generate with `openssl rand -hex 32`
   - `PORT`: `8080` (Coolify will expose this)

3. **Configure domain**:
   - In Coolify, set your domain under "Domains" (e.g., `meals.obulou.org`)
   - Coolify automatically handles SSL with Let's Encrypt

4. **Deploy**:
   - Click "Deploy"
   - Coolify builds the Docker image and starts containers
   - MariaDB initialises automatically
   - App is accessible at your domain

### Coolify Compatibility Notes

This setup follows Coolify best practices:

- **No host port binding**: Uses `expose:` instead of `ports:`, Coolify's reverse proxy handles routing
- **No bind-mounts**: All application files baked into Docker image with `COPY`, no `volumes:` bind-mounts that would shadow files
- **Environment variable substitution**: Uses `$VAR_NAME` syntax, Coolify automatically populates from UI
- **Named volumes only**: Database persistence via `mariadb_data:` volume (managed by Coolify)
- **Health checks**: Built-in for both app and MariaDB containers

### Troubleshooting Coolify Deployment

**Port already allocated**:
If deployment fails with "port is already allocated", Coolify detected a stale container. Run:
```bash
docker ps
docker rm -f <container_id>
# Then redeploy in Coolify UI
```

**Files not found (503/No available servers)**:
Caused by stale bind-mounts from previous deploys. All files are baked into the image, so the container should work immediately. If not:
- Check Coolify logs: "docker logs <container_id>"
- Verify Dockerfile `COPY` commands executed
- Try a full rebuild: delete the container and redeploy

**SMTP issues**:
If password reset emails don't send:
- Check SMTP credentials in admin panel
- Ensure `from_email` is configured
- Check application logs for SMTP errors
- Some providers (Gmail) require app-specific passwords, not your account password

## First-Time Setup

When you first access the application:

1. **First user becomes admin automatically**
2. Create your account with email and password
3. Log in to the dashboard
4. Visit admin panel (Admin Panel link appears if you're admin)
5. Configure site name, logo, timezone, SMTP settings
6. Toggle public sign-ups if you want other users to register

## Admin Panel Features

### User Management
- View all users
- Edit user roles (admin/user)
- Change user timezone
- Reset user password
- Delete users (prevents deletion of last admin)

### Site Settings
- Site name and logo URL
- Default timezone
- Public signup toggle (allows/blocks new registrations)

### Email (SMTP) Settings
- SMTP host, port, username, password
- From email address
- Used for password reset emails

## API Endpoints

### Authentication
- `POST /api/auth/signup` - Create account
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Get current user
- `POST /api/auth/request-reset` - Request password reset
- `POST /api/auth/reset-password` - Confirm password reset

### Meals
- `POST /api/meals` - Create meal entry
- `GET /api/meals/today` - Get today's meals
- `GET /api/meals/range?start_date=2024-01-01&end_date=2024-01-31` - Get meals by date range
- `PUT /api/meals/update?id=1` - Update meal
- `DELETE /api/meals/delete?id=1` - Delete meal

### Admin (requires admin role)
- `GET /api/admin/settings` - Get settings
- `PUT /api/admin/settings` - Update settings
- `GET /api/admin/users` - List users
- `PUT /api/admin/users/role?id=1` - Update user role
- `PUT /api/admin/users/timezone?id=1` - Update user timezone
- `DELETE /api/admin/users?id=1` - Delete user
- `POST /api/admin/users/reset-password?id=1` - Reset user password

## Security Considerations

- Passwords hashed with bcrypt (cost 10)
- JWT tokens expire after 24 hours
- CORS headers allow all origins (safe for self-hosted)
- SQL injection prevention via prepared statements
- Password reset tokens expire after 1 hour
- Admin operations require admin role verification

## Database Schema

### users
- id, email, password, full_name, timezone, is_admin, created_at, updated_at

### meals
- id, user_id, meal_type, date, time, portion, notes, created_at, updated_at
- Indexed on (user_id, date) and (user_id, meal_type)

### settings
- id, site_name, logo_url, smtp_host, smtp_port, smtp_user, smtp_password, from_email, public_signup, default_timezone, updated_at

### password_reset_tokens
- id, user_id, token, expires_at, created_at
- Indexed on token and expires_at

## Future Enhancements

- [ ] User profile update endpoint
- [ ] Current password verification for password changes
- [ ] Email notifications for password resets
- [ ] Bulk meal import/export
- [ ] Meal history analytics and trends
- [ ] Mobile app
- [ ] Recurring meal logging
- [ ] Meal sharing between users
- [ ] Dark mode

## License

This project is provided as-is for mutual aid efforts. Modify and deploy freely.

## Support

For issues or questions:
1. Check the logs: `docker logs <container_id>` on Coolify
2. Verify environment variables are set correctly
3. Ensure MariaDB is healthy: `docker ps` should show both containers running
4. Check the admin panel settings are configured

---

Built with care for self-hosting and digital sovereignty. ✊
