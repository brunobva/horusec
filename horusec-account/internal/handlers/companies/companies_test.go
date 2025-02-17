// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package companies

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	companyUseCases "github.com/ZupIT/horusec/horusec-account/internal/usecases/company"
	"github.com/ZupIT/horusec/horusec-account/internal/usecases/repositories"

	"github.com/ZupIT/horusec/development-kit/pkg/databases/relational"
	accountEntities "github.com/ZupIT/horusec/development-kit/pkg/entities/account"
	"github.com/ZupIT/horusec/development-kit/pkg/entities/account/dto"
	authEntities "github.com/ZupIT/horusec/development-kit/pkg/entities/auth"
	"github.com/ZupIT/horusec/development-kit/pkg/entities/roles"
	authEnums "github.com/ZupIT/horusec/development-kit/pkg/enums/auth"
	errorsEnum "github.com/ZupIT/horusec/development-kit/pkg/enums/errors"
	"github.com/ZupIT/horusec/development-kit/pkg/services/broker"
	authGrpc "github.com/ZupIT/horusec/development-kit/pkg/services/grpc/auth"
	"github.com/ZupIT/horusec/development-kit/pkg/services/jwt"
	"github.com/ZupIT/horusec/development-kit/pkg/utils/repository/response"
	"github.com/ZupIT/horusec/horusec-account/config/app"
	companiesController "github.com/ZupIT/horusec/horusec-account/internal/controller/companies"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func getTestAuthorizationToken() string {
	account := &authEntities.Account{
		AccountID: uuid.New(),
		Email:     "test@test.com",
		Password:  "test123",
		Username:  "test",
	}
	token, _, _ := jwt.CreateToken(account, nil)
	return token
}

