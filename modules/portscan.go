package modules

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os/exec"
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

func (p *Portscan) inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func (p *Portscan) ping(host <-chan string, result chan<- struct {
	string
	bool
}) {
	for ip := range host {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		result <- struct {
			string
			bool
		}{ip, alive}
	}
}
func (p *Portscan) pingResults(pongNum int, pongChan <-chan struct {
	string
	bool
}, doneChan chan<- []struct {
	string
	bool
}) {
	var alives []struct {
		string
		bool
	}
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		if pong.bool {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func (p *Portscan) hostsList(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var ips []string
	for ip = ip.Mask(ipnet.Mask); ipnet.Contains(ip); p.inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips[1 : len(ips)-1], nil
}

func (p *Portscan) searchHosts(cidr string, searchAlive bool) {

	hosts, _ := p.hostsList(cidr)
	fmt.Printf("Searching through a maximum of %d hosts\n", len(hosts))
	maxConcurrent := 100
	pingChan := make(chan string, maxConcurrent /* max concurrent */)

	pongChan := make(chan struct {
		string
		bool
	}, len(hosts))

	doneChan := make(chan []struct {
		string
		bool
	})

	for i := 0; i < maxConcurrent; i++ {
		go p.ping(pingChan, pongChan)
	}
	for _, ip := range hosts {
		pingChan <- ip
	}

	go p.pingResults(len(hosts), pongChan, doneChan)
	alives := <-doneChan

	if searchAlive { /* display alive hosts */
		for _, host := range alives {
			fmt.Println(host.string)
		}
		fmt.Printf("There are a total of %d alive hosts on %s\n", len(alives), cidr)
	} else {
		for _, alivehost := range alives {
			for i, host := range hosts {

				if strings.Compare(alivehost.string, host) == 0 {
					hosts = append(hosts[:i], hosts[i+1:]...)

				}
			}
		}
		for _, unused := range hosts {
			fmt.Println(unused)
		}
		fmt.Printf("There are a total of %d unused hosts on %s\n", len(hosts), cidr)
	}
}
func (p *Portscan) unusedHosts(cidr string) {
	p.searchHosts(cidr, false)
}
func (p *Portscan) aliveHosts(cidr string) {
	p.searchHosts(cidr, true)
}
func (p *Portscan) generatePortScan(host string, prePort string) error {
	ports, err := p.portList(prePort)
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	ch := make(chan string)
	for i := 0; i < int(128); i++ {
		wg.Add(1)
		go p.attacker(
			host,
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

					host := c.Args().Get(0)
					preport := c.Args().Get(1)
					fmt.Println("Scanning: " + c.Args().Get(0) + " " + preport)
					if preport == "" {
						return errors.New("Requires port range to scan")
					}

					return p.generatePortScan(host, preport)
				},
			}, {
				Name:    "unused",
				Aliases: []string{"u"},
				Usage:   "Please provide <CIDR_BLOCK (e.g. 10.0.0.0/16)>",
				Action: func(c *cli.Context) error {
					cidr := c.Args().Get(0)
					if cidr == "" {
						errMessage := "Requires a single argument: <CIDR_BLOCK (e.g. 10.0.0.0/16)>"
						fmt.Println(errMessage)
						return errors.New(errMessage)
					}
					fmt.Printf("Finding unused IP addresses in CIDR_BLOCK [%s]...\n", cidr)

					p.unusedHosts(cidr)
					return nil
				},
			}, {
				Name:    "alive",
				Aliases: []string{"a"},
				Usage:   "Please provide <CIDR_BLOCK (e.g. 10.0.0.0/16)>",
				Action: func(c *cli.Context) error {
					cidr := c.Args().Get(0)
					if cidr == "" {
						errMessage := "Requires a single argument: <CIDR_BLOCK (e.g. 10.0.0.0/16)>"
						fmt.Println(errMessage)
						return errors.New(errMessage)
					}
					fmt.Printf("Finding alive IP addresses in CIDR_BLOCK [%s]...\n", cidr)

					p.aliveHosts(cidr)
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}
