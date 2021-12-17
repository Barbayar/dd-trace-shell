package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	ps "github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type span struct {
	ddtrace.Span
	context.Context
	finished bool
}

var interpreters = map[string]struct{}{
	"/bin/bash": {},
}

func cmdLine2ServiceName(cmdLine []string) (interpreter, service, args string) {
	if _, ok := interpreters[cmdLine[0]]; ok {
		interpreter = cmdLine[0]
		cmdLine = cmdLine[1:]
	}
	servicePath := strings.Split(cmdLine[0], "/")
	service = servicePath[len(servicePath)-1]
	args = strings.Join(cmdLine[1:], " ")

	return
}

func main() {
	tracer.Start()

	var wg sync.WaitGroup
	pIDs := make(map[int]span)
	pIDs[os.Getpid()] = span{nil, context.Background(), false}

	go func() {
		for true {
			processes, err := ps.Processes()
			if err != nil {
				continue
			}
			runningPIDs := make(map[int]struct{}, 2*len(processes))

			for i := range processes {
				pid := processes[i].Pid()
				ppid := processes[i].PPid()
				runningPIDs[pid] = struct{}{}

				if _, ok := pIDs[ppid]; ok {
					parentCtx := pIDs[processes[i].PPid()].Context

					if _, ok := pIDs[pid]; !ok {
						p, err := process.NewProcess(int32(pid))
						if err != nil {
							continue
						}
						cmdLine, err := p.CmdlineSlice()
						if err != nil {
							continue
						}

						wg.Add(1)
						interpreter, service, args := cmdLine2ServiceName(cmdLine)
						spn, ctx := tracer.StartSpanFromContext(parentCtx, service)
						spn.SetTag(ext.ServiceName, service)
						spn.SetTag("pid", pid)
						if interpreter != "" {
							spn.SetTag("interpreter", interpreter)
						}
						if args != "" {
							spn.SetTag("args", args)
						}
						pIDs[pid] = span{spn, ctx, false}
					}
				}
			}

			for pID := range pIDs {
				if _, ok := runningPIDs[pID]; !ok && !pIDs[pID].finished {
					wg.Done()
					pIDs[pID].Span.Finish()
					pIDs[pID] = span{pIDs[pID].Span, pIDs[pID].Context, true}
				}
			}
		}
	}()

	if len(os.Args) < 2 {
		fmt.Printf("usage: %s command [ARGS]\n", os.Args[0])
		return
	}

	command := os.Args[1]
	args := os.Args[2:]
	cmd := exec.Command(command, args...)
	cmd.Run()
	fmt.Println("waiting for all processes to be finished...")
	wg.Wait()
	fmt.Println("waiting for 5 seconds to give the tracer a time to flush...")
	time.Sleep(5 * time.Second)
}
