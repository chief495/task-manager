package sheduler

import (
	"log"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/user"
)

type Scheduler struct {
	userRepo *user.UserRepo
}

func NewScheduler(userRepo *user.UserRepo) *Scheduler {
	return &Scheduler{
		userRepo: userRepo,
	}
}

func (s *Scheduler) Start() {
	// Определяем время следующего запуска — 09:00
	nextRun := s.next9AM()
	log.Printf("Запуск планировщика. Следующее выполнение: %v", nextRun)

	// Ждём до 09:00
	time.Sleep(time.Until(nextRun))

	// Выполняем задачу и запускаем цикл
	s.runDailyTask()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runDailyTask()
		}
	}
}

func (s *Scheduler) next9AM() time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())

	if !now.Before(next) {
		// Если уже прошло 09:00, переносим на завтра
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (s *Scheduler) runDailyTask() {
	log.Println("Запуск ежедневной задачи: сохранение SocialRating всех пользователей")

	users, err := s.userRepo.GetActiveUsers()
	if err != nil {
		log.Printf("Ошибка получения пользователей: %v", err)
		return
	}

	for _, user := range users {
		logEntry := &models.SocialRatingLog{
			UserID:       user.ID,
			Rating:       user.SocialRating,
			ChangeReason: "Ежедневное логирование",
		}

		if err := s.userRepo.GetDB().Create(logEntry).Error; err != nil {
			log.Printf("Ошибка сохранения лога для пользователя %d: %v", user.ID, err)
		} else {
			log.Printf("Сохранён SocialRating для пользователя %d: %.2f", user.ID, user.SocialRating)
		}
	}
}
