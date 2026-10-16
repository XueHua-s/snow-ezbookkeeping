package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/api"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// Generation is application-owned. Signal handling first stops new claims and
// cancels existing upstream work, then drains HTTP and workers within a deadline.
func serveAssistantApplication(c *core.CliContext, config *settings.Config, handler http.Handler, address string) error {
	network := "tcp"
	switch config.Protocol {
	case settings.SCHEME_SOCKET:
		network = "unix"
		address = config.UnixSocketPath
	case settings.SCHEME_HTTP, settings.SCHEME_HTTPS:
	default:
		return errs.ErrInvalidProtocol
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	defer listener.Close()
	if network == "unix" {
		defer os.Remove(address)
	}
	app, stop := signal.NotifyContext(c, os.Interrupt, syscall.SIGTERM)
	defer stop()
	api.LargeLanguageModels.InitializeAssistantExperience()
	api.LargeLanguageModels.StartAssistantExperience(app)
	server := &http.Server{Handler: handler}
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		<-app.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
		_ = api.LargeLanguageModels.StopAssistantExperience(shutdown)
	}()
	if config.Protocol == settings.SCHEME_HTTPS {
		err = server.ServeTLS(listener, config.CertFile, config.CertKeyFile)
	} else {
		err = server.Serve(listener)
	}
	stop()
	<-drained
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
