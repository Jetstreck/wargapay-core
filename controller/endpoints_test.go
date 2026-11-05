package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"wargapay-core/controller"
	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/routes"
	"wargapay-core/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Mock Repositories
type mockUserRepo struct{}

func (m *mockUserRepo) Create(u *entity.User) error { u.ID = 2; return nil }
func (m *mockUserRepo) FindByEmail(e string) (*entity.User, error) {
	return &entity.User{ID: 2, Name: "User Test", Email: e, Role: "user"}, nil
}
func (m *mockUserRepo) FindByID(id uint) (*entity.User, error) {
	role := "user"
	if id == 1 {
		role = "admin"
	}
	return &entity.User{ID: id, Name: "User Test", Email: "user@test.com", Role: role}, nil
}
func (m *mockUserRepo) FindAll() ([]entity.User, error) { return nil, nil }

type mockEmailSvc struct{}

func (m *mockEmailSvc) SendNotificationEmail(to, subject, body string) error { return nil }

type mockWasteRepo struct{}

func (m *mockWasteRepo) Create(item *entity.WasteService) error { item.ID = 1; return nil }
func (m *mockWasteRepo) FindByID(id uint) (*entity.WasteService, error) {
	return &entity.WasteService{ID: id, Name: "Sampah Reguler", Price: 35000, IsActive: true}, nil
}
func (m *mockWasteRepo) FindAll(f map[string]interface{}) ([]entity.WasteService, error) {
	return []entity.WasteService{{ID: 1, Name: "Sampah Reguler", Price: 35000, IsActive: true}}, nil
}
func (m *mockWasteRepo) Update(item *entity.WasteService) error { return nil }
func (m *mockWasteRepo) Delete(id uint) error                  { return nil }

type mockDuesRepo struct{}

func (m *mockDuesRepo) Create(item *entity.Dues) error { item.ID = 1; return nil }
func (m *mockDuesRepo) FindByID(id uint) (*entity.Dues, error) {
	return &entity.Dues{ID: id, UserID: 2, Period: "2026-11", Total: 85000, Status: "unpaid"}, nil
}
func (m *mockDuesRepo) FindAll(u uint, p, s string) ([]entity.Dues, error) {
	return []entity.Dues{{ID: 1, UserID: 2, Period: "2026-11", Total: 85000, Status: "unpaid"}}, nil
}
func (m *mockDuesRepo) Update(item *entity.Dues) error { return nil }
func (m *mockDuesRepo) Delete(id uint) error           { return nil }

type mockTxRepo struct{}

func (m *mockTxRepo) Create(tx *entity.Transaction) error { tx.ID = 1; return nil }
func (m *mockTxRepo) FindByID(id uint) (*entity.Transaction, error) {
	return &entity.Transaction{ID: id, UserID: 2, OrderID: "WP-TEST-1", Amount: 85000, Status: "pending"}, nil
}
func (m *mockTxRepo) FindByOrderID(orderID string) (*entity.Transaction, error) {
	return &entity.Transaction{ID: 1, UserID: 2, OrderID: orderID, Amount: 85000, Status: "pending"}, nil
}
func (m *mockTxRepo) FindAll(userID uint, txType, status string) ([]entity.Transaction, error) {
	return []entity.Transaction{{ID: 1, UserID: 2, OrderID: "WP-TEST-1", Amount: 85000, Status: "pending"}}, nil
}
func (m *mockTxRepo) Update(tx *entity.Transaction) error { return nil }
func (m *mockTxRepo) GetReport(s, e string) (*dto.TransactionReportResponse, error) {
	return &dto.TransactionReportResponse{TotalRevenue: 100000, TotalTransactions: 2, TotalPaid: 1}, nil
}

type mockMidtransSvc struct{}

func (m *mockMidtransSvc) CreateSnapTransaction(o string, a int64, n, e, i string) (string, string, error) {
	return "mock-token-123", "https://mock.midtrans.com/pay", nil
}
func (m *mockMidtransSvc) VerifySignature(o, s, g, sig string) bool { return true }

