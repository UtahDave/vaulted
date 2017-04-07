package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/miquella/vaulted/lib"
)

var (
	envFishOutput = `# To load these variables into your shell, execute:
#   eval (vaulted env one)
set -x ONE "111111";
set -x THREE "333";
set -x TWO "222";
`
	envFishOutputWithPermCreds = `# To load these variables into your shell, execute:
#   eval (vaulted env one)
set -e AWS_SECURITY_TOKEN;
set -e AWS_SESSION_TOKEN;
set -x AWS_ACCESS_KEY_ID "aws-key-id";
set -x AWS_SECRET_ACCESS_KEY "aws-secret-key";
set -x ONE "111111";
set -x THREE "333";
set -x TWO "222";
`

	envShOutput = `# To load these variables into your shell, execute:
#   eval $(vaulted env one)
export ONE="111111"
export THREE="333"
export TWO="222"
`
	envShOutputWithPermCreds = `# To load these variables into your shell, execute:
#   eval $(vaulted env one)
unset AWS_SECURITY_TOKEN
unset AWS_SESSION_TOKEN
export AWS_ACCESS_KEY_ID="aws-key-id"
export AWS_SECRET_ACCESS_KEY="aws-secret-key"
export ONE="111111"
export THREE="333"
export TWO="222"
`

	envJSONOutput = `{
  "ONE": "111111",
  "THREE": "333",
  "TWO": "222"
}
`
	envCustom = "[AWS_SECURITY_TOKEN AWS_SESSION_TOKEN]"
)

func TestEnv(t *testing.T) {
	steward := NewTestSteward()
	steward.Vaults["one"] = &vaulted.Vault{
		Vars: map[string]string{
			"TWO":   "222",
			"ONE":   "111111",
			"THREE": "333",
		},
	}

	output := CaptureStdout(func() {
		e := Env{
			VaultName:     "one",
			DetectedShell: "fish",
			Format:        "shell",
			Command:       "vaulted env one",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})
	if string(output) != envFishOutput {
		t.Error(failureMessage(envFishOutput, output))
	}

	output = CaptureStdout(func() {
		e := Env{
			VaultName:     "one",
			DetectedShell: "sh",
			Format:        "shell",
			Command:       "vaulted env one",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})
	if string(output) != envShOutput {
		t.Error(failureMessage(envShOutput, output))
	}

	output = CaptureStdout(func() {
		args := os.Args
		os.Args = []string{"vaulted", "env", "one", "--format", "json"}
		defer func() { os.Args = args }()

		e := Env{
			VaultName:     "one",
			DetectedShell: "sh",
			Format:        "json",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})
	if string(output) != envJSONOutput {
		t.Error(failureMessage(envJSONOutput, output))
	}

	// cached environment
	steward.Environments["one"] = &vaulted.Environment{
		AWSCreds: &vaulted.AWSCredentials{
			ID:     "aws-key-id",
			Secret: "aws-secret-key",
		},
		Vars: map[string]string{
			"TWO":   "222",
			"ONE":   "111111",
			"THREE": "333",
		},
	}

	output = CaptureStdout(func() {
		e := Env{
			VaultName:     "one",
			DetectedShell: "fish",
			Format:        "fish",
			Command:       "vaulted env one",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})
	if string(output) != envFishOutputWithPermCreds {
		t.Error(failureMessage(envFishOutputWithPermCreds, output))
	}

	output = CaptureStdout(func() {
		e := Env{
			VaultName:     "one",
			DetectedShell: "sh",
			Format:        "shell",
			Command:       "vaulted env one",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})

	if string(output) != envShOutputWithPermCreds {
		t.Error(failureMessage(envShOutputWithPermCreds, output))
	}

	output = CaptureStdout(func() {
		args := os.Args
		os.Args = []string{"vaulted", "env", "one"}
		defer func() { os.Args = args }()

		e := Env{
			VaultName: "one",
			Format:    "{{ .Unset }}",
		}
		err := e.Run(steward)
		if err != nil {
			t.Error(err)
		}
	})

	if string(output) != envCustom {
		t.Error(failureMessage(envCustom, output))
	}

}

func failureMessage(expected string, got []byte) string {
	return fmt.Sprintf("Incorrect output!\nExpected:\n\"%s\"\ngot:\n\"%s\"", expected, got)
}
