package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"testmind/internal/config"
	"testmind/internal/handler"
	"testmind/internal/middleware"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置运行模式
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 创建Handler
	userHandler := handler.NewUserHandler(cfg)
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
			// authorized.POST("/organizations", orgHandler.Create)
			// authorized.GET("/organizations/:org_id", orgHandler.Get)
			// authorized.PUT("/organizations/:org_id", orgHandler.Update)
			// authorized.GET("/organizations/:org_id/projects", orgHandler.ListProjects)

			// 项目管理
			// authorized.POST("/projects", projectHandler.Create)
			// authorized.GET("/projects/:project_id", projectHandler.Get)
			// authorized.PUT("/projects/:project_id", projectHandler.Update)
			// authorized.DELETE("/projects/:project_id", projectHandler.Delete)
			// authorized.GET("/projects/:project_id/members", projectHandler.ListMembers)
			// authorized.POST("/projects/:project_id/members", projectHandler.AddMember)
			// authorized.DELETE("/projects/:project_id/members/:user_id", projectHandler.RemoveMember)

			// 测试计划
			// authorized.POST("/projects/:project_id/plans", planHandler.Create)
			// authorized.GET("/projects/:project_id/plans", planHandler.List)
			// authorized.GET("/plans/:plan_id", planHandler.Get)
			// authorized.PUT("/plans/:plan_id", planHandler.Update)
			// authorized.DELETE("/plans/:plan_id", planHandler.Delete)

			// 测试用例
			// authorized.POST("/projects/:project_id/cases", caseHandler.Create)
			// authorized.GET("/projects/:project_id/cases", caseHandler.List)
			// authorized.GET("/cases/:case_id", caseHandler.Get)
			// authorized.PUT("/cases/:case_id", caseHandler.Update)
			// authorized.DELETE("/cases/:case_id", caseHandler.Delete)

			// AI生成用例
			// authorized.POST("/projects/:project_id/cases/ai/generate", aiHandler.GenerateCases)
			// authorized.GET("/projects/:project_id/cases/ai/generate/:task_id/status", aiHandler.GetGenerateStatus)

			// 测试执行
			// authorized.POST("/plans/:plan_id/executions", execHandler.Create)
			// authorized.GET("/plans/:plan_id/executions", execHandler.List)
			// authorized.GET("/executions/:execution_id", execHandler.Get)

			// 缺陷管理
			// authorized.POST("/projects/:project_id/defects", defectHandler.Create)
			// authorized.GET("/projects/:project_id/defects", defectHandler.List)
			// authorized.GET("/defects/:defect_id", defectHandler.Get)
			// authorized.PUT("/defects/:defect_id", defectHandler.Update)
			// authorized.POST("/defects/:defect_id/transition", defectHandler.Transition)
			// authorized.POST("/defects/:defect_id/comments", defectHandler.AddComment)

			// 测试报告
			// authorized.POST("/plans/:plan_id/reports", reportHandler.Generate)
			// authorized.GET("/projects/:project_id/reports", reportHandler.List)
			// authorized.GET("/reports/:report_id", reportHandler.Get)

			// 系统配置
			// authorized.POST("/projects/:project_id/custom-fields", configHandler.CreateCustomField)
			// authorized.GET("/projects/:project_id/custom-fields", configHandler.ListCustomFields)
			// authorized.POST("/projects/:project_id/modules", configHandler.CreateModule)
			// authorized.GET("/projects/:project_id/modules", configHandler.ListModules)
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
	addr := cfg.Server.Host + ":" + string(rune(cfg.Server.Port))
	log.Printf("Starting user-svc on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}