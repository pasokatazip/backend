package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/database"
	notificationinfra "github.com/pasokatazip/backend/internal/infrastructure/notification"
	"github.com/pasokatazip/backend/internal/infrastructure/persistence"
	"github.com/pasokatazip/backend/internal/usecases"
)

const (
	defaultCronLocation = "Asia/Tokyo"
	defaultReportHour   = 23
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresDB(requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	location, err := time.LoadLocation(envOrDefault("CRON_LOCATION", defaultCronLocation))
	if err != nil {
		log.Fatalf("failed to load cron location: %v", err)
	}

	reportHour := envIntOrDefault("REPORT_NOTIFICATION_HOUR", defaultReportHour)
	if reportHour < 0 || reportHour > 23 {
		log.Fatalf("REPORT_NOTIFICATION_HOUR must be between 0 and 23")
	}

	sender, err := notificationinfra.NewWebPushSender(notificationinfra.WebPushSenderConfig{
		VAPIDPublicKey:  requiredEnv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: requiredEnv("VAPID_PRIVATE_KEY"),
		Subject:         requiredEnv("VAPID_SUBJECT"),
		TTL:             envIntOrDefault("WEB_PUSH_TTL", 60),
	})
	if err != nil {
		log.Fatalf("failed to create web push sender: %v", err)
	}

	notificationRepo := persistence.NewNotificationRepository(db)
	sendNotification := usecases.NewSendNotification(notificationRepo, sender)

	simulationRepo := persistence.NewPetSimulationRepository(db)
	runHourlySimulation := usecases.NewRunHourlyPetSimulation(simulationRepo)

	log.Printf(
		"cron started: hourly simulation runs every hour; report notification runs every day at %02d:00 %s",
		reportHour,
		location.String(),
	)

	go runHourly(ctx, location, func(runCtx context.Context, simulatedAt time.Time) {
		_ = runCtx

		output, err := runHourlySimulation.Execute(usecases.RunHourlyPetSimulationInput{
			SimulatedAt: &simulatedAt,
		})
		if err != nil {
			log.Printf("failed to run hourly simulation: %v", err)
			return
		}

		log.Printf(
			"hourly simulation completed: simulated_at=%s total_pets=%d processed=%d skipped=%d",
			output.SimulatedAt.Format(time.RFC3339),
			output.TotalPets,
			output.Processed,
			output.Skipped,
		)
	})

	go runDaily(ctx, location, reportHour, func(runCtx context.Context) {
		output, err := sendNotification.Execute(runCtx, usecases.SendNotificationInput{
			Type:  domain.NotificationTypeReport,
			Title: "Reportができました",
			Body:  "今日のReportを確認してみよう",
			Data:  json.RawMessage(`{"type":"report"}`),
		})
		if err != nil {
			log.Printf("failed to send report notification: %v", err)
			return
		}

		log.Printf(
			"report notification sent: targets=%d sent=%d failed=%d",
			output.TargetCount,
			output.SentCount,
			output.FailedCount,
		)
		for _, sendErr := range output.Errors {
			log.Printf("report notification send error: %s", sendErr)
		}
	})

	<-ctx.Done()
	log.Println("cron stopped")
}

func runDaily(ctx context.Context, location *time.Location, hour int, job func(context.Context)) {
	for {
		next := nextDailyRun(time.Now().In(location), hour, location)
		timer := time.NewTimer(time.Until(next))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			job(ctx)
		}
	}
}

func runHourly(ctx context.Context, location *time.Location, job func(context.Context, time.Time)) {
	for {
		next := nextHourlyRun(time.Now().In(location), location)
		timer := time.NewTimer(time.Until(next))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			job(ctx, next)
		}
	}
}

func nextDailyRun(now time.Time, hour int, location *time.Location) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, location)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func nextHourlyRun(now time.Time, location *time.Location) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, location).Add(time.Hour)
	if !next.After(now) {
		next = next.Add(time.Hour)
	}
	return next
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func envOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func envIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("%s must be an integer", key)
	}
	return n
}
