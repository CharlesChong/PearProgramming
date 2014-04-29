package main

import (
	"common"
	"rpc"
	"pear/server"
	"pear/rpc/serverrpc"
)

type centralTester struct {
	srv 		*rpc.Client
	myHostPort 	string
	delay 		float32
}

type testFunc struct {
	name string
	f    func()
}

var (
	portnum   = flag.Int("port", 9019, "port # to listen on")
	passCount 	int
	failCount 	int
	ct 			*centralTester
)

var LOGE = common.LOGE

func initCentralTest(server, myhostport string) (*centralTester,error) {
	tester := new(centralTester)
	tester.myHostPort = myhostport
	// tester.recvRevoke = make(map[string]bool)
	// tester.compRevoke = make(map[string]bool)

	// Create RPC connection to storage server.
	srv, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server %s", server)
	}

	rpc.RegisterName("PearServer", serverrpc.Wrap(tester))
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *portnum))
	if err != nil {
		LOGE.Fatalln("Failed to listen:", err)
	}
	go http.Serve(l, nil)
	tester.srv = srv
	return tester, nil
}

func testAddDoc() {
	passCount++
}

func testRemoveDoc() {
	passCount++
}

func main() {
	tests := []testFunc{
		{"testAddDoc", testAddDoc},
		{"testRemoveDoc", testRemoveDoc},
	}

	flag.Parse()

	// Run the tests with a single tester
	centralTester, err := initCentralTester(flag.Arg(0), fmt.Sprintf("localhost:%d", *portnum))
	if err != nil {
		LOGE.Fatalln("Failed to initialize test:", err)
	}
	st = centralTester
	for _, t := range tests {
		if b, err := regexp.MatchString(*testRegex, t.name); b && err == nil {
			fmt.Printf("Running %s:\n", t.name)
			t.f()
		}
	}
	fmt.Printf("Passed (%d/%d) tests\n", passCount, passCount+failCount)

}