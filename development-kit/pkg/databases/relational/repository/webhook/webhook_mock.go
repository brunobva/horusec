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

package webhook

import (
	"github.com/ZupIT/horusec/development-kit/pkg/entities/webhook"
	utilsMock "github.com/ZupIT/horusec/development-kit/pkg/utils/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) GetByRepositoryID(_ uuid.UUID) (*webhook.Webhook, error) {
	args := m.MethodCalled("GetByRepositoryID")
	return args.Get(0).(*webhook.Webhook), utilsMock.ReturnNilOrError(args, 1)
}
func (m *Mock) GetByWebhookID(_ uuid.UUID) (*webhook.Webhook, error) {
	args := m.MethodCalled("GetByWebhookID")
	return args.Get(0).(*webhook.Webhook), utilsMock.ReturnNilOrError(args, 1)
}
func (m *Mock) GetAllByCompanyID(_ uuid.UUID) (*[]webhook.ResponseWebhook, error) {
	args := m.MethodCalled("GetAllByCompanyID")
	return args.Get(0).(*[]webhook.ResponseWebhook), utilsMock.ReturnNilOrError(args, 1)
}
func (m *Mock) Create(_ *webhook.Webhook) error {
	args := m.MethodCalled("Create")
	return utilsMock.ReturnNilOrError(args, 0)
}
func (m *Mock) Update(_ *webhook.Webhook) error {
	args := m.MethodCalled("Update")
	return utilsMock.ReturnNilOrError(args, 0)
}
func (m *Mock) Remove(_ uuid.UUID) error {
	args := m.MethodCalled("Remove")
	return utilsMock.ReturnNilOrError(args, 0)
}