type mockExpenseRepo struct{}

func (m *mockExpenseRepo) Create(e *entity.Expense) error { e.ID = 1; return nil }
func (m *mockExpenseRepo) FindAll(cat string) ([]entity.Expense, error) {
	return []entity.Expense{{ID: 1, Title: "Honor Satpam", Category: "Keamanan", Amount: 4500000}}, nil
}
func (m *mockExpenseRepo) FindByID(id uint) (*entity.Expense, error) {
	return &entity.Expense{ID: id, Title: "Honor Satpam"}, nil
}
func (m *mockExpenseRepo) Delete(id uint) error          { return nil }
func (m *mockExpenseRepo) GetTotalExpense() (int64, error) { return 4500000, nil }

func generateToken(id uint, email, role string) string {
	claims := jwt.MapClaims{
		"id":    id,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}

func setupTestApp() (*echo.Echo, string, string) {
	os.Setenv("JWT_SECRET", "test_jwt_secret_wargapay")

	e := echo.New()

	emailSvc := &mockEmailSvc{}
	userRepo := &mockUserRepo{}
	userSvc := service.NewUserService(userRepo, emailSvc)
	userCtl := controller.NewUserController(userSvc)

	wasteRepo := &mockWasteRepo{}
	wasteSvc := service.NewWasteServiceService(wasteRepo)
	wasteCtl := controller.NewWasteServiceController(wasteSvc)

	duesRepo := &mockDuesRepo{}
	duesSvc := service.NewDuesService(duesRepo)
	duesCtl := controller.NewDuesController(duesSvc)

	weatherSvc := service.NewWeatherService()
	weatherCtl := controller.NewWeatherController(weatherSvc)

	txRepo := &mockTxRepo{}
	midtransSvc := &mockMidtransSvc{}
	txSvc := service.NewTransactionService(txRepo, userRepo, duesRepo, wasteRepo, midtransSvc, emailSvc)
	txCtl := controller.NewTransactionController(txSvc)

	expenseRepo := &mockExpenseRepo{}
	expenseSvc := service.NewExpenseService(expenseRepo)
	expenseCtl := controller.NewExpenseController(expenseSvc)

	routes.SetupRoutes(e, userCtl, wasteCtl, duesCtl, weatherCtl, txCtl, expenseCtl)

	adminToken := generateToken(1, "admin@wargapay.id", "admin")
	userToken := generateToken(2, "user@wargapay.id", "user")

	return e, adminToken, userToken
}

func TestAllEndpoints(t *testing.T) {
	e, adminToken, userToken := setupTestApp()

	tests := []struct {
		name         string
		method       string
		path         string
		token        string
		body         interface{}
		expectedCode int
	}{
		// 1. Auth Endpoints
		{
			name:         "Register Endpoint",
			method:       http.MethodPost,
			path:         "/api/v1/register",
			body:         dto.RegisterRequest{Name: "Tester", Email: "tester@wargapay.id", Password: "secretpassword"},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Login Endpoint Validation (Wrong credentials)",
			method:       http.MethodPost,
			path:         "/api/v1/login",
			body:         dto.LoginRequest{Email: "notfound@wargapay.id", Password: "wrong"},
			expectedCode: http.StatusUnauthorized,
		},

		// 2. Waste Services Endpoints
		{
			name:         "List Services (User Auth)",
			method:       http.MethodGet,
			path:         "/api/v1/services",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Service By ID",
			method:       http.MethodGet,
			path:         "/api/v1/services/1",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Create Service (Admin Only - Success)",
			method:       http.MethodPost,
			path:         "/api/v1/admin/services",
			token:        adminToken,
			body:         dto.WasteServiceRequest{Name: "Sampah Puing", Price: 75000},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Create Service (User - Forbidden 403)",
			method:       http.MethodPost,
			path:         "/api/v1/admin/services",
			token:        userToken,
			body:         dto.WasteServiceRequest{Name: "Sampah Puing", Price: 75000},
			expectedCode: http.StatusForbidden,
		},

		// 3. Dues (Iuran) Endpoints
		{
			name:         "List Dues (User Auth)",
			method:       http.MethodGet,
			path:         "/api/v1/dues",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Dues By ID",
			method:       http.MethodGet,
			path:         "/api/v1/dues/1",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Create Dues (Admin Only)",
			method:       http.MethodPost,
			path:         "/api/v1/admin/dues",
			token:        adminToken,
			body:         dto.DuesRequest{UserID: 2, Period: "2026-11", WasteFee: 35000, SecurityFee: 30000, RTFund: 20000},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Pay Dues Manual (Admin)",
			method:       http.MethodPost,
			path:         "/api/v1/admin/dues/1/pay",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},

		// 4. Transactions & Midtrans Payment Endpoints
		{
			name:   "Create Transaction for Dues",
			method: http.MethodPost,
			path:   "/api/v1/transactions",
			token:  userToken,
			body: func() dto.CreateTransactionRequest {
				duesID := uint(1)
				return dto.CreateTransactionRequest{Type: "dues", DuesID: &duesID}
			}(),
			expectedCode: http.StatusCreated,
		},
		{
			name:         "List Transactions (User)",
			method:       http.MethodGet,
			path:         "/api/v1/transactions",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Get Transaction By ID",
			method:       http.MethodGet,
			path:         "/api/v1/transactions/1",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Simulate Payment Success (Local Sandbox)",
			method:       http.MethodPost,
			path:         "/api/v1/transactions/1/simulate-paid",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:   "Midtrans Webhook Callback (Public)",
			method: http.MethodPost,
			path:   "/api/v1/payment/webhook",
			body: dto.MidtransNotificationRequest{
				OrderID:           "WP-TEST-1",
				StatusCode:        "200",
				GrossAmount:       "85000",
				TransactionStatus: "settlement",
				SignatureKey:      "dummy",
			},
			expectedCode: http.StatusOK,
		},

		// 5. Admin Dashboard & Reports
		{
			name:         "Admin Dashboard (Admin Token)",
			method:       http.MethodGet,
			path:         "/api/v1/admin/dashboard",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Admin Dashboard (User Token - 403 Forbidden)",
			method:       http.MethodGet,
			path:         "/api/v1/admin/dashboard",
			token:        userToken,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Admin Transaction Financial Report",
			method:       http.MethodGet,
			path:         "/api/v1/admin/transactions/report",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},

		// 6. Weather Endpoint Validation
		{
			name:         "Weather Endpoint Validation (Missing Coordinates - 400 Bad Request)",
			method:       http.MethodGet,
			path:         "/api/v1/weather",
			token:        userToken,
			expectedCode: http.StatusBadRequest,
		},

		// 7. Expenses & User Management
		{
			name:         "Get Expenses List",
			method:       http.MethodGet,
			path:         "/api/v1/expenses",
			token:        userToken,
			expectedCode: http.StatusOK,
		},
		{
			name:   "Admin Create Expense",
			method: http.MethodPost,
			path:   "/api/v1/admin/expenses",
			token:  adminToken,
			body: map[string]interface{}{
				"title":       "Honor Satpam",
				"category":    "Keamanan",
				"amount":      4500000,
				"vendor":      "Pos Satpam",
				"date":        "01 Nov 2026",
				"description": "Gaji satpam",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Admin List Users",
			method:       http.MethodGet,
			path:         "/api/v1/admin/users",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("%s failed: expected status %d, got %d. Body: %s", tt.name, tt.expectedCode, rec.Code, rec.Body.String())
			} else {
				t.Logf("✓ [%d %s] %s -> PASS", rec.Code, tt.method, tt.path)
			}
		})
	}
}
