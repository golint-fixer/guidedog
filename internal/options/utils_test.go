package options

import (
	"strings"
	"syscall"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestParseEnvsEmpty(t *testing.T) {
	parsed := parseEnvs([]string{})

	assert.Equal(t, parsed, make(map[string]string))
}

func TestParseEnvs(t *testing.T) {
	envs := []string{"env=1", "empty", "empty2=", "complex=1=1", "complex2=1=1=1"}
	parsed := parseEnvs(envs)

	assert.Equal(t, len(envs), len(parsed))
	assert.Equal(t, parsed["env"], "1")
	assert.Equal(t, parsed["empty"], "")
	assert.Equal(t, parsed["empty2"], "")
	assert.Equal(t, parsed["complex"], "1=1")
	assert.Equal(t, parsed["complex2"], "1=1=1")
}

func TestParseSignalName(t *testing.T) {
	signals := map[string]syscall.Signal{
		"abrt":   syscall.SIGABRT,
		"alrm":   syscall.SIGALRM,
		"bus":    syscall.SIGBUS,
		"chld":   syscall.SIGCHLD,
		"cont":   syscall.SIGCONT,
		"fpe":    syscall.SIGFPE,
		"hup":    syscall.SIGHUP,
		"ill":    syscall.SIGILL,
		"int":    syscall.SIGINT,
		"io":     syscall.SIGIO,
		"iot":    syscall.SIGIOT,
		"kill":   syscall.SIGKILL,
		"pipe":   syscall.SIGPIPE,
		"prof":   syscall.SIGPROF,
		"quit":   syscall.SIGQUIT,
		"segv":   syscall.SIGSEGV,
		"stop":   syscall.SIGSTOP,
		"sys":    syscall.SIGSYS,
		"term":   syscall.SIGTERM,
		"trap":   syscall.SIGTRAP,
		"tstp":   syscall.SIGTSTP,
		"ttin":   syscall.SIGTTIN,
		"ttou":   syscall.SIGTTOU,
		"urg":    syscall.SIGURG,
		"usr1":   syscall.SIGUSR1,
		"usr2":   syscall.SIGUSR2,
		"vtalrm": syscall.SIGVTALRM,
		"winch":  syscall.SIGWINCH,
		"xcpu":   syscall.SIGXCPU,
		"xfsz":   syscall.SIGXFSZ,
	}

	for name, signal := range signals {
		for _, signame := range []string{name, "sig" + name, strings.ToUpper(name), "SIG" + strings.ToUpper(name)} {
			sigNo, err := parseSignalName(signame)
			assert.Equal(t, sigNo, signal)
			assert.Nil(t, err)
		}
	}
}

func TestUnknownSignalName(t *testing.T) {
	_, err := parseSignalName("WTF")
	assert.NotNil(t, err)
}
