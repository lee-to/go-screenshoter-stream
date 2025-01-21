package screenshot

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockCapture struct {
	mock.Mock
}

func (m *MockCapture) Capture(ctx context.Context, url string) ([]byte, error) {
	args := m.Called(ctx, url)

	return args.Get(0).([]byte), args.Error(1)
}