func TestCreateCompany(t *testing.T) {
	t.Run("should return status code 201 when create a company successfully", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{
			Name: "test",
		}

		resp := &response.Response{}
		resp.SetData(company)
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{ConfigAuth: authEntities.ConfigAuth{AuthType: authEnums.Horusec}})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))
		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return status code 200 when create a company successfully with application admin", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}
		appConfig := &app.Config{}

		company := &accountEntities.CompanyApplicationAdmin{
			Name:       "test",
			AdminEmail: "admin-horusec@example.com",
		}

		resp := &response.Response{}
		resp.SetData(company.ToCompany())
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		mockRead.On("SetFilter").Return(&gorm.DB{})
		mockRead.On("Find").Return(response.NewResponse(0, nil, &authEntities.Account{
			AccountID: uuid.New(),
		}))

		handler := NewHandler(mockWrite, mockRead, brokerMock, appConfig)
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()

		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))

		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return status code 401 when user is unauthorized", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{
			Name: "test",
		}

		resp := &response.Response{}
		resp.SetData(company)
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+"123")
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, "test"))
		w := httptest.NewRecorder()

		handler.Create(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return status code 400 when the body is'nt compatible with the entity", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := struct{ Test string }{Test: "test"}

		resp := &response.Response{}
		resp.SetData(company)
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()

		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	//t.Run("should return status code 400 when create a company and email is invalid with application admin", func(t *testing.T) {
	//	mockRead := &relational.MockRead{}
	//	mockWrite := &relational.MockWrite{}
	//	mockTx := &relational.MockWrite{}
	//	brokerMock := &broker.Mock{}
	//	appConfig := &app.Config{}
	//
	//	company := &accountEntities.CompanyApplicationAdmin{
	//		Name:       "test",
	//		AdminEmail: "invalid email",
	//	}
	//
	//	resp := &response.Response{}
	//	resp.SetData(company.ToCompany())
	//	mockTx.On("Create").Return(resp)
	//	mockTx.On("CommitTransaction").Return(&response.Response{})
	//
	//	mockWrite.On("StartTransaction").Return(mockTx)
	//
	//	mockRead.On("SetFilter").Return(&gorm.DB{})
	//	mockRead.On("Find").Return(response.NewResponse(0, nil, &authEntities.Account{
	//		AccountData: uuid.New(),
	//	}))
	//
	//	handler := NewHandler(mockWrite, mockRead, brokerMock, appConfig)
	//	body, _ := json.Marshal(company)
	//	r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))
	//
	//	_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
	//	r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())
	//
	//	w := httptest.NewRecorder()
	//
	//  r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New(), Permissions: []string{}}))
	//	r = r.WithContext(context.WithValue(r.Context(), authEnums.ConfigAuth, authEntities.ConfigAuth{
	//		ApplicationAdminEnable: true,
	//		AuthType:               authEnums.Horusec,
	//	}))
	//
	//	handler.Create(w, r)
	//
	//	assert.Equal(t, http.StatusBadRequest, w.Code)
	//})

	t.Run("should return status code 500 when create a company and return error in Find with application admin", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}
		appConfig := &app.Config{}

		company := &accountEntities.CompanyApplicationAdmin{
			Name:       "test",
			AdminEmail: "email@example.com",
		}

		resp := &response.Response{}
		resp.SetError(errors.New("test"))
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		mockRead.On("SetFilter").Return(&gorm.DB{})
		mockRead.On("Find").Return(response.NewResponse(0, errors.New("unexpected error"), nil))

		handler := NewHandler(mockWrite, mockRead, brokerMock, appConfig)
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()

		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData,
			&authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))
		handler.Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return status code 400 when ldap group error", func(t *testing.T) {
		controllerMock := &companiesController.Mock{}

		company := &accountEntities.Company{
			Name: "test",
		}

		controllerMock.On("Create").Return(company, errorsEnum.ErrorInvalidLdapGroup)

		handler := Handler{
			companyController: controllerMock,
			companyUseCases:   companyUseCases.NewCompanyUseCases(),
			appConfig:         &app.Config{ConfigAuth: authEntities.ConfigAuth{AuthType: authEnums.Ldap}},
		}

		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()

		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData,
			&authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))

		handler.Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status code 500 when the body is'nt compatible with the entity", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		mockTx := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{
			Name: "test",
		}

		resp := &response.Response{}
		resp.SetError(errors.New("test"))
		mockTx.On("Create").Return(resp)
		mockTx.On("CommitTransaction").Return(&response.Response{})

		mockWrite.On("StartTransaction").Return(mockTx)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()

		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))
		handler.Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return status code 201 when create a company successfully with app admin", func(t *testing.T) {
		appConfig := &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}
		body, _ := json.Marshal(&accountEntities.CompanyApplicationAdmin{Name: "test", AdminEmail: "test@test.com"})

		controllerMock := &companiesController.Mock{}
		controllerMock.On("Create").Return(&accountEntities.Company{}, nil)
		controllerMock.On("GetAccountIDByEmail").Return(uuid.New(), nil)

		handler := Handler{
			companyController:  controllerMock,
			repositoryUseCases: repositories.NewRepositoryUseCases(),
			companyUseCases:    companyUseCases.NewCompanyUseCases(),
			appConfig:          &app.Config{ConfigAuth: authEntities.ConfigAuth{ApplicationAdminEnable: true}},
		}

		r, _ := http.NewRequest(http.MethodPost, "api/companies", bytes.NewReader(body))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		w := httptest.NewRecorder()
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, appConfig))

		handler.Create(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestUpdateCompany(t *testing.T) {
	t.Run("should return status code 200 when update a company successfully", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{Name: "test"}
		resp := &response.Response{}
		resp.SetData(company)
		mockWrite.On("Update").Return(resp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPatch, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		appConfig := &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, appConfig))

		handler.Update(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return status code 400 when retrieve a company fails", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{Name: "test"}
		resp := &response.Response{}
		resp.SetError(errors.New("test"))
		mockWrite.On("Update").Return(resp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(company)
		r, _ := http.NewRequest(http.MethodPatch, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		appConfig := &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, appConfig))

		handler.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status code 400 when invalid request body", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodPatch, "api/companies/123", bytes.NewReader([]byte("")))
		w := httptest.NewRecorder()

		handler.Update(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetCompany(t *testing.T) {
	t.Run("should return status code 200 when retrieve a company successfully", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		company := &accountEntities.Company{Name: "test"}
		resp := &response.Response{}
		resp.SetData(company)
		mockRead.On("Find").Return(resp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", nil)
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))
		w := httptest.NewRecorder()

		handler.Get(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return status code 400 when retrieve a company fails", func(t *testing.T) {
		mockRead := &relational.MockRead{}
		mockWrite := &relational.MockWrite{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		resp.SetError(errors.New("test"))
		mockRead.On("Find").Return(resp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", nil)
		w := httptest.NewRecorder()
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}}))

		handler.Get(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestList(t *testing.T) {
	t.Run("should return status 200 when successfully retrieve companies of an account", func(t *testing.T) {
		controllerMock := &companiesController.Mock{}

		controllerMock.On("List").Return(&[]accountEntities.CompanyResponse{}, nil)

		handler := Handler{
			companyController: controllerMock,
		}

		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", nil)
		w := httptest.NewRecorder()

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		handler.List(w, r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}})))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return status 401 the account is not authorizes", func(t *testing.T) {
		controllerMock := &companiesController.Mock{}

		handler := Handler{
			companyController: controllerMock,
		}

		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", nil)
		w := httptest.NewRecorder()
		r = r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: "", Permissions: []string{}}))

		handler.List(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return status 400 when an error occurred", func(t *testing.T) {
		controllerMock := &companiesController.Mock{}

		controllerMock.On("List").Return(&[]accountEntities.CompanyResponse{}, errors.New("test"))

		handler := Handler{
			companyController: controllerMock,
		}

		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", nil)
		w := httptest.NewRecorder()

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		handler.List(w, r.WithContext(context.WithValue(r.Context(), authEnums.AccountData, &authGrpc.GetAccountDataResponse{AccountID: uuid.New().String(), Permissions: []string{}})))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateAccountCompany(t *testing.T) {
	t.Run("should return status 200 when role update successfully", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		accountCompany := &roles.AccountCompany{Role: "admin", AccountID: uuid.New(), CompanyID: uuid.New()}
		companiesResp := &response.Response{}
		mockWrite.On("Update").Return(companiesResp.SetData(accountCompany))
		mockRead.On("Find").Return(companiesResp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(accountCompany)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		handler.UpdateAccountCompany(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return status 400 when role has an invalid value", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		accountCompany := &roles.AccountCompany{Role: "test", AccountID: uuid.New(), CompanyID: uuid.New()}
		companiesResp := &response.Response{}
		mockWrite.On("Update").Return(companiesResp.SetData(accountCompany))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(accountCompany)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		handler.UpdateAccountCompany(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status 400 when an error occurred", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		accountCompany := &roles.AccountCompany{Role: "admin", AccountID: uuid.New(), CompanyID: uuid.New()}
		companiesResp := &response.Response{}
		mockWrite.On("Update").Return(companiesResp.SetError(errors.New("test")))
		mockRead.On("Find").Return(companiesResp)
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})
		body, _ := json.Marshal(accountCompany)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		_ = os.Setenv("HORUSEC_JWT_SECRET_KEY", "testscret123")
		r.Header.Add("X-Horusec-Authorization", "Bearer "+getTestAuthorizationToken())

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", "85d08ec1-7786-4c2d-bf4e-5fee3a010315")
		ctx.URLParams.Add("accountID", "85d08ec1-7786-4c2d-bf4e-5fee3a010315")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.UpdateAccountCompany(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status 400 when invalid or missing company id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		accountCompany := &roles.AccountCompany{Role: "admin", AccountID: uuid.New(), CompanyID: uuid.New()}
		body, _ := json.Marshal(accountCompany)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateAccountCompany(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status 400 when invalid or missing account id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		accountCompany := &roles.AccountCompany{Role: "admin", AccountID: uuid.New(), CompanyID: uuid.New()}
		body, _ := json.Marshal(accountCompany)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/123", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", "85d08ec1-7786-4c2d-bf4e-5fee3a010315")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.UpdateAccountCompany(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestInviteUser(t *testing.T) {
	inviteUser := &dto.InviteUser{
		Role:  "admin",
		Email: "test@test.com",
	}

	company := &accountEntities.Company{
		CompanyID: uuid.New(),
		Name:      "test",
	}

	account := &authEntities.Account{
		AccountID: uuid.New(),
		Email:     "test@test.com",
		Username:  "test",
	}

	t.Run("should return status 204 when successfully added user", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		respCompany := &response.Response{}
		respAccount := &response.Response{}
		mockRead.On("Find").Once().Return(respAccount.SetData(account))
		mockRead.On("Find").Return(respCompany.SetData(company))
		mockRead.On("SetFilter").Return(&gorm.DB{})
		brokerMock.On("Publish").Return(nil)
		mockWrite.On("Create").Return(respCompany)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		body, _ := json.Marshal(inviteUser)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return status 500 when something unexpected happened", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		respCompany := &response.Response{}
		respAccount := &response.Response{}
		mockRead.On("Find").Once().Return(respAccount.SetData(account))
		mockRead.On("Find").Return(respCompany.SetData(company))
		mockRead.On("SetFilter").Return(&gorm.DB{})
		brokerMock.On("Publish").Return(errors.New("test"))
		mockWrite.On("Create").Return(respCompany)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		body, _ := json.Marshal(inviteUser)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return status 409 when user already in company", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		respCompany := &response.Response{}
		respAccount := &response.Response{}
		respWithError := &response.Response{}

		mockRead.On("Find").Once().Return(respAccount.SetData(account))
		mockRead.On("Find").Return(respCompany.SetData(company))
		mockRead.On("SetFilter").Return(&gorm.DB{})
		mockWrite.On("Create").Return(
			respWithError.SetError(errors.New(errorsEnum.ErrorAlreadyExistingCompanyID)))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		body, _ := json.Marshal(inviteUser)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("should return status 404 when not found account or company", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		respCompany := &response.Response{}
		respAccount := &response.Response{}
		respWithError := &response.Response{}

		mockRead.On("Find").Once().Return(respAccount.SetData(account))
		mockRead.On("Find").Return(respCompany.SetData(company))
		mockRead.On("SetFilter").Return(&gorm.DB{})
		mockWrite.On("Create").Return(respWithError.SetError(errorsEnum.ErrNotFoundRecords))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		body, _ := json.Marshal(inviteUser)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return status 400 when missing or invalid company id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		body, _ := json.Marshal(inviteUser)
		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return status 400 when invalid request body", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodPost, "api/companies/", bytes.NewReader([]byte("")))
		w := httptest.NewRecorder()

		handler.InviteUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteCompany(t *testing.T) {
	t.Run("should return 204 when success delete company", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.Delete(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return 500 when something went wrong", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp.SetError(errors.New("test")))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.Delete(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return 400 when invalid request id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		handler.Delete(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAccountsFromCompany(t *testing.T) {
	t.Run("should return 200 when success get accounts", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		accounts := &[]roles.AccountRole{{Email: "test@test.com", Username: "test", Role: "member"}}

		accountsResp := &response.Response{}
		mockRead.On("RawSQL").Return(accountsResp.SetData(accounts))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodGet, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.GetAccounts(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 500 when something went wrong", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		accountsResp := &response.Response{}
		mockRead.On("RawSQL").Return(accountsResp.SetError(errors.New("test")))

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodGet, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.GetAccounts(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return 400 when invalid company id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodGet, "api/companies/", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRemoveUser(t *testing.T) {
	account := authEntities.Account{}

	t.Run("should return 204 when successfully delete", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetData(account))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return 500 when something went wrong", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetError(errors.New("test")))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return 500 when something went wrong", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetError(errors.New("test")))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should return 404 when account was not found", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetError(errorsEnum.ErrNotFoundRecords))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("companyID", uuid.New().String())
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when missing company id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetError(errorsEnum.ErrNotFoundRecords))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("accountID", uuid.New().String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when missing account id", func(t *testing.T) {
		mockWrite := &relational.MockWrite{}
		mockRead := &relational.MockRead{}
		brokerMock := &broker.Mock{}

		resp := &response.Response{}
		mockWrite.On("Delete").Return(resp)
		mockRead.On("Find").Return(resp.SetError(errorsEnum.ErrNotFoundRecords))
		mockRead.On("SetFilter").Return(&gorm.DB{})

		handler := NewHandler(mockWrite, mockRead, brokerMock, &app.Config{})

		r, _ := http.NewRequest(http.MethodDelete, "api/companies/", nil)
		w := httptest.NewRecorder()

		handler.RemoveUser(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
