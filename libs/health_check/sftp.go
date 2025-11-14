package health_check

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	utility "health-check-on-go/libs/utility"
)

type SFTPServices struct {
	Servers []Server
}

type Server struct {
	Name            string
	Host            string
	Port            int
	Username        string
	Password        string
	PrivateKeyPath  string
	Passphrase      string
	Timeout         time.Duration
	HostKeyCallback ssh.HostKeyCallback
	Commands        []Command
}

type CommandMode string

const (
	CommandModeConnect CommandMode = "connect"
	CommandModeStat    CommandMode = "stat"
	CommandModeRead    CommandMode = "read"
	CommandModeList    CommandMode = "list"
)

type CommandFunc func(*sftp.Client) (string, error)

type Command struct {
	Name     string
	Path     string
	Mode     CommandMode
	MaxBytes int
	Action   CommandFunc
	Asserts  []Assert
}

type SFTPAssert struct {
	Fn          func(body string) bool
	Description string
}

type SFTPResult struct {
	Name     string
	Mode     CommandMode
	Path     string
	Output   string
	Err      error
	Duration time.Duration
	Asserts  []map[string]bool
}

type serverTestResult struct {
	server  string
	results []SFTPResult
}

func (services *SFTPServices) SetServer(server Server) {
	services.Servers = append(services.Servers, server)
}

func (s *Server) SetCommand(cmd Command) {
	s.Commands = append(s.Commands, cmd)
}

func (s Server) displayName() string {
	if s.Name != "" {
		return s.Name
	}
	if s.Host != "" {
		return s.address()
	}
	return "sftp-server"
}

func (s Server) address() string {
	if strings.Contains(s.Host, ":") {
		return s.Host
	}
	port := s.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s:%d", s.Host, port)
}

func (s Server) authMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if s.Password != "" {
		methods = append(methods, ssh.Password(s.Password))
	}
	if s.PrivateKeyPath != "" {
		signer, err := loadPrivateKey(s.PrivateKeyPath, s.Passphrase)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("no authentication method provided for %s", s.displayName())
	}
	return methods, nil
}

func loadPrivateKey(path string, passphrase string) (ssh.Signer, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(data, []byte(passphrase))
	}
	return ssh.ParsePrivateKey(data)
}

func (s Server) sshConfig() (*ssh.ClientConfig, error) {
	methods, err := s.authMethods()
	if err != nil {
		return nil, err
	}
	if s.Username == "" {
		return nil, fmt.Errorf("username is required for %s", s.displayName())
	}
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	callback := s.HostKeyCallback
	if callback == nil {
		callback = ssh.InsecureIgnoreHostKey()
	}
	return &ssh.ClientConfig{
		User:            s.Username,
		Auth:            methods,
		HostKeyCallback: callback,
		Timeout:         timeout,
	}, nil
}

func (s Server) newClient() (*sftp.Client, *ssh.Client, time.Duration, error) {
	config, err := s.sshConfig()
	if err != nil {
		return nil, nil, 0, err
	}
	start := time.Now()
	sshClient, err := ssh.Dial("tcp", s.address(), config)
	duration := time.Since(start)
	if err != nil {
		return nil, nil, duration, err
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, duration, err
	}
	return client, sshClient, duration, nil
}

func (cmd Command) modeOrDefault() CommandMode {
	if cmd.Mode == "" {
		return CommandModeStat
	}
	return cmd.Mode
}

func (cmd Command) displayName() string {
	if cmd.Name != "" {
		return cmd.Name
	}
	if cmd.Path != "" {
		return fmt.Sprintf("%s %s", cmd.modeOrDefault(), cmd.Path)
	}
	if cmd.Action != nil {
		return "custom command"
	}
	return string(cmd.modeOrDefault())
}

func (cmd Command) execute(client *sftp.Client) SFTPResult {
	result := SFTPResult{
		Name: cmd.displayName(),
		Mode: cmd.modeOrDefault(),
		Path: cmd.Path,
	}
	start := time.Now()
	output, err := cmd.run(client)
	result.Duration = time.Since(start)
	if err != nil {
		result.Err = err
		return result
	}
	result.Output = output
	for _, assert := range cmd.Asserts {
		success := assert.Fn(output)
		result.Asserts = append(result.Asserts, map[string]bool{assert.Description: success})
	}
	return result
}

func (cmd Command) run(client *sftp.Client) (string, error) {
	if cmd.Action != nil {
		return cmd.Action(client)
	}
	switch cmd.modeOrDefault() {
	case CommandModeRead:
		return cmd.readFile(client)
	case CommandModeList:
		return cmd.listDirectory(client)
	default:
		return cmd.statPath(client)
	}
}

