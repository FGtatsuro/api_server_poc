package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type contextKey string

const (
	apiBase         = "http://httpbin.org/headers"
	tokenContextKey = contextKey("token")
)

type service interface {
	init()
	run()
}

type apiExampleService struct {
	server *http.Server
	wg     *sync.WaitGroup
	stopCh *chan struct{}
}

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

func (srv *apiExampleService) init() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {

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

	srv.server = &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      nil,
	}
}

func (srv *apiExampleService) run() {
	go func() {
		if err := srv.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("%v\n", err)
		}
	}()
	go func() {
		defer srv.wg.Done()

		<-*srv.stopCh

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		//ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		if err := srv.server.Shutdown(ctx); err != nil {
			// Call cancel func before os.Exit.
			//   FYI: https://golang.org/pkg/os/#Exit
			cancel()
			log.Fatalf("%v\n", err)
		} else {
			fmt.Printf("Successful shutdown\n")
		}

	}()
}

func main() {
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	wg.Add(1)
	srv := apiExampleService{
		wg:     &wg,
		stopCh: &stopCh,
	}
	srv.init()
	srv.run()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
		close(stopCh)
		wg.Wait()
	}
}
