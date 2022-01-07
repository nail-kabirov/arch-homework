package postgres

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"arch-homework/pkg/common/infrastructure/postgres"
	"arch-homework/pkg/order/app"
)

func NewOrderRepository(client postgres.Client) app.OrderRepository {
	return &orderRepository{client: client}
}

type orderRepository struct {
	client postgres.Client
}

func (repo *orderRepository) Store(order *app.Order) error {
	const query = `
			INSERT INTO orders (id, user_id, price, created_at)
			VALUES (:id, :user_id, :price, :created_at)
			ON CONFLICT (id) DO UPDATE SET
				price = excluded.price;
		`

	orderx := sqlxOrder{
		ID:           string(order.ID),
		UserID:       string(order.UserID),
		Price:        order.Price.RawValue(),
		CreationDate: order.CreationDate,
	}

	_, err := repo.client.NamedExec(query, &orderx)
	return errors.WithStack(err)
}

func (repo *orderRepository) FindByID(id app.OrderID) (*app.Order, error) {
	const query = `SELECT id, user_id, price, created_at FROM orders WHERE id = $1`

	var order sqlxOrder
	err := repo.client.Get(&order, query, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrOrderNotFound
		}
		return nil, errors.WithStack(err)
	}
	res := sqlxOrderToOrder(&order)
	return &res, nil
}

func (repo *orderRepository) FindAllByUserID(userID app.UserID) ([]app.Order, error) {
	const query = `SELECT id, user_id, price, created_at FROM orders where user_id = $1`

	var orders []*sqlxOrder
	err := repo.client.Select(&orders, query, string(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := make([]app.Order, 0, len(orders))
	for _, order := range orders {
		res = append(res, sqlxOrderToOrder(order))
	}
	return res, nil
}

func sqlxOrderToOrder(order *sqlxOrder) app.Order {
	return app.Order{
		ID:           app.OrderID(order.ID),
		Price:        app.PriceFromRawValue(order.Price),
		CreationDate: order.CreationDate,
	}
}

type sqlxOrder struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Price        uint64    `db:"price"`
	CreationDate time.Time `db:"created_at"`
}
