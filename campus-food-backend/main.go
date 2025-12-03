package main

import (
	"campus-food-backend/internal/config"
	"campus-food-backend/internal/database"
	"campus-food-backend/internal/handlers"
	"campus-food-backend/internal/middleware"
	"campus-food-backend/internal/services"
	"campus-food-backend/internal/utils"
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

var dbConn *sql.DB

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化 JWT（使用配置中的密钥）
	utils.InitJWT(cfg.JWTSecret)

	// 连接数据库
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	dbConn = db.DB // 保存数据库连接供中间件使用

	// 初始化服务
	authService := services.NewAuthService(db.DB)
	orderService := services.NewOrderService(db.DB)
	complaintService := services.NewComplaintService(db.DB)
	dataService := services.NewDataService(db.DB)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService)
	complaintHandler := handlers.NewComplaintHandler(complaintService)

	// 创建 Gin 实例
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "校园食堂代买餐平台后端服务运行正常",
		})
	})

	// 表检查接口
	r.GET("/api/debug/tables", func(c *gin.Context) {
		tables := []string{"orders", "orderdetail", "dish"}
		results := make(map[string]interface{})

		for _, table := range tables {
			// 检查表是否存在
			var tableExists int
			err := db.DB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'campus_food' AND table_name = ?", table).Scan(&tableExists)

			if err != nil {
				results[table] = gin.H{"error": err.Error()}
				continue
			}

			if tableExists > 0 {
				// 获取表结构
				rows, err := db.DB.Query("DESCRIBE " + table)
				if err != nil {
					results[table] = gin.H{"exists": true, "structure_error": err.Error()}
				} else {
					var columns []string
					for rows.Next() {
						var field, fieldType, null, key, extra string
						var defaultValue *string
						rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
						columns = append(columns, field)
					}
					rows.Close()
					results[table] = gin.H{"exists": true, "columns": columns}
				}
			} else {
				results[table] = gin.H{"exists": false}
			}
		}

		c.JSON(200, gin.H{
			"success": true,
			"tables":  results,
		})
	})

	// 路由组
	api := r.Group("/api")
	{
		// 认证路由
		//无需认证
		pub := api.Group("/auth")
		{
			pub.POST("/login", authHandler.Login)
			pub.POST("/register", authHandler.Register)
		}
		//需要认证
		pri := api.Group("/auth").Use(middleware.AuthMiddleware())
		{
			pri.GET("/profile", authHandler.GetProfile)
			pri.PUT("/profile", authHandler.UpdateProfile)
			pri.PUT("/password", authHandler.UpdatePassword)
			pri.POST("logout", authHandler.Logout)
		}
		// 订单路由
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			// 创建和查询
			orders.POST("", orderHandler.CreateOrder) // 发布代买需求
			orders.GET("", orderHandler.GetOrders)    // 获取订单列表（可按状态筛选）

			// 统计和状态
			orders.GET("/stats", orderHandler.GetOrderStats)          // 订单统计
			orders.GET("/running", orderHandler.GetMyRunningOrders)   // 进行中的订单
			orders.GET("/available", orderHandler.GetAvailableOrders) // 可接订单（跑腿员）

			// 订单详情
			orders.GET("/:id", orderHandler.GetFullOrderDetails)    // 完整详情（带菜品信息）
			orders.GET("/:id/simple", orderHandler.GetOrderDetails) // 简单详情

			// 订单操作
			orders.POST("/:id/take", orderHandler.TakeOrder)         // 接单
			orders.POST("/:id/complete", orderHandler.CompleteOrder) // 完成订单
			orders.POST("/:id/cancel", orderHandler.CancelOrder)     // 取消订单

			// 状态管理
			orders.PUT("/status", orderHandler.UpdateOrderStatus)                       // 通用状态更新
			orders.GET("/:id/available-status", orderHandler.GetAvailableStatusOptions) // 获取可操作状态
		}

		// 投诉路由
		complaints := api.Group("/complaints")
		complaints.Use(middleware.AuthMiddleware())
		{
			complaints.POST("", complaintHandler.CreateComplaint)
			complaints.GET("", middleware.AdminMiddleware(), complaintHandler.GetComplaints)
		}

		// 数据接口
		data := api.Group("/data")
		//data.Use(middleware.AuthMiddleware())
		{
			data.GET("/canteens", func(c *gin.Context) {
				canteens, err := dataService.GetCanteens()
				if err != nil {
					utils.Error(c, "获取食堂列表失败")
					return
				}
				utils.Success(c, canteens)
			})

			data.GET("/merchants", func(c *gin.Context) {
				canteenIDStr := c.Query("canteen_id")
				if canteenIDStr == "" {
					utils.Error(c, "缺少食堂ID参数")
					return
				}

				canteenID, err := strconv.Atoi(canteenIDStr)
				if err != nil {
					utils.Error(c, "食堂ID格式错误")
					return
				}

				merchants, err := dataService.GetMerchants(canteenID)
				if err != nil {
					utils.Error(c, "获取商家列表失败")
					return
				}
				utils.Success(c, merchants)
			})

			data.GET("/dishes", func(c *gin.Context) {
				merchantIDStr := c.Query("merchant_id")
				if merchantIDStr == "" {
					utils.Error(c, "缺少商家ID参数")
					return
				}

				merchantID, err := strconv.Atoi(merchantIDStr)
				if err != nil {
					utils.Error(c, "商家ID格式错误")
					return
				}

				dishes, err := dataService.GetDishes(merchantID)
				if err != nil {
					utils.Error(c, "获取菜品列表失败")
					return
				}
				utils.Success(c, dishes)
			})
		}
	}

	// 启动服务器
	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
