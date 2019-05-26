package main

import (
	"fmt"
	"net"
	"os"

	"github.com/go-redis/redis"
	"github.com/urfave/cli"
)

func main() {

	var app = cli.NewApp()

	app.Name = "redis-bench-server"
	app.Usage = "The server side that handles banchmarking requests"
	app.Author = "Peter Burger"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "run",
			Usage: "Runs the server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen",
					Value: "127.0.0.1:7000",
					Usage: "redis-server location - (<host>:<port>)",
				},
				cli.StringFlag{
					Name:  "redis",
					Value: "127.0.0.1:6379",
					Usage: "redis-server location - (<host>:<port>)",
				},
				cli.BoolFlag{
					Name:  "silent",
					Usage: "Hides all log information"},
			},
			Action: func(c *cli.Context) error {

				fmt.Fprintf(c.App.Writer, "Evaluating flags...\n")

				fmt.Fprintf(c.App.Writer, "[host] -> %s\n", c.String("listen"))
				listen := c.String("listen")

				fmt.Fprintf(c.App.Writer, "[redis] -> %s\n", c.String("redis"))
				redisHost := c.String("redis")

				// fmt.Fprintf(c.App.Writer, "[silent] -> %d\n", c.Bool("silent"))
				// silent := c.Bool("silent")

				client := redis.NewClient(&redis.Options{
					Addr:     redisHost,
					Password: "",
					DB:       0,
				})

				tcpAddr, err := net.ResolveTCPAddr("tcp4", listen)
				checkError(err)

				listener, err := net.ListenTCP("tcp", tcpAddr)
				checkError(err)

				for {
					conn, err := listener.Accept()
					fmt.Printf("Connection appepted: %s\n", conn.LocalAddr().String())

					if err != nil {
						continue
					}

					go handleClient(conn, client)
				}
			},
		},
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "%q not implemented.\n", command)
	}

	_ = app.Run(os.Args)
}

func handleClient(conn net.Conn, redis *redis.Client) {
	buf := make([]byte, 1024)

	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
		}

		fmt.Printf("Receive -> len: %d, message: %s\n", reqLen, buf[:reqLen])

		redis.Set(string(buf[:reqLen]), string(buf[:reqLen]), 0)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
