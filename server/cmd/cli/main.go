package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"chatix/internal/config"
	"chatix/internal/models"
	"chatix/internal/repository"
	"chatix/internal/service"
	"chatix/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("не удалось загрузить конфиг: %v", err)
	}

	ctx := context.Background()

	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.Pool)
	roleRepo := repository.NewRoleRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)
	userService := service.NewUserService(userRepo, roleRepo, auditRepo)

	switch os.Args[1] {
	case "create-admin":
		handleCreateAdmin(ctx, userService, cfg)
	case "list-users":
		fmt.Println("Команда в разработке")
	default:
		printUsage()
		os.Exit(1)
	}
}

func handleCreateAdmin(ctx context.Context, userService *service.UserService, cfg *config.Config) {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	username := fs.String("username", "", "Логин администратора")
	password := fs.String("password", "", "Пароль администратора")
	displayName := fs.String("name", "", "Отображаемое имя")
	email := fs.String("email", "", "Email (опционально)")
	fs.Parse(os.Args[2:])

	if *username == "" || *password == "" || *displayName == "" {
		fmt.Println("Ошибка: --username, --password и --name обязательны")
		fs.Usage()
		os.Exit(1)
	}

	input := models.UserCreate{
		Username:    *username,
		Password:    *password,
		DisplayName: *displayName,
		Status:      models.UserStatusActive,
	}

	if *email != "" {
		input.Email = email
	}

	user, err := userService.CreateUser(ctx, input, []string{"admin"}, nil, nil)
	if err != nil {
		log.Fatalf("ошибка создания админа: %v", err)
	}

	fmt.Printf("Администратор создан:\n")
	fmt.Printf("  ID: %s\n", user.ID)
	fmt.Printf("  Логин: %s\n", user.Username)
	fmt.Printf("  Имя: %s\n", user.DisplayName)
}

func printUsage() {
	fmt.Println("Использование: chatix-cli <команда>")
	fmt.Println()
	fmt.Println("Команды:")
	fmt.Println("  create-admin    Создать администратора")
	fmt.Println("  list-users      Список пользователей (в разработке)")
	fmt.Println()
	fmt.Println("Пример:")
	fmt.Println("  chatix-cli create-admin --username admin --password secret --name \"Администратор\"")
}