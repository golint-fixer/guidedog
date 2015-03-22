package execution

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	environment "github.com/9seconds/guide-dog/environment"
	options "github.com/9seconds/guide-dog/options"
)

func Execute(command []string, env *environment.Environment) int {
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

	attachSignalChannel(supervisorChannel, signalChannel)
	if env.Options.Supervisor == options.SUPERVISOR_MODE_RESTARTING {
		attachSupervisorChannel(supervisorChannel, watcherChannel)
	}

	supervisor := NewSupervisor(command,
		exitCodeChannel,
		env.Options.Signal,
		env.Options.GracefulTimeout,
		env.Options.PTY,
		env.Options.Supervisor != options.SUPERVISOR_MODE_NONE,
		supervisorChannel)

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
	go func() {
		for incomingSignal := range signalChannel {
			log.WithField("signal", incomingSignal).Debug("Signal from OS received.")
			channel <- SUPERVISOR_STOP
		}
	}()
}

func attachSupervisorChannel(channel chan SupervisorAction, supervisorChannel chan bool) {
	go func() {
		for event := range supervisorChannel {
			log.WithFields(log.Fields{
				"event":   event,
				"channel": supervisorChannel,
			}).Debug("Event from supervisor channel is captured.")
			channel <- SUPERVISOR_RESTART
		}
	}()
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
