package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/go-redis/redis"
	"github.com/urfave/cli"
)

const DEFAULT_BUF_SIZE int = 1024

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
					Usage: "redis-server location",
				},
				cli.StringFlag{
					Name:  "redis",
					Value: "127.0.0.1:6379",
					Usage: "redis database location",
				},
				cli.BoolFlag{
					Name:  "silent",
					Usage: "Hides all log information"},
			},
			Action: func(c *cli.Context) error {
				run(c)
				return nil
			},
		},
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "%q not implemented.\n", command)
	}

	_ = app.Run(os.Args)
}

type request struct {
	seat uint32
}

func ParseRequest(buf []byte) (r *request) {

	seat := binary.LittleEndian.Uint32(buf[:4])
	fmt.Printf("Seat: %d\n", seat)

	return &request{
		seat,
	}
}

func handleClient(conn net.Conn, redis *redis.Client, c *cli.Context) {

	// make buffer
	buf := make([]byte, DEFAULT_BUF_SIZE)

	for {

		// read an incomming message
		reqLen, err := conn.Read(buf)
		checkError(err)

		// printing out received message
		fmt.Printf("Receive -> len: %d, message: %s\n", reqLen, buf[:reqLen])

		// parsing request
		request := ParseRequest(buf)
		fmt.Println(request.seat)

		// calling the redis database
		redis.Set("test", request.seat, 0)

		fmt.Println(redis.Get("test"))
	}
}

func run(c *cli.Context) {

	// retreiving flags
	localListeningInterface := c.String("listen")
	redisHost := c.String("redis")

	// printing flags
	fmt.Fprintf(c.App.Writer, "[listen] -> %s\n", c.String("listen"))
	fmt.Fprintf(c.App.Writer, "[redis] -> %s\n", c.String("redis"))

	// setting up redis client
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	// building tcp address to listen to
	tcpAddr, err := net.ResolveTCPAddr("tcp4", localListeningInterface)
	checkError(err)

	// set to listening mode 'tcp'
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {

		// appepts an incomming client
		conn, err := listener.Accept()
		fmt.Printf("Connection appepted: %s\n", conn.LocalAddr().String())
		checkError(err)

		// handle client connection
		go handleClient(conn, client, c)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
