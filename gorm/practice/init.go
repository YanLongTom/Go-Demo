package par

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           int       `gorm:"column:id;primaryKey"` // 明确映射表字段名（若与结构体字段名一致可省略）
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	Profile      Profile   `gorm:"foreignKey:Username;references:Username"` // 添加关联
	Orders       Order     `gorm:"foreignKey:Username;references:Username"` // 一对多关联
	Roles        []Role    `gorm:"foreignKey:Username;references:Username"` // 一对多关联
}

// 定义关联模型
type Profile struct {
	Username string `gorm:"column:username"`
	Realname string `gorm:"column:real_name"`
	Age      int    `gorm:"column:age"`
	phone    string `gorm:"column:phone"`
}
type Order struct {
	ID       int     `gorm:"column:id;primaryKey"`
	Username string  `gorm:"column:username"` // 外键关联User
	Amount   float64 `gorm:"column:amount"`
	Status   string  `gorm:"column:status"`
}

// 添加Role模型
type Role struct {
	ID       int    `gorm:"column:id;primaryKey"`
	Username string `gorm:"column:username"` // 外键关联User
	RoleName string `gorm:"column:role_name"`
}

func (User) TableName() string {
	return "user"
}

func (Profile) TableName() string {
	return "user_profile"
}

func (Order) TableName() string {
	return "user_orders"
}

func (Role) TableName() string {
	return "user_roles"
}

var db *gorm.DB

func init() {
	var err error
	dsn := "root:123456@tcp(172.31.0.7:3306)/testtable?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}
}

// Scan 方法：将数据库值转换为User结构体
func (u *User) Scan(value interface{}) error {
	var ret string
	switch v := value.(type) {
	case string:
		ret = v
	case []byte:
		ret = string(v)
	default:
		ret = "nil"
	}
	return json.Unmarshal([]byte(ret), u)
}

// Valuer: 将 Go 类型转换为数据库可存储的值
func (u User) Value() (driver.Value, error) {
	ret, _ := json.Marshal(u)
	return string(ret), nil
}
