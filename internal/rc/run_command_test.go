package rc

import "testing"

func TestRunCommand(t *testing.T) {
	stdout, _, err := runCommand("/bin/echo", []string{"beep", "boop"}, []string{})

	if stdout != "beep boop\n" {
		t.Errorf("expected output 'beep boop', got %s (err: %s)", stdout, err)
	}

	stdout, _, err = runCommand("/hurgleblurgle", []string{}, []string{})

	if err == nil {
		t.Errorf("expected failure on nonexistent command, got success?")
	}
}
