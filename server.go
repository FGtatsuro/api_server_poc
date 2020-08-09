package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type contextKey string

const (
	apiBase         = "http://httpbin.org/headers"
	tokenContextKey = contextKey("token")
)

// TODO: どうやってテストする?
func innerHandler(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", ctx.Value(tokenContextKey).(string))

	// TODO: bodyをスマートに受け取る方法が知りたい
	errCh := make(chan error, 1)
	bodyCh := make(chan []byte, 1)
	defer close(errCh)
	defer close(bodyCh)
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}

		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errCh <- err
			return
		}
		bodyCh <- b
	}()

	select {
	case body := <-bodyCh:
		return body, nil
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return nil, nil
}

func apiBuild() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Middlewareで上手く処理する例: https://deeeet.com/writing/2016/07/22/context/
		token := r.Header.Get("authorization")
		if token == "" {
			http.Error(w, "Bearer token is required", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), tokenContextKey, token)
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		// ctx, cancel := context.WithTimeout(ctx, 10*time.Nanosecond)
		defer cancel()

		// innerHandlerがrequest時に新たなgoroutineを作っているため、ここでgoroutineを作る必要は実のところなさそう
		// innerHandlerがgoroutineを使っている/使っていないによらず、requestが別goroutine上で行われる事を保証できるイディオムとして使えるか?
		errCh := make(chan error, 1)
		bodyCh := make(chan []byte, 1)
		defer close(errCh)
		defer close(bodyCh)
		go func() {
			b, err := innerHandler(ctx)
			if err != nil {
				errCh <- err
				return
			}
			bodyCh <- b

		}()

		// TODO: errorハンドリング: 特にHTTPステータスをどう判定するべきか?
		select {
		case body := <-bodyCh:
			_, err := w.Write(body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		case err := <-errCh:
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	})
}

func main() {
	apiBuild()

	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      nil,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Error in server termination: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	errCh := make(chan error, 1)
	defer close(errCh)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Nanosecond)
	go func() {
		defer cancel()
		errCh <- server.Shutdown(ctx)
	}()

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err == context.DeadlineExceeded {
			log.Printf("Error in graceful shutdown: %v\n", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil {
			log.Printf("Error in connection close: %v\n", err)
		}
	}
	fmt.Printf("Successful shutdown\n")
	os.Exit(0)
}
