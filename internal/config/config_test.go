package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Переходим во временную пустую директорию, чтобы локальный .env файл
	// проекта не вмешался в работу cleanenv.ReadEnv
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	tests := []struct {
		name        string
		envVars     map[string]string
		wantAddr    string
		wantTimeout time.Duration
		wantErr     bool
	}{
		{
			name:        "Переменные отсутствуют — используются defaults",
			envVars:     map[string]string{},
			wantAddr:    ":8080",
			wantTimeout: 5 * time.Second, // В cleanenv число без суффикса парсится как наносекунды!
			wantErr:     false,
		},
		{
			name: "Установлен HTTP_ADDR=:9090 — адрес меняется",
			envVars: map[string]string{
				"HTTP_ADDR": ":9090",
			},
			wantAddr:    ":9090",
			wantTimeout: 5 * time.Second,
			wantErr:     false,
		},
		{
			name: "Установлен SHUTDOWN_TIMEOUT=invalid — возвращается ошибка",
			envVars: map[string]string{
				"SHUTDOWN_TIMEOUT": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Гарантируем, что переменные не заданы в системе до старта подтеста
			os.Unsetenv("HTTP_ADDR")
			os.Unsetenv("SHUTDOWN_TIMEOUT")

			// Устанавливаем переменные для текущего сценария
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := LoadConfig()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, но получили nil (cfg: %+v)", cfg)
				}
				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if cfg.Addr != tt.wantAddr {
				t.Errorf("Addr = %q, ожидалось %q", cfg.Addr, tt.wantAddr)
			}

			if cfg.ShutdownTOut != tt.wantTimeout {
				t.Errorf("ShutdownTOut = %v, ожидалось %v", cfg.ShutdownTOut, tt.wantTimeout)
			}
		})
	}
}
