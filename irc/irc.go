package irc

import (
	"crypto/tls"
	"github.com/fluffle/goirc/client"
	"io"
	"log"
)

type TeeHandler struct {
	out io.Writer
}

func NewTeeHandler(w io.Writer) *TeeHandler {
	return &TeeHandler{
		out: w,
	}
}

func (*TeeHandler) parseMessage(msg string) (byte, bool) {
	switch {
	case msg == ".":
		// special case: if the *entire* message is ".", interpret as space
		return ' ', true
	case len(msg) == 2 && msg[0] == '.':
		return msg[1], true
	default:
		return ' ', false
	}
}

func (h *TeeHandler) Handle(conn *client.Conn, line *client.Line) {
	target := line.Target()
	text := line.Text()

	log.Printf("received message to %s: %s", target, text)

	command, isCommand := h.parseMessage(text)

	if !isCommand {
		log.Printf("message is not command")
		return
	}

	log.Printf("message is command: writing")

	_, errWrite := h.out.Write([]byte{command})

	if errWrite != nil {
		log.Printf("error writing message: %s", errWrite.Error())
	}
}

type Client struct {
	conn     *client.Conn
	removers []client.Remover
}

func NewClient(nick, server, password, channel string, useSSL bool) *Client {
	config := client.NewConfig(nick)
	config.Server = server
	config.Pass = password

	if useSSL {
		config.SSL = useSSL
		config.SSLConfig = &tls.Config{
			ServerName: server,
		}
	}

	conn := client.Client(config)

	c := &Client{
		conn: conn,
	}

	connectedHandler := func(conn *client.Conn, line *client.Line) {
		config := conn.Config()
		me := conn.Me()

		log.Printf("successfully connected to %s as %s", config.Server, me.Nick)

		log.Printf("joining %s", channel)

		conn.Join(channel)
	}

	c.RegisterHandler(client.CONNECTED, client.HandlerFunc(connectedHandler))

	return c
}

func (c *Client) Connect() error {
	return c.conn.Connect()
}

func (c *Client) Disconnect() {
	c.conn.Quit("bye!")
}

func (c *Client) Close() error {
	log.Printf("removing handlers...")

	for _, r := range c.removers {
		r.Remove()
	}

	log.Printf("...done")

	c.Disconnect()

	return nil
}

func (c *Client) RegisterHandler(name string, h client.Handler) {
	remover := c.conn.Handle(name, h)
	c.removers = append(c.removers, remover)
}
