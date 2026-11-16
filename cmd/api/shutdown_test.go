package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestServer_GracefulShutdown_ControlledHandler(t *testing.T) {
	// Каналы для жесткой синхронизации порядка событий
	handlerStarted := make(chan struct{})
	allowFinish := make(chan struct{})

	// 1. Создаем HTTP-сервер с управляемым обработчиком
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Сигнализируем тесту, что запрос принят и обработка началась
		close(handlerStarted)

		// Блокируемся, пока тест не разрешит завершить запрос (или пока не сработает защита от зависания)
		select {
		case <-allowFinish:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("processed"))
		case <-r.Context().Done():
			return
		}
	}))
	// Гарантируем закрытие ресурсов при любом исходе теста
	defer ts.Close()

	// 2. Запускаем сервер и отправляем запрос в отдельной горутине
	ts.Start()

	type clientResult struct {
		status int
		body   string
		err    error
	}
	clientResCh := make(chan clientResult, 1)

	go func() {
		resp, err := ts.Client().Get(ts.URL + "/work")
		if err != nil {
			clientResCh <- clientResult{err: err}
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		clientResCh <- clientResult{
			status: resp.StatusCode,
			body:   string(data),
			err:    err,
		}
	}()

	// 3. Дожидаемся сигнала, что обработчик начал работу (с защитным таймаутом)
	select {
	case <-handlerStarted:
		// Обработчик внутри, соединение активно
	case <-time.After(2 * time.Second):
		t.Fatal("таймаут ожидания старта обработчика")
	}

	// 4. Вызываем Shutdown() в отдельной горутине
	shutdownErrCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Обращаемся напрямую к базовому *http.Server внутри httptest.Server
		shutdownErrCh <- ts.Config.Shutdown(ctx)
	}()

	// 5. Убеждаемся, что завершение не произошло, пока обработчик заблокирован.
	// Используем небольшой таймаут, чтобы дать горутине с Shutdown() гарантированно дойти до точки ожидания.
	select {
	case err := <-shutdownErrCh:
		t.Fatalf("Shutdown() завершился преждевременно при активном запросе (err: %v)", err)
	case <-time.After(50 * time.Millisecond):
		// Ожидаемое поведение: Shutdown() честно ждет завершения активного хендлера
	}

	// 6. Разблокируем обработчик
	close(allowFinish)

	// 7. Проверяем, что клиент получил корректный ответ, а Shutdown() завершился без ошибки
	select {
	case res := <-clientResCh:
		if res.err != nil {
			t.Fatalf("ошибка HTTP-клиента: %v", res.err)
		}
		if res.status != http.StatusOK {
			t.Errorf("ожидался статус 200 OK, получен %d", res.status)
		}
		if res.body != "processed" {
			t.Errorf("ожидалось тело ответа 'processed', получено %q", res.body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("таймаут ожидания ответа от клиента")
	}

	select {
	case err := <-shutdownErrCh:
		if err != nil {
			t.Fatalf("ожидалось успешное завершение Shutdown(), получена ошибка: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("таймаут ожидания завершения Shutdown()")
	}
}

func TestServer_ShutdownTimeout_DeadlineExceeded(t *testing.T) {
	handlerStarted := make(chan struct{})
	allowFinish := make(chan struct{})

	// Гарантируем безопасное однократное закрытие канала при очистке теста
	var finishOnce sync.Once
	unblockHandler := func() {
		finishOnce.Do(func() {
			close(allowFinish)
		})
	}

	// 1. Создаем HTTP-сервер с управляемым обработчиком
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted)

		select {
		case <-allowFinish:
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	}))

	// Регистрируем очистку в правильном порядке (t.Cleanup выполняется в порядке LIFO):
	// Сначала разблокируется хендлер, затем ts.Close() закроет соединения и дождется завершения горутин.
	t.Cleanup(func() {
		unblockHandler()
		ts.Close()
	})

	// 2. Запускаем сервер и отправляем запрос в отдельной горутине
	ts.Start()

	clientErrCh := make(chan error, 1)
	go func() {
		resp, err := ts.Client().Get(ts.URL + "/work")
		if err != nil {
			clientErrCh <- err
			return
		}
		resp.Body.Close()
		clientErrCh <- nil
	}()

	// 3. Дожидаемся сигнала, что обработчик начал работу
	select {
	case <-handlerStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("таймаут ожидания старта обработчика")
	}

	// 4. Вызываем Shutdown() с коротким таймаутом и НЕ разблокируем обработчик
	shutdownErrCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		shutdownErrCh <- ts.Config.Shutdown(ctx)
	}()

	// 5. Проверяем, что Shutdown() вернул управление по истечении таймаута с ошибкой context.DeadlineExceeded
	select {
	case err := <-shutdownErrCh:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("ожидалась ошибка context.DeadlineExceeded, получено: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown() не вернул управление после истечения таймаута")
	}

	// 6. Симулируем принудительное закрытие сервера из main.go (server.Close())
	// и проверяем, что клиент сразу получил ошибку обрыва соединения
	if err := ts.Config.Close(); err != nil {
		t.Fatalf("ошибка принудительного закрытия сервера: %v", err)
	}

	select {
	case err := <-clientErrCh:
		if err == nil {
			t.Error("ожидалась ошибка обрыва соединения у клиента после server.Close(), но запрос завершился успешно")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("клиент завис после принудительного закрытия сервера")
	}

	// После выхода из теста сработает t.Cleanup():
	// unblockHandler() отпустит горутину хендлера, а ts.Close() корректно завершит работу тестового сервера.
}
