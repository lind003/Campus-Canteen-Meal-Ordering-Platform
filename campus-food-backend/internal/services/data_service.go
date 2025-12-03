package services

import (
	"database/sql"
	"fmt"
)

type DataService struct {
	db *sql.DB
}

func NewDataService(db *sql.DB) *DataService {
	return &DataService{db: db}
}

type Canteen struct {
	CanteenID   int    `json:"canteen_id" db:"canteen_id"`
	CanteenName string `json:"canteen_name" db:"canteen_name"`
	Location    string `json:"location" db:"location"`
	OpenTime    string `json:"open_time" db:"open_time"`
}

type Merchant struct {
	MerchantID    int    `json:"merchant_id" db:"merchant_id"`
	MerchantName  string `json:"merchant_name" db:"merchant_name"`
	WindowNo      string `json:"window_no" db:"window_no"`
	CanteenID     int    `json:"canteen_id" db:"canteen_id"`
	Description   string `json:"description,omitempty" db:"description"`
	BusinessHours string `json:"business_hours,omitempty" db:"business_hours"`
	Phone         string `json:"phone,omitempty" db:"mer_phone"`
}

type Dish struct {
	DishID      int     `json:"dish_id" db:"dish_id"`
	DishName    string  `json:"dish_name" db:"dish_name"`
	Price       float64 `json:"price" db:"price"`
	SpicyLevel  int     `json:"spicy_level" db:"spicy_level"`
	MerchantID  int     `json:"merchant_id" db:"merchant_id"`
	Description string  `json:"description,omitempty" db:"description"`
	Category    string  `json:"category,omitempty" db:"category"`
	IsAvailable bool    `json:"is_available,omitempty" db:"is_available"`
}

func (s *DataService) GetCanteens() ([]Canteen, error) {
	rows, err := s.db.Query("SELECT canteen_id, canteen_name, location, open_time FROM Canteen")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var canteens []Canteen
	for rows.Next() {
		var canteen Canteen
		err := rows.Scan(&canteen.CanteenID, &canteen.CanteenName, &canteen.Location, &canteen.OpenTime)
		if err != nil {
			return nil, err
		}
		canteens = append(canteens, canteen)
	}

	return canteens, nil
}

func (s *DataService) GetMerchants(canteenID int) ([]Merchant, error) {
	rows, err := s.db.Query(`
        SELECT merchant_id, merchant_name, window_no, canteen_id,
               description, business_hours, mer_phone
        FROM Merchant WHERE canteen_id = ?
    `, canteenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []Merchant
	for rows.Next() {
		var merchant Merchant
		var description, businessHours, phone sql.NullString // 处理可能为NULL的字段

		err := rows.Scan(
			&merchant.MerchantID, &merchant.MerchantName, &merchant.WindowNo, &merchant.CanteenID,
			&description, &businessHours, &phone,
		)
		if err != nil {
			return nil, err
		}

		// 处理NULL值
		if description.Valid {
			merchant.Description = description.String
		}
		if businessHours.Valid {
			merchant.BusinessHours = businessHours.String
		}
		if phone.Valid {
			merchant.Phone = phone.String
		}

		merchants = append(merchants, merchant)
	}

	return merchants, nil
}

func (s *DataService) GetDishes(merchantID int) ([]Dish, error) {
	rows, err := s.db.Query(`
        SELECT dish_id, dish_name, price, spicy_level, merchant_id,
               description, category, is_available
        FROM Dish WHERE merchant_id = ?
    `, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []Dish
	for rows.Next() {
		var dish Dish
		var description, category sql.NullString
		var isAvailable sql.NullBool

		err := rows.Scan(
			&dish.DishID, &dish.DishName, &dish.Price, &dish.SpicyLevel, &dish.MerchantID,
			&description, &category, &isAvailable,
		)
		if err != nil {
			return nil, err
		}

		// 处理NULL值
		if description.Valid {
			dish.Description = description.String
		}
		if category.Valid {
			dish.Category = category.String
		}
		if isAvailable.Valid {
			dish.IsAvailable = isAvailable.Bool
		} else {
			dish.IsAvailable = true // 默认为可用
		}

		dishes = append(dishes, dish)
	}

	// 添加日志输出
	if len(dishes) > 0 {
		fmt.Printf("GetDishes返回 %d 个菜品，第一个菜品: %+v\n", len(dishes), dishes[0])
	} else {
		fmt.Printf("GetDishes返回 0 个菜品\n")
	}

	return dishes, nil
}
