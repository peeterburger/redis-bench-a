package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/go-redis/redis"
	"github.com/urfave/cli"
)

const defaultBufferSize int = 1024

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

func parseRequest(buf []byte) (r *request) {

	seat := binary.LittleEndian.Uint32(buf[:4])
	fmt.Printf("Seat: %d\n", seat)

	return &request{
		seat,
	}
}

func handleClient(conn net.Conn, redis *redis.Client, c *cli.Context) {

	// make buffer
	buf := make([]byte, defaultBufferSize)

	for {

		// read an incomming message
		reqLen, err := conn.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Closing connection from %s", conn.LocalAddr().String())
			break
		}

		// printing out received message
		fmt.Printf("Receive -> len: %d, message: %s\n", reqLen, buf[:reqLen])

		// parsing request
		request := parseRequest(buf)
		fmt.Println(request.seat)

		// calling the redis database
		redis.Set("test", request.seat, 0)

		fmt.Println(redis.Get("test"))
	}

	conn.Close()
}

func run(c *cli.Context) {

	// retreiving flags
	localListeningInterface := c.String("listen")
	redisHost := c.String("redis")

	// printing flags
	fmt.Fprintf(c.App.Writer, "[listen] -> %s\n", c.String("listen"))
	fmt.Fprintf(c.App.Writer, "[redis] -> %s\n", c.String("redis"))

	// setting up redis client
	fmt.Fprintf(c.App.Writer, "Creating redis client (%s)\n", redisHost)
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	// testing connection to redis host
	fmt.Fprintf(c.App.Writer, "Testing connection to redis host (%s)\n", redisHost)
	pong, err := client.Ping().Result()
	checkError(err)
	if pong != "PONG" {
		panic("Ping request did not answer 'PONG'")
	}

	// building tcp address to listen to
	fmt.Fprintf(c.App.Writer, "Setting up local iterface (%s)\n", localListeningInterface)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", localListeningInterface)
	checkError(err)

	// set to listening mode 'tcp'
	fmt.Fprintf(c.App.Writer, "Setting listening mode 'tcp' (%s)\n", localListeningInterface)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	fmt.Fprintf(c.App.Writer, "Waiting for incomming connections (%s)\n", localListeningInterface)

	for {

		// appepts an incomming client
		conn, err := listener.Accept()
		fmt.Printf("Connection accepted from %s\n", conn.LocalAddr().String())
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
