package gbs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GBSFile temp file that stores the modification time of the build script.
var GBSFile string = fmt.Sprintf("%s/gbs_file", os.TempDir())

func storeModTime(buildFile string) (bool, error) {
	fileInfo, err := os.Stat(buildFile)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(GBSFile)
	if os.IsNotExist(err) {
		if err := os.WriteFile(GBSFile, []byte(fileInfo.ModTime().String()), 0666); err != nil {
			return false, err
		}
		return false, nil
	}

	data, err := os.ReadFile(GBSFile)
	if err != nil {
		return false, err
	}

	sp := strings.Split(string(data), " ")
	strTime := strings.Join(sp[:len(sp)-2], " ")
	t, err := time.Parse("2006-01-02 15:04:05", strTime)
	if err != nil {
		return false, err
	}

	if !fileInfo.ModTime().Equal(t) {
		if err := os.WriteFile(GBSFile, []byte(fileInfo.ModTime().String()), 0666); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// BuilderFunc type that represents the build function that we pass into Build
// function.
type BuilderFunc func() error

// BuildFuncOpt build options.
type BuildFuncOpt struct {
	FuncName string
	Func     BuilderFunc
}

// GoBuildYourSelf checks buildFile modification time and decide whether to
// re-build and run the script or not.
func GoBuildYourSelf(buildFile string) (bool, error) {
	ok, err := storeModTime(buildFile)
	if err != nil {
		return false, err
	}
	if ok {
		var command string = fmt.Sprintf("go build -o %s %s", os.Args[0], buildFile)
		if err := new(Sh).Init(command).Run().Error(); err != nil {
			return false, err
		}
		if err := new(Sh).Init(fmt.Sprintf("./%s", os.Args[0])).Run().Error(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Build accept BuildFuncOpt and retuns an error if any of the BuilderFunc has
// failed otherwise nil is returned.
func Build(opts ...BuildFuncOpt) error {
	for _, opt := range opts {
		log.Printf("Executing %s", opt.FuncName)
		if err := opt.Func(); err != nil {
			return fmt.Errorf("Failed to execute %s [%w]", opt.FuncName, err)
		}
	}
	return nil
}

// Builder function `Build`
type Builder func(...BuildFuncOpt) error

// LiveFiles
type LiveFile struct {
	Name    string
	ModTime time.Time
}

// LiveBuild live build your project by tracking all go files and check
// their modification time. then run the builder function once any go file
// has been changed.
func LiveBuild(dir string, builder Builder, cancel chan struct{}) error {

	var goFiles []LiveFile
	err := searchGoFiles(dir, &goFiles)
	if err != nil {
		return err
	}

	ping := make(chan struct{})
	stop := make(chan bool)

	defer close(ping)
	defer close(stop)

	var errG error
	go func() {
		for {
			select {
			case <-cancel:
				stop <- true
				return
			case <-ping: 
				if err := builder(); err != nil {
					errG = err
					stop <- true
					return
				}
			}
		}
	}()

	queue := goFilesToQueue(goFiles)
	if err := runGoFilesTrackers(queue, ping, stop); err != nil {
		return err
	}
	return errG
}

func searchGoFiles(dir string, q *[]LiveFile) error {
	const goExt string = ".go"
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			dirPath := filepath.Join(dir, f.Name())
			searchGoFiles(dirPath, q)
			continue
		}
		if filepath.Ext(f.Name()) == goExt {
			info, err := f.Info()
			if err != nil {
				return err
			}
			liveFile := LiveFile{
				Name:    filepath.Join(dir, f.Name()),
				ModTime: info.ModTime(),
			}
			*q = append(*q, liveFile)
		}
	}
	return nil
}

func goFilesToQueue(files []LiveFile) <-chan LiveFile {
	out := make(chan LiveFile)
	go func() {
		defer close(out)
		for _, f := range files {
			out <- f
		}
	}()

	return out
}

func runGoFilesTrackers(c <-chan LiveFile, out chan struct{}, cancel chan bool) error {

	errC := make(chan error)

	for f := range c {
		go func(file LiveFile) {
			const timeout = time.Second * 1
			t := time.NewTimer(timeout)
			mod := file.ModTime
			for {
				fInfo, err := os.Stat(file.Name)
				if err != nil {
					errC <- err
					return
				}
				if !mod.Equal(fInfo.ModTime()) {
					out <- struct{}{}
					mod = fInfo.ModTime()
				}
				<-t.C
				t.Reset(timeout)
			}
		}(f)
	}


	select {
	case <-cancel:
		return nil
	case err := <-errC:
		if err != nil {
			return err
		}
	}
	return nil
}

// Sh structure that represents a shell that runs shell commands.
type Sh struct {
	cmd *exec.Cmd

	err error
}

// Init initialize Sh shell. by passing the command that you are going to run.
func (s *Sh) Init(command string) *Sh {
	sp := strings.Split(command, " ")
	var (
		root string
		args []string
	)
	for i, n := range sp {
		switch i {
		case 0:
			root = n
		case 1:
			args = sp[i:]
			break
		}
	}
	cmd := exec.Command(root, args...)
	s.cmd = cmd
	return s
}

// Run Sh method that runs the command provided in Init method.
func (s *Sh) Run() *Sh {
	if s.cmd == nil {
		s.err = fmt.Errorf("sh is not initialized please use Init method.")
		return s
	}

	s.cmd.Stdout = os.Stdout
	s.cmd.Stdin = os.Stdin
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		s.err = err
		return s
	}

	if err := s.cmd.Wait(); err != nil {
		s.err = err
	}

	return s
}

// In accept run command and pass the `inCommand` param into the program via
// stdin.
func (s *Sh) In(inCommand string) *Sh {
	writer, err := s.cmd.StdinPipe()
	if err != nil {
		s.err = err
		return s
	}

	isExec := make(chan struct{})
	done := make(chan bool)

	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	defer close(isExec)
	defer close(done)

	go func() {
		<-isExec
		if err := s.cmd.Wait(); err != nil {
			s.err = err
		}
	}()

	go func() {
		if err := s.cmd.Start(); err != nil {
			s.err = err
			done <- true
			return
		}
		isExec <- struct{}{}
		if _, err := writer.Write([]byte(inCommand)); err != nil {
			s.err = err
		}
		done <- true
	}()

	<-done
	return s
}

// Error returns any error that occurs during the process of running the
// command with Sh.
func (s *Sh) Error() error {
	return s.err
}
