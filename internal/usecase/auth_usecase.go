package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo  domain.UserRepository
	jwtSecret string // Thêm vào đây
	jwtExpire int    // Thêm vào đây
}

func NewAuthUseCase(userRepo domain.UserRepository, jwtSecret string, jwtExpire int) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

// expireHours: Token này sống được bao lâu
func (uc *authUseCase) generateToken(user *models.User) (string, error) {
	claims := dtos.AuthClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(uc.jwtExpire) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *authUseCase) Register(ctx context.Context, req *dtos.RegisterRequest) (*dtos.AuthResponse, error) {
	//1 check tồn tại
	existingUser, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}
	//2 hashpassword
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 3. Tạo User
	role := "customer"
	if req.Role != "" {
		role = req.Role
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	//4 sinh token nagy tại lúc này
	token, err := uc.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &dtos.AuthResponse{Token: token, User: user}, nil
}

func (uc *authUseCase) Login(ctx context.Context, req *dtos.LoginRequest) (*dtos.AuthResponse, error) {
	//1 tìm user
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	//3 sinh token
	token, err := uc.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dtos.AuthResponse{Token: token, User: user}, nil
}
