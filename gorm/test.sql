-- user 表
CREATE TABLE USER (
                      id INT AUTO_INCREMENT PRIMARY KEY,  -- 唯一标识用户，自增主键
                      username VARCHAR(50) NOT NULL UNIQUE,  -- 用户名，唯一且非空
                      email VARCHAR(100) NOT NULL UNIQUE,  -- 邮箱，唯一且非空（用于登录/通知）
                      password_hash VARCHAR(255) NOT NULL,  -- 密码哈希（不存储明文）
                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- 账户创建时间，默认当前时间
);

-- 插入单条用户数据
INSERT INTO USER (username, email, password_hash, created_at)
VALUES (
           'zhangsan',
           'zhangsan@example.com',
           '$2a$10$xG9rE7hQv8L3Gf8KjH7eju5rY8mZ9nO0p1q2r3s4t5u6v7w8x9y0',  -- 模拟bcrypt加密后的密码哈希
           '2024-08-25 08:30:00'
       );

-- 批量插入多条用户数据
INSERT INTO USER (username, email, password_hash, created_at)
VALUES
    (
        'qianliu',
        'qianliu@example.com',
        '$2a$10$aBcDeFgHiJkLmNoPqRsTuVwXyZ1234567890abcdefghijklmno',
        '2024-08-25 09:15:00'
    ),
    (
        'wuba',
        'wuba@example.com',
        '$2a$10$1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN',
        '2024-08-25 10:00:00'
    );



-- 关联表

-- 1. 创建用户资料表（存储用户详细信息，一对一关联）
CREATE TABLE `user_profile` (
                                `profile_id` INT(11) NOT NULL AUTO_INCREMENT,
                                `username` VARCHAR(50) NOT NULL, -- 关联user表的username
                                `real_name` VARCHAR(100) DEFAULT NULL,
                                `age` INT(11) DEFAULT NULL,
                                `phone` VARCHAR(20) DEFAULT NULL,
                                PRIMARY KEY (`profile_id`),
    -- 外键约束：关联user表的username
                                CONSTRAINT `fk_profile_username` FOREIGN KEY (`username`)
                                    REFERENCES `user` (`username`)
                                    ON DELETE CASCADE  -- 用户删除时，资料也删除
                                    ON UPDATE CASCADE  -- 用户名更新时，资料同步更新
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 2. 创建用户订单表（存储用户订单，一对多关联）
CREATE TABLE `user_orders` (
                               `order_id` INT(11) NOT NULL AUTO_INCREMENT,
                               `username` VARCHAR(50) NOT NULL, -- 关联user表的username
                               `order_no` VARCHAR(50) NOT NULL,
                               `amount` DECIMAL(10,2) NOT NULL,
                               `order_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (`order_id`),
    -- 外键约束：关联user表的username
                               CONSTRAINT `fk_orders_username` FOREIGN KEY (`username`)
                                   REFERENCES `user` (`username`)
                                   ON DELETE CASCADE
                                   ON UPDATE CASCADE
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 3. 创建用户权限表（存储用户角色权限，一对多关联）
CREATE TABLE `user_roles` (
                              `role_id` INT(11) NOT NULL AUTO_INCREMENT,
                              `username` VARCHAR(50) NOT NULL, -- 关联user表的username
                              `role_name` VARCHAR(50) NOT NULL, -- 例如：admin, user, vip
                              PRIMARY KEY (`role_id`),
    -- 外键约束：关联user表的username
                              CONSTRAINT `fk_roles_username` FOREIGN KEY (`username`)
                                  REFERENCES `user` (`username`)
                                  ON DELETE CASCADE
                                  ON UPDATE CASCADE
) ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



-- 插入数据

-- 1. 向用户资料表插入数据
INSERT INTO `user_profile` (`username`, `real_name`, `age`, `phone`)
VALUES
    ('zhangsan', '张三', 28, '13800138000'),
    ('lisi', '李四', 32, '13900139000'),
    ('wangwu', '王五', 25, '13700137000'),
    ('qianliu', '钱六', 40, '13600136000'),
    ('wuba', '吴八', 35, '13500135000');

-- 2. 向用户订单表插入数据
INSERT INTO `user_orders` (`username`, `order_no`, `amount`)
VALUES
    ('zhangsan', 'ORD20240825001', 199.99),
    ('zhangsan', 'ORD20240825002', 499.50),
    ('lisi', 'ORD20240825003', 2999.00),
    ('wangwu', 'ORD20240825004', 89.90),
    ('qianliu', 'ORD20240825005', 5999.99),
    ('wuba', 'ORD20240825006', 129.00),
    ('wuba', 'ORD20240825007', 359.00);

-- 3. 向用户权限表插入数据
INSERT INTO `user_roles` (`username`, `role_name`)
VALUES
    ('zhangsan', 'user'),
    ('lisi', 'vip'),
    ('wangwu', 'user'),
    ('qianliu', 'admin'),
    ('qianliu', 'vip'), -- 一个用户可以有多个角色
    ('wuba', 'user');


SELECT u.username, u.email, p.real_name, p.age
FROM `user` u
         JOIN `user_profile` p ON u.username = p.username;
