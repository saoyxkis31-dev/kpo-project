package repository

import (
	"Confetki/internal/model"
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotEnoughProduct = errors.New("not enough product")
var ErrOrderAlreadyCanceled = errors.New("order is already canceled")
var ErrCannotCancelOrder = errors.New("can not cancel order (it is completed)")

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (o *OrderRepository) CreateFullOrder(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback(ctx)

	slices.SortFunc(items, func(a, b model.OrderItem) int {
		return a.ProductID - b.ProductID
	})

	for _, item := range items {
		var currentStock int
		err := tx.QueryRow(ctx,
			`select stock
		from products where id = $1 for update`,
			item.ProductID).Scan(&currentStock)
		if err != nil {
			return model.Order{}, err
		}
		if currentStock < item.Quantity {
			return model.Order{}, ErrNotEnoughProduct
		}
		_, err = tx.Exec(ctx,
			`update products set stock = stock - $1  where id = $2`,
			item.Quantity, item.ProductID)
		if err != nil {
			return model.Order{}, err
		}
	}
	var createdOrder model.Order
	err = tx.QueryRow(ctx, `insert into orders (user_id, status, total_price)
		values ($1, $2, $3)
		returning id, user_id, status, total_price, created_at`,
		order.UserID, order.Status, order.TotalPrice,
	).Scan(
		&createdOrder.ID,
		&createdOrder.UserID,
		&createdOrder.Status,
		&createdOrder.TotalPrice,
		&createdOrder.CreatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	for _, item := range items {
		_, err := tx.Exec(ctx, `insert into order_items (order_id, product_id, quantity, price)
    	values ($1, $2, $3, $4)`,
			createdOrder.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return model.Order{}, err
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return model.Order{}, err
	}
	return createdOrder, nil
}

func (o *OrderRepository) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	var order model.Order
	sqlQuery := `select 
					o.id, o.user_id, o.status, o.total_price, o.created_at,
    				oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price
				from orders o
				join order_items oi on o.id = oi.order_id
				where o.id = $1`
	rows, err := o.db.Query(ctx, sqlQuery, id)
	if err != nil {
		return model.Order{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalPrice,
			&order.CreatedAt,
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return model.Order{}, err
		}
		order.Items = append(order.Items, item)
	}
	err = rows.Err()
	if err != nil {
		return model.Order{}, err
	}
	if order.ID == 0 {
		return model.Order{}, pgx.ErrNoRows
	}
	return order, nil
}

func (o *OrderRepository) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := o.db.Query(
		ctx,
		`SELECT id, user_id, status, total_price, created_at
	FROM orders
	ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalPrice,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *OrderRepository) UpdateStatus(ctx context.Context, orderID int, status string) error {
	res, err := o.db.Exec(ctx,
		`update orders set status = $1 where id = $2;`,
		status, orderID)

	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (o *OrderRepository) CancelOrderWithStockReturn(ctx context.Context, orderID int) error {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var status string

	err = tx.QueryRow(ctx,
		`select status from orders where id = $1 for update;`,
		orderID).Scan(&status)
	if err != nil {
		return err
	}

	switch status {
	case "canceled":
		return ErrOrderAlreadyCanceled
	case "new", "paid":
	default:
		return ErrCannotCancelOrder
	}

	rows, err := tx.Query(ctx,
		`select product_id, quantity from order_items where order_id = $1`,
		orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		productID int
		quantity  int
	}
	var items []item

	for rows.Next() {
		var it item
		if err := rows.Scan(&it.productID, &it.quantity); err != nil {
			return err
		}
		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`update products set stock = stock + $1 where id = $2`,
			item.quantity, item.productID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx,
		`update orders set status = 'canceled' where id = $1`,
		orderID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (o *OrderRepository) GetOrderByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	rows, err := o.db.Query(
		ctx,
		`select id, user_id, status, total_price, created_at
		from  orders where user_id = $1
		order by id`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalPrice,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
