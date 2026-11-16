package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// unsetEnv безопасно удаляет переменную окружения на время теста
// и гарантированно восстанавливает её исходное состояние после завершения.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	origVal, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, origVal)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	// Путь к заведомо отсутствующему .env во временной папке:
	// рабочая директория процесса (os.Getwd) вообще не трогается.
	missingEnvPath := filepath.Join(t.TempDir(), ".env")

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
			wantTimeout: 5 * time.Second,
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
			// Очищаем переменные на время подтеста с гарантированным откатом через t.Cleanup
			unsetEnv(t, "HTTP_ADDR")
			unsetEnv(t, "SHUTDOWN_TIMEOUT")

			// Устанавливаем нужные для сценария переменные
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := LoadConfig(missingEnvPath)

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
