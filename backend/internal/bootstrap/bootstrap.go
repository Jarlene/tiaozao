package bootstrap

import (
	"fmt"
	"log"

	"github.com/Jarlene/tiaozao/backend/internal/config"
	"github.com/Jarlene/tiaozao/backend/internal/handler"
	"github.com/Jarlene/tiaozao/backend/internal/middleware"
	"github.com/Jarlene/tiaozao/backend/internal/model"
	"github.com/Jarlene/tiaozao/backend/internal/repository"
	"github.com/Jarlene/tiaozao/backend/internal/router"
	"github.com/Jarlene/tiaozao/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run 启动应用
func Run() {
	// 加载 .env 文件（忽略加载失败的情况）
	_ = godotenv.Load()

	// 加载配置
	cfg := config.Load()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 连接数据库
	db := connectDB(&cfg.Database)

	// 自动迁移数据库表
	autoMigrate(db)

	// 初始化各层依赖
	// Repository 层
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Service 层
	userService := service.NewUserService(userRepo, cfg)
	productService := service.NewProductService(productRepo, userRepo)
	orderService := service.NewOrderService(orderRepo, productRepo, userRepo)

	// Handler 层
	userHandler := handler.NewUserHandler(userService, productService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)

	// Middleware
	authMiddleware := middleware.NewJWTAuth(cfg.JWT.Secret)

	// 初始化路由
	r := gin.Default()
	router.Setup(r, authMiddleware, userHandler, productHandler, orderHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("跳蚤市场服务启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// connectDB 连接PostgreSQL数据库
func connectDB(cfg *config.DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	log.Println("数据库连接成功")
	return db
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.ProductImage{},
		&model.Order{},
		&model.OrderItem{},
		&model.Cart{},
		&model.Address{},
		&model.Message{},
		&model.Review{},
		&model.Favorite{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")
}
