package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const dbTimeout = time.Second * 3

type PostgresRepository struct {
	Conn *sql.DB
}

func NewPostgresRepository(pool *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		Conn: pool,
	}
}

// User is the structure which holds one user from the database.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `json:"-"`
	Active    int       `json:"active,omitempty"`
	Score     int       `json:"score,omitempty"`
	Referrer  string    `json:"referrer,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *PostgresRepository) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	return u.Conn.ExecContext(ctx, query, args...)
}

func (u *PostgresRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	return u.Conn.QueryRowContext(ctx, query, args...)
}

// UserExists проверяет, существует ли пользователь с указанным id
func (u *PostgresRepository) UserExists(id int) (bool, error) {
	var exists bool
	err := u.queryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		log.Println("failed to check if user exists: ", err)
		return false, err
	}
	return exists, nil
}

// AddPoints adds  some points
func (u *PostgresRepository) AddPoints(id, point int) error {
	idExists, err := u.UserExists(id)
	if err != nil {
		return err
	}

	if !idExists {
		log.Println("User does not exist")
		return errors.New("user does not exist")
	}
	stmt := `update users set score = score + $1, updated_at = $2 where id = $3`
	_, err = u.execQuery(context.Background(), stmt, point, time.Now(), id)
	if err != nil {
		log.Printf("Error adding points to user %d: %v", id, err)
		return fmt.Errorf("failed to add points: %w", err)
	}
	return nil
}

// GetAll returns a slice of all users, sorted by last name
func (u *PostgresRepository) GetAll() ([]*User, error) {
	query := `select id, email, first_name, last_name, active, score, created_at, updated_at, referrer
              from users order by score desc`

	rows, err := u.Conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.Score,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Referrer,
		)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// EmailCheck using to auth, gets password by provided email
func (u *PostgresRepository) EmailCheck(email string) (*User, error) {
	var emailExists bool
	err := u.queryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&emailExists)
	if err != nil {
		log.Println("failed to check email: ", err)
		return nil, err
	}

	if !emailExists {
		log.Println("User with that email does not exists: ", err)
		return nil, err
	}

	query := `select first_name, password from users where email = $1`

	var user User
	err = u.queryRow(context.Background(), query, email).Scan(
		&user.FirstName,
		&user.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user's password by email: %w", err)
	}
	return &user, nil
}

// GetByEmail returns info of one user by email
func (u *PostgresRepository) GetByEmail(email string) (*User, error) {
	query := `select id, email, first_name, last_name, password, active, score, created_at, updated_at 
              from users where email = $1`

	var user User
	err := u.queryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Active,
		&user.Score,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	return &user, nil
}

// RedeemReferrer redeems the referrer with provided id and referrer, adds points to both users
func (u *PostgresRepository) RedeemReferrer(id int, referrer string) error {
	var referrerExists, idExists bool
	var sameCheck string
	err := u.queryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE referrer = $1)", referrer).Scan(&referrerExists)
	if err != nil {
		log.Println("failed to check referrer: ", err)
		return err
	}

	if !referrerExists {
		log.Println("Referrer does not exists")
		return err
	}

	idExists, err = u.UserExists(id)
	if err != nil {
		return err
	}

	if !idExists {
		log.Println("User does not exist")
		return err
	}

	err = u.queryRow(context.Background(), "SELECT referrer FROM users WHERE id = $1", id).Scan(&sameCheck)
	if err != nil {
		log.Println("failed to get user's referrer: ", err)
		return err
	}

	if sameCheck == referrer {
		log.Println("User cannot redeem for their own referrer")
		return err
	}

	_, err = u.execQuery(context.Background(), "UPDATE users SET score = score + 100 WHERE referrer = $1", referrer)
	if err != nil {
		log.Println("failed to update referrer's score: ", err)
		return err
	}

	_, err = u.execQuery(context.Background(), "UPDATE users SET score = score + 25 WHERE id = $1", id)
	if err != nil {
		log.Println("failed to update score for who redeemed referrer: ", err)
		return err
	}

	return nil
}

// GetOne returns one user by id
func (u *PostgresRepository) GetOne(id int) (*User, error) {
	idExists, err := u.UserExists(id)
	if err != nil {
		return nil, err
	}

	if !idExists {
		log.Println("User does not exist")
		return nil, errors.New("user does not exist")
	}
	query := `select id, email, first_name, last_name, active, score, created_at, updated_at, referrer 
              from users where id = $1`

	var user User
	err = u.queryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.Score,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Referrer,
	)
	if err != nil {
		log.Println("failed to fetch user by id: ", err)
		return nil, err
	}

	return &user, nil
}

// Update updates one user in the database, using the information stored in the receiver u
func (u *PostgresRepository) Update(user User) error {
	idExists, err := u.UserExists(user.ID)
	if err != nil {
		return err
	}

	if !idExists {
		log.Println("User does not exist")
		return errors.New("user does not exist")
	}
	stmt := `update users set
             email = $1,
             first_name = $2,
             last_name = $3,
             active = $4,
             updated_at = $5
             where id = $6`

	_, err = u.execQuery(context.Background(), stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Active,
		time.Now(),
		user.ID,
	)
	if err != nil {
		log.Println("failed to update user: ", err)
		return err
	}

	return nil
}

// UpdateScore provides whole new score to the user
func (u *PostgresRepository) UpdateScore(user User) error {
	idExists, err := u.UserExists(user.ID)
	if err != nil {
		return err
	}

	if !idExists {
		log.Println("User does not exist")
		return errors.New("user does not exist")
	}
	stmt := `update users set
             score = $1,
             updated_at = $2
             where id = $3`

	_, err = u.execQuery(context.Background(), stmt,
		user.Score,
		time.Now(),
		user.ID,
	)
	if err != nil {
		log.Println("failed to update user's score: ", err)
		return err
	}

	return nil
}

// DeleteByID deletes one user from the database, by ID
func (u *PostgresRepository) DeleteByID(id int) error {
	idExists, err := u.UserExists(id)
	if err != nil {
		return err
	}

	if !idExists {
		log.Println("User does not exist")
		return errors.New("user does not exist")
	}
	stmt := `delete from users where id = $1`

	_, err = u.execQuery(context.Background(), stmt, id)
	if err != nil {
		log.Println("failed to delete user by id: ", err)
		return err
	}

	return nil
}

func (u *PostgresRepository) Insert(user User) (int, error) {
	if len(user.Password) < 8 {
		return 0, errors.New("password must be at least 8 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	var newID int
	stmt := `insert into users (email, first_name, last_name, password, active, score, created_at, updated_at, referrer)
             values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err = u.queryRow(context.Background(), stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		hashedPassword,
		user.Active,
		user.Score,
		time.Now(),
		time.Now(),
		user.Referrer,
	).Scan(&newID)
	if err != nil {
		log.Println("failed to insert new user: ", err)
		return 0, err
	}

	return newID, nil
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *PostgresRepository) PasswordMatches(plainText string, user User) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, fmt.Errorf("failed to compare passwords: %w", err)
		}
	}
	return true, nil
}
