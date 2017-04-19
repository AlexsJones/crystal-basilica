package modules

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"sync/atomic"

	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"
)

var (
	slogger *log.Logger
	start   = time.Now()
	nSuc    uint64
)

const (
	//RETRYWAIT Wait at most this between retries
	RETRYWAIT = time.Second * 30
	//PROGINTERVAL interval
	PROGINTERVAL = time.Second * 15
)

//Portscan struct todo
type Portscan struct {
}

//Init portscan members
func (*Portscan) Init() {

}
func attacker(
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
		n, err = attackOne(t, buf, to)
		if nil != err {
			emsg = err.Error()
		}
		/* Workarounds */
		if nil != err && retry &&
			strings.HasSuffix(emsg, "connect: no route to host") {
			/* Sleep some amount of time */
			bst, eg := rand.Int(
				rand.Reader,
				big.NewInt(RETRYWAIT.Nanoseconds()),
			)
			if nil != eg {
				log.Fatalf(
					"Unable to make retry time: %v",
					err,
				)
			}
			st := time.Duration(bst.Uint64()) * time.Nanosecond
			/* Seems unnecessarily noisy
			log.Printf(
				"INFO Will retry port %v in %v due to "+
					"\"connect: no route to host\" error",
				port,
				st,
			)
			*/
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
		slog(t, buf)
	}
}

/* attackOne tries to banner t, which must be a host:port pair.  It'll log
successful connects and banner grabs.  buf is the read buffer, which will be
populated if nil is returned and a banner was sent back.  If so, the number
of bytes read is also returned. */
func attackOne(t string, buf []byte, to time.Duration) (int, error) {
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
func portList(rs string) ([]string, error) {
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
func (*Portscan) LoadFlags() []cli.Command {

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
					ports, err := portList(preport)
					if err != nil {
						return err
					}
					slogger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

					wg := &sync.WaitGroup{}
					ch := make(chan string)
					for i := 0; i < int(128); i++ {
						wg.Add(1)
						go attacker(
							c.Args().Get(0),
							ch,
							false,
							false,
							time.Second,
							int(128),
							wg,
						)
					}

					lastt := time.Now()
					lasti := 0
					for i, p := range ports {
						ch <- p
						/* Log progress every so often */
						if time.Now().After(lastt.Add(PROGINTERVAL)) {
							n := time.Now()
							/* Time this interval */
							itime := n.Sub(lastt)
							/* Ports per minute */
							ppm := float64(i-lasti) / n.Sub(lastt).Minutes()
							/* Estimated time remaining */
							est := "forever"
							etc := "never"
							if lasti < i {
								/* Duration per port */
								dpp := itime / time.Duration(i-lasti)
								rem := time.Duration(len(ports)-i) * dpp
								est = fmt.Sprintf("%v", rem)
								etc = n.Add(rem).Format("15:04:05")
							}
							log.Printf(
								"INFO Working on port %v/%v "+
									"(%0.2f ports/min, "+
									"%v open, "+
									"%4v remaining, "+
									"est. completion: %v)",
								i+1,
								len(ports),
								ppm,
								nSuc,
								est,
								etc,
							)
							lastt = n
							lasti = i
						}
					}
					close(ch)
					log.Printf("INFO Waiting for the attackers to finish")

					/* Wait for attackers to finish */
					wg.Wait()
					d := time.Now().Sub(start)
					log.Printf(
						"INFO Scanned %v ports in %v (%0.2f ports per minute), "+
							"found %v open",
						len(ports),
						d,
						float64(len(ports))/d.Minutes(),
						nSuc,
					)

					log.Printf("Done.")
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}

/* slog logs success */
func slog(t string, buf []byte) {
	atomic.AddUint64(&nSuc, 1)
	slogger.Printf("SUCCESS %v %q", t, buf)
}
