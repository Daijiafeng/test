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
	planHandler := handler.NewTestPlanHandler(db)
	caseHandler := handler.NewTestCaseHandler(db)
	moduleHandler := handler.NewModuleHandler(db)
	customFieldHandler := handler.NewCustomFieldHandler(db)
	executionHandler := handler.NewExecutionHandler(db)
	defectHandler := handler.NewDefectHandler(db)
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

			// 测试计划
			authorized.POST("/projects/:project_id/plans", planHandler.Create)
			authorized.GET("/projects/:project_id/plans", planHandler.List)
			authorized.GET("/plans/:plan_id", planHandler.Get)
			authorized.PUT("/plans/:plan_id", planHandler.Update)
			authorized.DELETE("/plans/:plan_id", planHandler.Delete)
			authorized.POST("/plans/:plan_id/requirements", planHandler.AddRequirement)
			authorized.GET("/plans/:plan_id/requirements", planHandler.ListRequirements)
			authorized.POST("/plans/:plan_id/cases", planHandler.AddCases)
			authorized.DELETE("/plans/:plan_id/cases", planHandler.RemoveCases)
			authorized.GET("/plans/:plan_id/progress", planHandler.GetProgress)

			// 测试用例
			authorized.POST("/projects/:project_id/cases", caseHandler.Create)
			authorized.GET("/projects/:project_id/cases", caseHandler.List)
			authorized.GET("/projects/:project_id/cases/search", caseHandler.Search)
			authorized.POST("/projects/:project_id/cases/batch", caseHandler.BatchCreate)
			authorized.GET("/cases/:case_id", caseHandler.Get)
			authorized.PUT("/cases/:case_id", caseHandler.Update)
			authorized.DELETE("/cases/:case_id", caseHandler.Delete)
			authorized.GET("/cases/:case_id/versions", caseHandler.GetVersions)
			authorized.POST("/cases/:case_id/review", caseHandler.Review)

			// 模块管理
			authorized.POST("/projects/:project_id/modules", moduleHandler.Create)
			authorized.GET("/projects/:project_id/modules", moduleHandler.List)
			authorized.PUT("/modules/:module_id", moduleHandler.Update)
			authorized.DELETE("/modules/:module_id", moduleHandler.Delete)

			// 自定义字段
			authorized.POST("/projects/:project_id/custom-fields", customFieldHandler.Create)
			authorized.GET("/projects/:project_id/custom-fields", customFieldHandler.List)
			authorized.PUT("/custom-fields/:field_id", customFieldHandler.Update)
			authorized.DELETE("/custom-fields/:field_id", customFieldHandler.Delete)

			// 测试执行
			authorized.POST("/plans/:plan_id/executions", executionHandler.Create)
			authorized.GET("/plans/:plan_id/executions", executionHandler.List)
			authorized.GET("/executions/:execution_id", executionHandler.Get)
			authorized.PUT("/executions/:execution_id", executionHandler.Update)
			authorized.POST("/plans/:plan_id/executions/batch", executionHandler.BatchCreate)
			authorized.POST("/executions/:execution_id/defect", executionHandler.CreateDefect)
			authorized.GET("/plans/:plan_id/statistics", executionHandler.GetStatistics)
			authorized.GET("/cases/:case_id/executions", executionHandler.ListByCase)
			// 自动化执行
			authorized.POST("/plans/:plan_id/auto-run", executionHandler.AutoRun)
			authorized.GET("/auto-tasks/:task_id/status", executionHandler.AutoStatus)
			authorized.POST("/executions/:execution_id/external-result", executionHandler.ExternalResult)

			// 缺陷管理
			authorized.POST("/projects/:project_id/defects", defectHandler.Create)
			authorized.GET("/projects/:project_id/defects", defectHandler.List)
			authorized.GET("/projects/:project_id/defects/search", defectHandler.Search)
			authorized.GET("/projects/:project_id/defects/statistics", defectHandler.Statistics)
			authorized.GET("/projects/:project_id/defects/export", defectHandler.Export)
			authorized.PUT("/projects/:project_id/defects/batch", defectHandler.BatchUpdate)
			authorized.GET("/defects/:defect_id", defectHandler.Get)
			authorized.PUT("/defects/:defect_id", defectHandler.Update)
			authorized.POST("/defects/:defect_id/transition", defectHandler.TransitionStatus)
			authorized.POST("/defects/:defect_id/assign", defectHandler.Assign)
			authorized.POST("/defects/:defect_id/comments", defectHandler.AddComment)
			authorized.GET("/defects/:defect_id/comments", defectHandler.ListComments)
			authorized.GET("/defects/:defect_id/history", defectHandler.GetHistory)

			// TODO: 测试报告
			// TODO: AI生成
			// TODO: 飞书集成
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "testmind-api",
			"version": "1.0.0",
		})
	})

	// 启动服务
	addr := cfg.Server.Host + ":" + itoa(cfg.Server.Port)
	log.Printf("🦞 TestMind API running on %s", addr)
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