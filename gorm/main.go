package main

import (
	"fmt"
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

func main() {
	//findDemo()
	var u User
	// 1. 关联查询默认是懒加载，需要手动preload
	// 按需加载：只有当你实际访问关联字段时才会执行SQL查询
	// 额外查询：每个关联都会产生单独的查询（N+1问题）
	// 默认行为：GORM默认使用懒加载方式处理关联
	// 1. Preload 关联需要每次preload
	// 预加载方式（一次查询解决）
	db.Debug().Preload("Profile").Preload("Orders").Where("id =?", 1).Find(&u)
	fmt.Println(u)
}

func findDemo() {
	var u User
	ret := db.Debug().Find(&u, 1)
	if ret.Error != nil {
		if gorm.ErrRecordNotFound == ret.Error {
			fmt.Println("记录不存在")
		}
		fmt.Println("查询失败", ret.Error)
	}
	fmt.Println(u)
	// 1.
	var u1 []User
	db.Debug().Where("id<?", 7).Find(&u1)
	fmt.Println(u1)
	// 2. find情况切片并填充
	db.Debug().Where("username like ?", "%zhang%").Find(&u1)
	fmt.Println(u1)
	// 3.子查询
	avg := db.Debug().Model(&User{}).Select("avg(id)")
	db.Debug().Where("id<?", avg).Find(&u1)
	fmt.Println(u1)
	// 4.指定字段
	var u2 []struct {
		Username string
	}
	db.Debug().Model(&User{}).Select("username").Where("id <?", 5).Find(&u2)
	fmt.Println(u2)
	// 5.链式
	db.Debug().Select("id > ?", 2).Order("username desc").Limit(3).Find(&u1)
	fmt.Println(u1)
}
