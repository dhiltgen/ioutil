// Get %util from iostat
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
)

var (
	DataFilename     = ".ioutil.data"
	PidFilename      = ".ioutil.pid"
	IntervalFilename = ".ioutil.interval"
)

// Run forever
func backend(dirname string, interval int) {
	args := []string{"-dx", fmt.Sprintf("%d", interval)}
	cmd := exec.Command("iostat", args...)
	rawStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get pipe for iostat: %s", err)
	}
	_ = ioutil.WriteFile(path.Join(dirname, IntervalFilename), []byte(fmt.Sprintf("%d\n", interval)), 0644)
	write := func(util string) {
		_ = ioutil.WriteFile(path.Join(dirname, DataFilename), []byte(util+"\n"), 0644)
		_ = ioutil.WriteFile(path.Join(dirname, PidFilename), []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
	}

	stdout := bufio.NewReader(rawStdout)

	log.Printf("Starting the daemon - view %s for latest snapshot\n", path.Join(dirname, DataFilename))
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start iostat: %s", err)
	}
	for {
		line, err := stdout.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		} else if strings.HasPrefix(line, "Device:") {
			continue
		} else if len(line) > 120 {
			write(strings.TrimSpace(line[strings.LastIndex(line, " "):]))
		}
	}
}

func start(dirname string, interval int) error {
	// TODO
	fmt.Fprintf(os.Stderr, "Starting daemon\n")
	args := []string{"--serve"}
	args = append(args, os.Args[1:]...)
	cmd := exec.Command(os.Args[0], args...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	// Wait for the file to exist...
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)

		_, err := os.Stat(path.Join(dirname, DataFilename))
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("Daemon didn't start up properly")
}

func validateBackend(dirname string, interval int) error {
	_, err := os.Stat(path.Join(dirname, DataFilename))
	// TODO Handle perm problems...
	if err != nil {
		//log.Println("XXX Failed stat")
		return start(dirname, interval)
	}

	// Data file exists, now check for the pid
	data, err := ioutil.ReadFile(path.Join(dirname, PidFilename))
	if err != nil {
		//log.Println("XXX Failed pid read")
		return start(dirname, interval)
	}
	pidString := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidString)
	if err != nil {
		//log.Println("XXX Failed conversion on pid")
		return start(dirname, interval)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		//log.Println("XXX Failed process find")
		return start(dirname, interval)
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return start(dirname, interval)
	}

	// Validate the interval is the right
	data, err = ioutil.ReadFile(path.Join(dirname, IntervalFilename))
	if err != nil {
		//log.Println("XXX Failed interval read")
		process.Signal(syscall.Signal(9))
		return start(dirname, interval)
	}

	oldIntervalString := strings.TrimSpace(string(data))
	oldInterval, err := strconv.Atoi(oldIntervalString)
	if err != nil {
		//log.Println("XXX Failed process interval")
		process.Signal(syscall.Signal(9))
		return start(dirname, interval)
	}
	if oldInterval != interval {
		//log.Println("XXX wrong interval")
		process.Signal(syscall.Signal(9))
		return start(dirname, interval)
	}

	// Everything looks good, continue using it
	return nil
}

func run(dirname string, interval int, color bool) {

	// TODO - validate iostat in the path and fail gracefully if not

	// Look for existing files
	validateBackend(dirname, interval)

	data, err := ioutil.ReadFile(path.Join(dirname, DataFilename))
	if err != nil {
		log.Fatalf("Failed to read data: %s", err)
	}
	percentString := strings.TrimSpace(string(data))
	if !color {
		fmt.Println(percentString + "%")
		return
	}
	percent, err := strconv.ParseFloat(percentString, 64)
	if err != nil {
		fmt.Println(percentString + "%x")
	} else if percent > 50.0 {
		fmt.Printf("<fc=#FF0000>%0.2f%%</fc>\n", percent)
	} else if percent > 25.0 {
		fmt.Printf("<fc=#FFFF00>%0.2f%%</fc>\n", percent)
	} else {
		fmt.Printf("%0.2f%%\n", percent)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "ioutil"
	app.Usage = "Process iostat output to give most recent I/O utilization"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "interval, i",
			Usage: "Seconds between iostat polls",
			Value: 10,
		},
		cli.StringFlag{
			Name:   "dir",
			Usage:  "Directory to place tracking files",
			EnvVar: "HOME",
		},
		cli.BoolFlag{
			Name:  "serve",
			Usage: "Run the server",
		},
		cli.BoolFlag{
			Name:  "color",
			Usage: "Turn on color codes for xmobar",
		},
	}
	app.Action = func(c *cli.Context) {
		if c.Bool("serve") {
			backend(c.String("dir"), c.Int("interval"))
		} else {
			run(c.String("dir"), c.Int("interval"), c.Bool("color"))
		}
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
