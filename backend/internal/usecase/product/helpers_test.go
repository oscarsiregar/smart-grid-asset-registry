package product_usecase

import (
	"io"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	mocksrepo "smart-grid-asset-registry/internal/mocks/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var fixedNow = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

// deps bundles the mocks one test case needs. Mocks are created per case so
// mockery's AssertExpectations (registered via t.Cleanup) checks exactly the
// calls that case expects — an unexpected call fails the test.
type deps struct {
	txManager *mocksrepo.TransactionManager
	trx       *mocksrepo.Tx
	repo      *mocksrepo.ProductRepository
}

func newDeps(t *testing.T) deps {
	return deps{
		txManager: mocksrepo.NewTransactionManager(t),
		trx:       mocksrepo.NewTx(t),
		repo:      mocksrepo.NewProductRepository(t),
	}
}

// withinTx wires txManager.WithinTx to run the callback against d.trx.
func (d deps) withinTx() {
	mocksrepo.StubWithinTx(d.txManager, d.trx)
}

func (d deps) usecase() productUsecase {
	return productUsecase{
		txManager:         d.txManager,
		productRepository: d.repo,
		log:               newTestLogger(),
		now:               func() time.Time { return fixedNow },
	}
}

func sampleProduct() *entity.Product {
	desc := "a widget"
	return &entity.Product{
		ID:            uuid.New(),
		Code:          "P-001",
		Name:          "Widget",
		Description:   &desc,
		Price:         10,
		Status:        valueobject.ProductStatusActive,
		CreatedAt:     fixedNow.Add(-time.Hour),
		CreatedByName: "alice",
		UpdatedAt:     fixedNow.Add(-time.Hour),
		UpdatedByName: "alice",
	}
}

func TestHasPermission(t *testing.T) {
	assert.True(t, hasPermission([]string{"a", "b"}, "b"))
	assert.False(t, hasPermission([]string{"a"}, "b"))
	assert.False(t, hasPermission(nil, "b"))
}

func TestNewProductUsecase(t *testing.T) {
	assert.NotNil(t, NewProductUsecase(nil, nil, newTestLogger()))
}
