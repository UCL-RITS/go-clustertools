package rc

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

func RunCommand(command string, args []string, env []string) (string, string, error) {
	cmd := exec.Command(command, args...)

	cmd.Env = append(os.Environ(), env...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", "", err
	}

	// We need to read from both in a non-blocking way, because waiting for either stderr or stdout
	//  output to be read can block program execution
	var waitGroup sync.WaitGroup
	var stdoutBytes []byte
	var stderrBytes []byte
	var stdoutString string
	var stderrString string
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		stdoutBytes, _ = io.ReadAll(stdout)
	}()
	go func() {
		defer waitGroup.Done()
		stderrBytes, _ = io.ReadAll(stderr)
	}()
	waitGroup.Wait()

	stdoutString = string(stdoutBytes)
	stderrString = string(stderrBytes)

	// Now that all the reading has finished, we can wait for the program to finish
	err = cmd.Wait()

	// err in this case also contains info about non-zero exit status -- apparently there's no other way to get those :(
	return stdoutString, stderrString, err
}

//func main() {
//	sout, serr, err := runCommand("ls", []string{"hurgleburgle"}, []string{})
//	fmt.Printf("stdout: %s\n\nstderr: %s\n\nerr: %s\n", sout, serr, err)
//}
