package par

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type CreateUser interface { //创建对象时使用的Hook
	// gorm 新版本钩子需要传递 *gorm.DB，V2 不需要
	BeforeCreate(*gorm.DB) error
	//BeforeSave() error
	AfterCreate(*gorm.DB) error
	//AfterSave() error
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == 1 {
		return errors.New("无效插入 id ==1")
	}
	return nil
}

// AfterCreate → 只有在SQL执行成功后才会执行
func (u *User) AfterCreate(tx *gorm.DB) error {
	fmt.Println("AfterCreate")
	return nil
}

func HookCreate() {
	u := User{ID: 1, Username: "李元芳", Email: "lyf@yf.com"}
	ret := db.Create(&u).Error
	fmt.Println(u, ret)
}
