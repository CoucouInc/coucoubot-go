package main

import "strings"
import "fmt"
import "regexp"
import "github.com/thoj/go-ircevent"
import "crypto/tls"

const channel = "#arch-fr-free"
const ssl_server = "irc.freenode.net:7000"

type State struct {
	conn       *irc.Connection
	logs       []string
	substRegex *regexp.Regexp
}

func (self *State) addLog(s string) {
	if len(self.logs) > 10 {
		self.logs = self.logs[1:10]
	}
	self.logs = append(self.logs, s)
}

func (self *State) handleRewrite(ev *irc.Event) {
	s := ev.Message()

	self.conn.Log.Printf("receive %+v\n", ev)

	idx := self.substRegex.FindStringSubmatchIndex(s)
	if idx != nil {
		lhs := s[idx[2]:idx[3]]
		rhs := s[idx[4]:idx[5]]

		for i := len(self.logs) - 1; i >= 0; i -= 1 {
			s2 := self.logs[i]
			if strings.Contains(s2, lhs) {
				replyTo := ev.Arguments[0]
				self.conn.Log.Printf("rewrite `%v` on %s with `%v` -> `%v`\n",
					s2, ev.Nick, lhs, rhs)
				self.conn.Log.Printf("reply to: %+v\n", replyTo)
				s3 := strings.ReplaceAll(s2, lhs, rhs)
				self.addLog(s3)
				self.conn.Privmsg(replyTo, s3)
				break
			}
		}
	}

}

func main() {
	fmt.Println("starting…")
	conn := irc.IRC("coucoubot-go", "coucoubot")
	if conn == nil {
		fmt.Println("could not create IRC object")
		return
	}
	conn.Debug = true
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	conn.AddCallback("001", func(e *irc.Event) { conn.Join(channel) })
	conn.AddCallback("366", func(e *irc.Event) {})

	substRegex := regexp.MustCompile("^s/([^/]*)/(.*)/?$")
	st := State{conn: conn, logs: nil, substRegex: substRegex}

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		conn.Log.Printf("got privmsg (from %v): %v\n", e.Source, e.Message())
		st.handleRewrite(e)
		st.addLog(e.Message()) // after rewriting
	})
	err := conn.Connect(ssl_server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	conn.Loop()
}
