package modules

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"
)

//Portscan struct todo
type Portscan struct {
}

func (p *Portscan) attacker(
	target string,
	ports <-chan string,
	fails bool,
	retry bool,
	to time.Duration,
	blen int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	var (
		buf  = make([]byte, blen)
		err  error
		emsg string
		t    string
		n    int
	)
	/* Attack each received port */
	for port := range ports {
		t = net.JoinHostPort(target, port)
	try:
		/* Reset buffer */
		buf = buf[:cap(buf)]
		err = nil
		emsg = ""
		n, err = p.attackOne(t, buf, to)
		if nil != err {
			emsg = err.Error()
		}
		/* Workarounds */
		if nil != err && retry &&
			strings.HasSuffix(emsg, "connect: no route to host") {
			/* Sleep some amount of time */
			bst, eg := rand.Int(
				rand.Reader,
				big.NewInt((time.Second * 30).Nanoseconds()),
			)
			if nil != eg {
				log.Fatalf(
					"Unable to make retry time: %v",
					err,
				)
			}
			st := time.Duration(bst.Uint64()) * time.Nanosecond

			time.Sleep(st)
			goto try /* Neener neener */
		}
		/* Log other errors if asked */
		if nil != err &&
			(strings.HasSuffix(
				emsg,
				"i/o timeout",
			) ||
				strings.HasSuffix(
					emsg,
					": connection refused",
				)) {
			if fails {
				log.Printf("FAIL %v %v", t, err)
			}
			continue
		}
		if nil != err {
			log.Printf("ERROR %v %v", t, err)
			continue
		}
		buf = buf[:n]
		fmt.Println(t)
	}
}

/* attackOne tries to banner t, which must be a host:port pair.  It'll log
successful connects and banner grabs.  buf is the read buffer, which will be
populated if nil is returned and a banner was sent back.  If so, the number
of bytes read is also returned. */
func (p *Portscan) attackOne(t string, buf []byte, to time.Duration) (int, error) {
	/* Try to connect */
	c, err := net.DialTimeout("tcp", t, to)
	if nil != err {
		return 0, err
	}
	defer c.Close()
	/* Banner-grab */
	if err := c.SetReadDeadline(time.Now().Add(to)); nil != err {
		return 0, err
	}
	n, _ := c.Read(buf)
	return n, nil
}

func (p *Portscan) portList(rs string) ([]string, error) {
	ns := make(map[int]struct{})

	for _, r := range strings.Split(rs, ",") {
		/* Ignore empty ranges */
		if "" == r {
			continue
		}
		/* If it's a single port, add it */
		if !strings.Contains(r, "-") {
			n, err := strconv.Atoi(r)
			if nil != err {
				return nil, err
			}
			ns[n] = struct{}{}
			continue
		}

		/* It must be a range, get the start and end */
		bounds := strings.Split(r, "-")
		if 2 != len(bounds) {
			return nil, fmt.Errorf(
				"port range not two numbers separated by a " +
					"hyphen",
			)
		}
		if "" == bounds[0] {
			return nil, fmt.Errorf("missing lower bound")
		}
		start, err := strconv.Atoi(bounds[0])
		if nil != err {
			return nil, err
		}
		if "" == bounds[1] {
			return nil, fmt.Errorf("missing upper bound")
		}
		end, err := strconv.Atoi(bounds[1])
		if nil != err {
			return nil, err
		}
		for i := start; i <= end; i++ {
			ns[i] = struct{}{}
		}
	}
	/* Slice of ports to scan */
	ps := make([]string, 0, len(ns))
	for n := range ns {
		ps = append(ps, fmt.Sprintf("%v", n))
	}
	/* Shuffle ports */
	for i := range ps {
		ri, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(i)+1),
		)
		if nil != err {
			return nil, err
		}
		j := int(ri.Uint64())
		ps[i], ps[j] = ps[j], ps[i]
	}
	return ps, nil
}

//LoadFlags for cli
func (p *Portscan) LoadFlags() []cli.Command {

	var commands []cli.Command = make([]cli.Command, 0)
	n := cli.Command{
		Name:    "portscan",
		Aliases: []string{"p"},
		Usage:   "options for task templates",
		Subcommands: []cli.Command{
			{
				Name:    "scan",
				Aliases: []string{"s"},
				Usage:   "Please provide <HOSTNAME (e.g. 10.65.1.0)> <PORT-RANGE (e.g. 1-1000)>",
				Action: func(c *cli.Context) error {

					preport := c.Args().Get(1)
					fmt.Println("Scanning: " + c.Args().Get(0) + " " + preport)
					if preport == "" {
						return errors.New("Requires port range to scan")
					}
					ports, err := p.portList(preport)
					if err != nil {
						return err
					}

					wg := &sync.WaitGroup{}
					ch := make(chan string)
					for i := 0; i < int(128); i++ {
						wg.Add(1)
						go p.attacker(
							c.Args().Get(0),
							ch,
							false,
							false,
							time.Second,
							int(128),
							wg,
						)
					}
					for _, dp := range ports {
						ch <- dp
					}

					close(ch)
					log.Printf("Waiting for the attackers to finish")
					wg.Wait()
					log.Printf("Done.")
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}
