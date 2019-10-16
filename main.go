package main

import (
	"errors"
	"os"

	"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"

	"github.com/al1img/lwm2mTestServer/bootstrap"
)

/*******************************************************************************
 * Variables
 ******************************************************************************/

var errWrongArgCount = errors.New("wrong argument count")

/*******************************************************************************
 * Init
 ******************************************************************************/

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		FullTimestamp:    true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

/*******************************************************************************
 * Main
 ******************************************************************************/

func main() {
	shell := ishell.New()

	shell.Println("lwm2m Test Server")

	b := bootstrap.New(":5685")
	b.Start()

	bootstrapCmd := &ishell.Cmd{Name: "bootstrap", Help: "bootstrap commands"}

	bootstrapCmd.AddCmd(&ishell.Cmd{
		Name: "discover",
		Help: "bootstrap discover <client> <path>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return b.GetClients()
			}
			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 2 {
				context.Err(errWrongArgCount)
				return
			}
			result, err := b.Discover(context.Args[0], context.Args[1])
			if err != nil {
				context.Err(err)
				return
			}
			context.Println(result)
		}})

	bootstrapCmd.AddCmd(&ishell.Cmd{
		Name: "read",
		Help: "bootstrap read <client> <path>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return b.GetClients()
			}
			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 2 {
				context.Err(errWrongArgCount)
				return
			}
			result, err := b.Read(context.Args[0], context.Args[1])
			if err != nil {
				context.Err(err)
				return
			}
			context.Println(result)
		}})

	bootstrapCmd.AddCmd(&ishell.Cmd{
		Name: "write",
		Help: "bootstrap write <client> <path> <data>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return b.GetClients()
			}
			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 3 {
				context.Err(errWrongArgCount)
				return
			}
			err := b.Write(context.Args[0], context.Args[1], []byte(context.Args[2]))
			if err != nil {
				context.Err(err)
				return
			}
		}})

	bootstrapCmd.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "bootstrap delete <client> <path>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return b.GetClients()
			}
			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 2 {
				context.Err(errWrongArgCount)
				return
			}
			err := b.Delete(context.Args[0], context.Args[1])
			if err != nil {
				context.Err(err)
				return
			}
		}})

	bootstrapCmd.AddCmd(&ishell.Cmd{
		Name: "finish",
		Help: "bootstrap finish <client>",
		Completer: func(args []string) []string {
			if len(args) == 0 {
				return b.GetClients()
			}
			return []string{}
		},
		Func: func(context *ishell.Context) {
			if len(context.Args) != 1 {
				context.Err(errWrongArgCount)
				return
			}
			err := b.Finish(context.Args[0])
			if err != nil {
				context.Err(err)
				return
			}
		}})

	shell.AddCmd(bootstrapCmd)

	shell.Run()
	shell.Close()
}
