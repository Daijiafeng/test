package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"testmind/internal/config"
	"testmind/internal/handler"
	"testmind/internal/middleware"
	"testmind/internal/repository"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置运行模式
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 连接数据库
	db, err := repository.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	log.Println("Database connected successfully")

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 创建Handler
	userHandler := handler.NewUserHandler(cfg)
	orgHandler := handler.NewOrganizationHandler(db)
	projectHandler := handler.NewProjectHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(cfg)

	// API路由
	api := r.Group("/api/v1")
	{
		// 认证相关（无需鉴权）
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.Refresh)
		}

		// 需要鉴权的路由
		authorized := api.Group("")
		authorized.Use(authMiddleware.JWTAuth())
		{
			// 用户相关
			authorized.GET("/auth/profile", userHandler.GetProfile)
			authorized.PUT("/auth/profile", userHandler.UpdateProfile)
			authorized.POST("/auth/logout", userHandler.Logout)

			// 组织管理
			authorized.POST("/organizations", orgHandler.Create)
			authorized.GET("/organizations", orgHandler.List)
			authorized.GET("/organizations/:org_id", orgHandler.Get)
			authorized.PUT("/organizations/:org_id", orgHandler.Update)
			authorized.GET("/organizations/:org_id/projects", orgHandler.ListProjects)

			// 项目管理
			authorized.POST("/projects", projectHandler.Create)
			authorized.GET("/projects/:project_id", projectHandler.Get)
			authorized.PUT("/projects/:project_id", projectHandler.Update)
			authorized.DELETE("/projects/:project_id", projectHandler.Delete)
			authorized.GET("/projects/:project_id/members", projectHandler.ListMembers)
			authorized.POST("/projects/:project_id/members", projectHandler.AddMember)
			authorized.DELETE("/projects/:project_id/members/:user_id", projectHandler.RemoveMember)
			authorized.PUT("/projects/:project_id/members/:user_id", projectHandler.UpdateMemberRole)

			// 测试计划（待实现）
			// authorized.POST("/projects/:project_id/plans", planHandler.Create)
			// authorized.GET("/projects/:project_id/plans", planHandler.List)
			// authorized.GET("/plans/:plan_id", planHandler.Get)
			// authorized.PUT("/plans/:plan_id", planHandler.Update)
			// authorized.DELETE("/plans/:plan_id", planHandler.Delete)

			// 测试用例（待实现）
			// authorized.POST("/projects/:project_id/cases", caseHandler.Create)
			// authorized.GET("/projects/:project_id/cases", caseHandler.List)
			// authorized.GET("/cases/:case_id", caseHandler.Get)
			// authorized.PUT("/cases/:case_id", caseHandler.Update)
			// authorized.DELETE("/cases/:case_id", caseHandler.Delete)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "user-svc",
			"version": "1.0.0",
		})
	})

	// 启动服务
	addr := cfg.Server.Host + ":" + itoa(cfg.Server.Port)
	log.Printf("Starting user-svc on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte(n%10) + '0'}, result...)
		n /= 10
	}
	return string(result)
}