func (cmd Command) readFile(client *sftp.Client) (string, error) {
	if cmd.Path == "" {
		return "", fmt.Errorf("path is required for read command")
	}
	f, err := client.Open(cmd.Path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	maxBytes := cmd.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 4096
	}
	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return string(buf[:n]), nil
}

func (cmd Command) listDirectory(client *sftp.Client) (string, error) {
	if cmd.Path == "" {
		return "", fmt.Errorf("path is required for list command")
	}
	entries, err := client.ReadDir(cmd.Path)
	if err != nil {
		return "", err
	}
	var lines []string
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("%s\t%10d\t%v", entry.Name(), entry.Size(), entry.Mode()))
	}
	return strings.Join(lines, "\n"), nil
}

func (cmd Command) statPath(client *sftp.Client) (string, error) {
	if cmd.Path == "" {
		return "", fmt.Errorf("path is required for stat command")
	}
	info, err := client.Stat(cmd.Path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("size=%d mode=%s mod=%s dir=%t", info.Size(), info.Mode(), info.ModTime().UTC().Format(time.RFC3339), info.IsDir()), nil
}

func (s *Server) testCommands(globalResults chan<- serverTestResult, wg *sync.WaitGroup) {
	defer wg.Done()

	serverResult := serverTestResult{server: s.displayName()}

	client, sshClient, duration, err := s.newClient()
	connectionResult := SFTPResult{
		Name:     "connection",
		Mode:     CommandModeConnect,
		Duration: duration,
	}

	if err != nil {
		connectionResult.Err = err
		serverResult.results = append(serverResult.results, connectionResult)
		globalResults <- serverResult
		return
	}

	connectionResult.Output = "connection established"
	serverResult.results = append(serverResult.results, connectionResult)

	defer client.Close()
	defer sshClient.Close()

	for _, cmd := range s.Commands {
		serverResult.results = append(serverResult.results, cmd.execute(client))
	}

	globalResults <- serverResult
}

const (
	commandLabelWidth = 40
)

func (r SFTPResult) success() bool {
	if r.Err != nil {
		return false
	}
	for _, assert := range r.Asserts {
		for _, ok := range assert {
			if !ok {
				return false
			}
		}
	}
	return true
}

func modeBadge(mode CommandMode) string {
	switch mode {
	case CommandModeConnect:
		return color.CyanString("connect")
	case CommandModeRead:
		return color.BlueString("read")
	case CommandModeList:
		return color.MagentaString("list")
	default:
		return color.YellowString(string(mode))
	}
}

func printServerHeader(server string, success bool) {
	fmt.Println(utility.DividerLine())
	fmt.Printf("%s %s\n", utility.StatusBadge(success), color.CyanString(server))
}

func printCommandSummary(result SFTPResult, success bool) {
	label := result.Name
	if label == "" {
		label = fmt.Sprintf("%s %s", result.Mode, result.Path)
	}
	fmt.Printf("%s%s %-*s %s %s\n", utility.Indent, utility.StatusBadge(success), commandLabelWidth, label, modeBadge(result.Mode), utility.DurationBadge(result.Duration))
	if result.Path != "" {
		fmt.Printf("%s%sPath: %s\n", utility.Indent, utility.Indent, color.BlueString(result.Path))
	}
}

func printCommandDetails(result SFTPResult) {
	if result.Err != nil {
		fmt.Printf("%s%sError: %s\n", utility.Indent, utility.Indent, color.RedString(result.Err.Error()))
		return
	}
	if strings.TrimSpace(result.Output) != "" {
		fmt.Printf("%s%sOutput: %s\n", utility.Indent, utility.Indent, color.YellowString(utility.FormatSnippet(result.Output, "no output", utility.DefaultSnippetLimit)))
	}
}

func (services *SFTPServices) RunTesting() {
	if len(services.Servers) == 0 {
		fmt.Println("no SFTP servers configured")
		return
	}

	var wg sync.WaitGroup
	serverChan := make(chan serverTestResult, len(services.Servers))

	for i := range services.Servers {
		wg.Add(1)
		go services.Servers[i].testCommands(serverChan, &wg)
	}

	go func() {
		wg.Wait()
		close(serverChan)
	}()

	for serverResult := range serverChan {
		serverSuccess := true
		for _, result := range serverResult.results {
			if !result.success() {
				serverSuccess = false
				break
			}
		}

		printServerHeader(serverResult.server, serverSuccess)

		for _, result := range serverResult.results {
			success := result.success()
			printCommandSummary(result, success)
			if !success {
				printCommandDetails(result)
			}
			for _, assert := range result.Asserts {
				for description, ok := range assert {
					fmt.Printf("%s%s %s\n", utility.DoubleIndent, utility.StatusBadge(ok), description)
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
