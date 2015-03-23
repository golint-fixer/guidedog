package execution

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	environment "github.com/9seconds/guide-dog/environment"
	options "github.com/9seconds/guide-dog/options"
)

func Execute(command []string, env *environment.Environment) int {
	if env.Options.LockFile != nil {
		for {
			if err := env.Options.LockFile.Acquire(); err == nil {
				defer env.Options.LockFile.Release()
				break
			}
			time.Sleep(LOCK_FILE_TIMEOUT)
		}
	}

	pathsToWatch := []string{env.Options.ConfigPath}
	for _, path := range env.Options.PathsToTrack {
		pathsToWatch = append(pathsToWatch, path)
	}

	watcherChannel := makeWatcher(pathsToWatch, env)
	defer close(watcherChannel)

	exitCodeChannel := make(chan int, 1)
	defer close(exitCodeChannel)

	supervisorChannel := make(chan SupervisorAction, 1)
	defer close(supervisorChannel)

	signalChannel := makeSignalChannel()
	defer close(signalChannel)

	go attachSignalChannel(supervisorChannel, signalChannel)
	if env.Options.Supervisor == options.SUPERVISOR_MODE_RESTARTING {
		go attachSupervisorChannel(supervisorChannel, watcherChannel)
	}

	supervisor := NewSupervisor(command,
		exitCodeChannel,
		env.Options.Signal,
		env.Options.GracefulTimeout,
		env.Options.PTY,
		env.Options.Supervisor&options.SUPERVISOR_MODE_SIMPLE > 0,
		supervisorChannel)

	log.WithField("supervisor", supervisor).Info("Start supervisor.")

	supervisor.Start()
	go func() {
		for {
			event, ok := <-supervisorChannel
			if !ok {
				return
			}
			supervisor.Signal(event)
		}
	}()

	return <-exitCodeChannel
}

func attachSignalChannel(channel chan SupervisorAction, signalChannel chan os.Signal) {
	for {
		incomingSignal, ok := <-signalChannel
		if !ok {
			return
		}
		log.WithField("signal", incomingSignal).Debug("Signal from OS received.")
		channel <- SUPERVISOR_STOP
	}
}

func attachSupervisorChannel(channel chan SupervisorAction, supervisorChannel chan bool) {
	for {
		event, ok := <-supervisorChannel
		if !ok {
			return
		}
		log.WithFields(log.Fields{
			"event":   event,
			"channel": supervisorChannel,
		}).Debug("Event from supervisor channel is captured.")
		channel <- SUPERVISOR_RESTART
	}
}

func makeSignalChannel() (channel chan os.Signal) {
	channel = make(chan os.Signal, 1)

	signal.Notify(channel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	return channel
}