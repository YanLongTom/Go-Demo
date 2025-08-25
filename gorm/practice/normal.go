package par

import (
	"fmt"
	"gorm.io/gorm"
)

// 基础操作
func FindDemo() {
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
	// 6.count 必须放在最后
	a := int64(0)
	db.Debug().Select("id > ?", 2).Order("username desc").Limit(3).Find(&u1).Count(&a)
	fmt.Println(a)
	// 	db.Debug().Select("id > ?", 2).Order("username desc").Count(&a).Limit(3).Find(&u1) 被find覆盖为0
	db.Debug().Select("id > ?", 2).Order("username desc").Find(&u1).Count(&a).Limit(3) //7
	fmt.Println(a)
}

// update & insert
func UI() {
	// 0值处理时，会使用默认值而非0值，需要使用指针
	var u = User{Username: "hello", PasswordHash: " "}
	db.Debug().Create(&u)
	// 或者实现 Scanner/Valuer 接口
	db.Debug().Create(&User{Username: "test"})
}
