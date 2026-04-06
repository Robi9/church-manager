package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (s *Service) Register(email, password string) (User, error) {
	if email == "" || password == "" {
		return User{}, errors.New("email and password are required")
	}

	if len(password) < 8 {
		return User{}, errors.New("password must be at least 8 characters")
	}

	exists, err := s.repo.EmailExists(email)
	if err != nil {
		return User{}, err
	}
	if exists {
		return User{}, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		Email:    email,
		Password: string(hash),
	}

	return s.repo.Create(user)
}

func (s *Service) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return generateToken(user.ID, s.jwtSecret)
}
