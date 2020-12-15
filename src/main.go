package main

//a
import "io/ioutil"
import "strings"
import "fmt"
import "os"

func g() {
	ts := TypeStmt{}
	pnode(ts)
}

func f2() {
	// rd, _ := ioutil.ReadDir(os.Args[1])
	//  for _, ofi := range rd {

}

func main() {
	if len(os.Args) <= 1 {
		os.Exit(1)
	}

	src, err := ioutil.ReadFile(os.Args[1] + ".pinn")
	if err != nil {
		panic(err)
	}
	ssrc := string(src)
	if len(os.Args) == 3 && os.Args[2] == "scan" {
		s := new(scan)
		s.init(strings.NewReader(ssrc))
		fmt.Println(s.tokenize())
		return

	}
	p := new(parser)
	p.init(strings.NewReader(ssrc))

	f := p.fileA()
	if len(os.Args) > 2 {
		if os.Args[2] == "x86_64" {
			L = true
		} else if os.Args[2] == "parse" {
			return
		} else if os.Args[2] == "visit" {

			visitFile(f)
			return
		}
	}
	e := emitter{}
	e.init(f)
	e.emitF()
	fmt.Println(e.p.sb.String())
	return
}
