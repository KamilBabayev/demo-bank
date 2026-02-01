package repository

import (
"context"
"errors"
"fmt"
"time"

"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
"user-service/models"
)

var (
ErrUserNotFound      = errors.New("user not found")
ErrUserAlreadyExists = errors.New("user already exists")
ErrInvalidInput      = errors.New("invalid input")
)

type UserRepository struct {
db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, req *models.CreateUserRequest, passwordHash string) (*models.User, error) {
query := `
INSERT INTO users (username, email, password_hash, first_name, last_name, phone, role)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, username, email, first_name, last_name, phone, role, status, 
          created_at, updated_at, last_login_at, failed_login_attempts, locked_until
`

user := &models.User{}
err := r.db.QueryRow(
ctx, query,
req.Username, req.Email, passwordHash, req.FirstName, req.LastName, req.Phone, req.Role,
).Scan(
&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
&user.Phone, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
&user.LastLoginAt, &user.FailedLoginAttempts, &user.LockedUntil,
)

if err != nil {
return nil, fmt.Errorf("failed to create user: %w", err)
}

return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
query := `
SELECT id, username, email, password_hash, first_name, last_name, phone, role, status,
       created_at, updated_at, last_login_at, failed_login_attempts, locked_until
FROM users
WHERE id = $1
`

user := &models.User{}
err := r.db.QueryRow(ctx, query, id).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName,
&user.LastName, &user.Phone, &user.Role, &user.Status, &user.CreatedAt,
&user.UpdatedAt, &user.LastLoginAt, &user.FailedLoginAttempts, &user.LockedUntil,
)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrUserNotFound
}
return nil, fmt.Errorf("failed to get user: %w", err)
}

return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
query := `
SELECT id, username, email, password_hash, first_name, last_name, phone, role, status,
       created_at, updated_at, last_login_at, failed_login_attempts, locked_until
FROM users
WHERE username = $1
`

user := &models.User{}
err := r.db.QueryRow(ctx, query, username).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName,
&user.LastName, &user.Phone, &user.Role, &user.Status, &user.CreatedAt,
&user.UpdatedAt, &user.LastLoginAt, &user.FailedLoginAttempts, &user.LockedUntil,
)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrUserNotFound
}
return nil, fmt.Errorf("failed to get user: %w", err)
}

return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
query := `
SELECT id, username, email, password_hash, first_name, last_name, phone, role, status,
       created_at, updated_at, last_login_at, failed_login_attempts, locked_until
FROM users
WHERE email = $1
`

user := &models.User{}
err := r.db.QueryRow(ctx, query, email).Scan(
&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FirstName,
&user.LastName, &user.Phone, &user.Role, &user.Status, &user.CreatedAt,
&user.UpdatedAt, &user.LastLoginAt, &user.FailedLoginAttempts, &user.LockedUntil,
)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrUserNotFound
}
return nil, fmt.Errorf("failed to get user: %w", err)
}

return user, nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) (*models.UserListResponse, error) {
var total int64
countQuery := `SELECT COUNT(*) FROM users`
if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
return nil, fmt.Errorf("failed to count users: %w", err)
}

query := `
SELECT id, username, email, first_name, last_name, phone, role, status,
       created_at, updated_at, last_login_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`

rows, err := r.db.Query(ctx, query, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to list users: %w", err)
}
defer rows.Close()

users := []models.User{}
for rows.Next() {
var user models.User
err := rows.Scan(
&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
&user.Phone, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
&user.LastLoginAt,
)
if err != nil {
return nil, fmt.Errorf("failed to scan user: %w", err)
}
users = append(users, user)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating users: %w", err)
}

return &models.UserListResponse{
Users: users,
Total: total,
}, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
query := `UPDATE users SET updated_at = NOW()`
args := []interface{}{}
argCount := 1

if req.FirstName != nil {
argCount++
query += fmt.Sprintf(`, first_name = $%d`, argCount)
args = append(args, *req.FirstName)
}

if req.LastName != nil {
argCount++
query += fmt.Sprintf(`, last_name = $%d`, argCount)
args = append(args, *req.LastName)
}

if req.Phone != nil {
argCount++
query += fmt.Sprintf(`, phone = $%d`, argCount)
args = append(args, *req.Phone)
}

if req.Status != nil {
argCount++
query += fmt.Sprintf(`, status = $%d`, argCount)
args = append(args, *req.Status)
}

argCount++
query += fmt.Sprintf(` WHERE id = $%d RETURNING id, username, email, first_name, last_name, phone, role, status, created_at, updated_at, last_login_at`, argCount)
args = append(args, id)

user := &models.User{}
err := r.db.QueryRow(ctx, query, args...).Scan(
&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
&user.Phone, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
&user.LastLoginAt,
)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrUserNotFound
}
return nil, fmt.Errorf("failed to update user: %w", err)
}

return user, nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
query := `UPDATE users SET status = 'closed', updated_at = NOW() WHERE id = $1`

result, err := r.db.Exec(ctx, query, id)
if err != nil {
return fmt.Errorf("failed to delete user: %w", err)
}

if result.RowsAffected() == 0 {
return ErrUserNotFound
}

return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
query := `UPDATE users SET last_login_at = $1, failed_login_attempts = 0, locked_until = NULL WHERE id = $2`

_, err := r.db.Exec(ctx, query, time.Now(), id)
if err != nil {
return fmt.Errorf("failed to update last login: %w", err)
}

return nil
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, username string, maxAttempts int, lockDuration time.Duration) error {
query := `
UPDATE users 
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE 
        WHEN failed_login_attempts + 1 >= $2 THEN $3
        ELSE locked_until
    END
WHERE username = $1
`

_, err := r.db.Exec(ctx, query, username, maxAttempts, time.Now().Add(lockDuration))
if err != nil {
return fmt.Errorf("failed to increment failed login attempts: %w", err)
}

return nil
}
