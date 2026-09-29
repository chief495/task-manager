package app

import (
	"net/http"
	"strings"

	platformauth "github.com/eduardnesterov/task-manager/internal/platform/auth"
	platformweb "github.com/eduardnesterov/task-manager/internal/platform/web"
	"github.com/gin-gonic/gin"
)

func buildRouter(handlers *appHandlers, debug bool) *gin.Engine {
	router := gin.Default()

	if debug {
		router.Use(noCacheStaticMiddleware())
	}

	router.SetHTMLTemplate(platformweb.LoadTemplates())
	router.Static("/static", platformweb.StaticDir())

	registerRootRoute(router)
	registerStaticAliases(router)
	registerPublicPages(router)
	registerAPI(router, handlers)
	registerProtectedPages(router)

	return router
}

func noCacheStaticMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	}
}

func registerStaticAliases(router *gin.Engine) {
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("images/logo.png", http.Dir(platformweb.StaticDir()))
	})
	router.GET("/apple-touch-icon.png", func(c *gin.Context) {
		c.FileFromFS("images/logo.png", http.Dir(platformweb.StaticDir()))
	})
	router.GET("/apple-touch-icon-precomposed.png", func(c *gin.Context) {
		c.FileFromFS("images/logo.png", http.Dir(platformweb.StaticDir()))
	})
}

func registerRootRoute(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			var err error
			token, err = c.Cookie("token")
			if err != nil {
				c.Redirect(http.StatusTemporaryRedirect, "/login")
				return
			}
		}

		claims := &platformauth.Claims{}
		_, err := platformauth.ParseToken(token, claims)
		if err != nil {
			secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
			c.SetCookie("token", "", -1, "/", "", secure, true)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/tasks")
	})
}

func registerPublicPages(router *gin.Engine) {
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	router.GET("/login.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
}

func registerAPI(router *gin.Engine, handlers *appHandlers) {
	v1 := router.Group("/api/v1")
	handlers.auth.RegisterRoutes(v1)

	protected := v1.Group("/")
	protected.Use(platformauth.AuthMiddleware())
	{
		handlers.project.RegisterRoutes(protected)
		handlers.projectPhase.RegisterRoutes(protected)
		handlers.task.RegisterRoutes(protected)
		handlers.report.RegisterRoutes(protected)
		handlers.user.RegisterRoutes(protected)
		handlers.status.RegisterRoutes(protected)
		handlers.statusTransition.RegisterRoutes(protected)
		handlers.stage.RegisterRoutes(protected)
		handlers.stageTransition.RegisterRoutes(protected)
		handlers.system.RegisterRoutes(protected)
		handlers.product.RegisterRoutes(protected)
		handlers.workType.RegisterRoutes(protected)
		handlers.systemType.RegisterRoutes(protected)
		handlers.productStage.RegisterRoutes(protected)
		handlers.organization.RegisterRoutes(protected)
		handlers.organizationContact.RegisterRoutes(protected)
		handlers.participant.RegisterRoutes(protected)
		handlers.trip.RegisterRoutes(protected)
		handlers.export.RegisterRoutes(protected)
		handlers.taskTemplate.RegisterRoutes(protected)
		handlers.labor.RegisterRoutes(protected)
	}
}

func registerProtectedPages(router *gin.Engine) {
	authenticated := router.Group("/")
	authenticated.Use(platformauth.AuthMiddleware())
	{
		authenticated.GET("/tasks", renderBasePage("Задачи", "/static/js/tasks.js"))
		authenticated.GET("/admin/statuses", renderBasePage("Админ: Статусы", "/static/js/admin-statuses.js"))
		authenticated.GET("/admin/stages", renderBasePage("Админ: Стадии", "/static/js/admin-stages.js"))
		authenticated.GET("/admin/projects", renderBasePage("Админ: Проекты", "/static/js/admin-projects.js"))
		authenticated.GET("/manager/contacts", renderBasePage("Компании и контакты", "/static/js/manager-contacts.js"))
		authenticated.GET("/manager/contacs", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/manager/contacts")
		})
		authenticated.GET("/admin/contacts", renderBasePage("Компании и контакты", "/static/js/manager-contacts.js"))
		authenticated.GET("/admin/users", renderBasePage("Админ: Пользователи", "/static/js/admin-users.js"))
		authenticated.GET("/reports", renderBasePage("Отчёты", "/static/js/reports.js"))
		authenticated.GET("/admin/reports", renderBasePage("Отчёты (Админ)", "/static/js/admin-reports.js"))
		authenticated.GET("/admin/task-templates", renderBasePage("Шаблоны задач", "/static/js/task-templates.js"))
		authenticated.GET("/admin/timesheet", renderBasePage("Табель", "/static/js/timesheet.js"))
		authenticated.GET("/labor", renderBasePage("Трудозатраты", "/static/js/labor.js"))
	}
}

func renderBasePage(title, script string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":   title,
			"Scripts": []string{platformweb.AssetURL(script)},
		})
	}
}
