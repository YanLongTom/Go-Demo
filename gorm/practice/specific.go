package par

import (
	"fmt"
	"gorm.io/gorm/clause"
)

func Specific1() {
	var u User
	// 1. 关联查询默认是懒加载，需要手动preload
	// 按需加载：只有当你实际访问关联字段时才会执行SQL查询
	// 额外查询：每个关联都会产生单独的查询（N+1问题）
	// 默认行为：GORM默认使用懒加载方式处理关联
	// 1. Preload 关联需要每次preload
	// 预加载方式（一次查询解决）
	db.Debug().Preload("Profile").Preload("Orders").Where("id =?", 1).Find(&u)
	fmt.Println(u)
	//
	users := []User{
		{Username: "user1", Email: "user1@example.com"},
		{Username: "hello", Email: "user2@example.com"},
	}
	// 冲突检测处理
	// INSERT INTO `user`  VALUES ('hello','user2@example.com','','2025-08-26 20:21:08.67')
	//ON DUPLICATE KEY UPDATE `email`=VALUES(`email`),`new@123`=VALUES(`new@123`)
	db.Debug().Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "username"}}, //检测冲突
			//DoNothing: true,
			DoUpdates: clause.AssignmentColumns([]string{"email", "new@123"}), //更新
		}).Create(users)
}

func Specific2() {
	//1. FirstOrInit 找不到记录，则初始化结构体但不保存到数据库
	var user User
	// 重点：会和Scanner/Valuer 冲突
	db.Debug().FirstOrInit(&user, User{Username: "non_existing"})
	fmt.Println("name", user.Username)
	// 2.找不到就要插入数据了，需要使用结构体或者map
	// 还是会和Scanner/Valuer 冲突
	db.Debug().FirstOrCreate(&user, User{Username: "non_existing", Email: "non_existing@example.com"})
}
