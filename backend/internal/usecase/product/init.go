package product_usecase

import (
	"time"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"

	"github.com/sirupsen/logrus"
)

type productUsecase struct {
	txManager         repository.TransactionManager
	productRepository repository.ProductRepository
	log               *logrus.Logger

	// now is injected so tests can pin timestamps. Production: time.Now.
	now func() time.Time
}

func NewProductUsecase(txManager repository.TransactionManager, productRepository repository.ProductRepository, log *logrus.Logger) usecase.ProductUsecase {
	return &productUsecase{
		txManager:         txManager,
		productRepository: productRepository,
		log:               log,
		now:               time.Now,
	}
}
