// @title API Управления Статусами и Переходами
// @version 1.0
// @description API для управления статусами, стадиями и правилами переходов между ними.
// @termsOfService http://example.com/terms/
//
// @contact.name Поддержка API
// @contact.url http://example.com/support
// @contact.email support@example.com
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
package main

import (
	"log"

	"github.com/eduardnesterov/task-manager/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
