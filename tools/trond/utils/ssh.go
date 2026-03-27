package utils

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// validPathPattern allows only safe characters in remote paths:
// alphanumeric, slash, dot, hyphen, underscore
var validPathPattern = regexp.MustCompile(`^[a-zA-Z0-9/._-]+$`)

// validateRemotePath checks that a path contains no shell-injectable characters.
// This is the primary defense against command injection via user-controlled paths.
func validateRemotePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !validPathPattern.MatchString(path) {
		return fmt.Errorf("invalid characters in remote path %q: only alphanumeric, '/', '.', '-', '_' are allowed", path)
	}
	// Block path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal ('..') is not allowed in remote path: %s", path)
	}
	return nil
}

// validateSCPFileName validates a filename for SCP transfer.
// SCP protocol is sensitive to special characters that could break the control line format.
// The control line format is: C<mode> <size> <filename>\n
// Filenames with newlines, carriage returns, or other control characters can break this format.
func validateSCPFileName(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return fmt.Errorf("filename cannot contain path separators: %s", fileName)
	}

	// Block path traversal attempts
	if strings.Contains(fileName, "..") {
		return fmt.Errorf("filename cannot contain path traversal sequences: %s", fileName)
	}

	// Check for control characters that could break SCP protocol
	// SCP control line ends with \n, so any \n or \r in filename breaks the protocol
	for _, ch := range fileName {
		// Block control characters (0x00-0x1F) and DEL (0x7F)
		if ch < 0x20 || ch == 0x7F {
			return fmt.Errorf("filename contains invalid control character (0x%02X): %s", ch, fileName)
		}
		// Block characters that have special meaning in SCP protocol or shell
		// While we use proper escaping, defense-in-depth means rejecting these outright
		if strings.ContainsRune("*?[]{}()<>|;&$`\\\"'", ch) {
			return fmt.Errorf("filename contains shell-special character %q: %s", ch, fileName)
		}
	}

	// Ensure filename is not too long (SCP protocol limitation)
	if len(fileName) > 255 {
		return fmt.Errorf("filename exceeds maximum length (255 bytes): %s", fileName)
	}

	return nil
}

// shellQuote wraps a string in single quotes with proper escaping.
// This is the secondary defense — even if validation is bypassed,
// the shell will treat the entire string as a literal argument.
func shellQuote(s string) string {
	// Replace each single quote with: end quote, escaped single quote, start quote
	// e.g., "it's" -> "'it'\''s'"
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// Check if a port is open on the given IP
func CheckPort(ip string, port int) bool {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// getHostKeyCallback returns the appropriate HostKeyCallback based on environment
// For production use, set TROND_STRICT_HOST_KEY_CHECK=true environment variable
func getHostKeyCallback() (ssh.HostKeyCallback, error) {
	// Check if strict host key checking is enabled via environment variable
	strictCheck := os.Getenv("TROND_STRICT_HOST_KEY_CHECK")

	if strictCheck == "true" || strictCheck == "1" {
		// Production mode: Use known_hosts file for verification
		knownHostsPath := os.Getenv("TROND_KNOWN_HOSTS_FILE")
		if knownHostsPath == "" {
			// Default to user's known_hosts file
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %v", err)
			}
			knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
		}

		// Check if known_hosts file exists
		if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("known_hosts file not found at %s. Please create it or disable strict host key checking for testing", knownHostsPath)
		}

		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load known_hosts file: %v", err)
		}

		fmt.Println("    ✓ Using strict host key verification (production mode)")
		return callback, nil
	}

	// Development/Testing mode: Accept any host key with warning
	fmt.Println("    ⚠️  WARNING: Host key verification is DISABLED (testing mode)")
	fmt.Println("    ⚠️  For production use, set TROND_STRICT_HOST_KEY_CHECK=true")
	return ssh.InsecureIgnoreHostKey(), nil
}

func SSHConnect(ip string, port int, user, password, keyPath string) (*ssh.Client, error) {
	// Get appropriate host key callback
	hostKeyCallback, err := getHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to configure host key verification: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second,
	}

	// Determine authentication method
	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		// Key-based authentication
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		// Password-based authentication
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		// Try SSH agent
		conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			defer conn.Close()
			agentClient := agent.NewClient(conn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no valid authentication method available")
	}

	sshConfig.Auth = authMethods

	// Connect to the remote server
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), sshConfig)
	return client, err
}

// CheckSSH tests the SSH connection using the provided authentication method
func CheckSSH(ip string, port int, user, password, keyPath string) error {
	// Connect to the remote server
	client, err := SSHConnect(ip, port, user, password, keyPath)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server at %s:%d: %v", ip, port, err)
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Run a test command
	output, err := session.CombinedOutput("echo ok")
	if err != nil {
		return fmt.Errorf("failed to run test command: %v", err)
	}

	if string(output) != "ok\n" {
		return fmt.Errorf("unexpected output from test command: %s", output)
	}

	return nil
}

// SCPFile transfers a local file to a remote destination via SCP with a progress bar
func SCPFile(ip string, port int, user, password, keyPath, localPath, remotePath string) error {
	// Connect to the remote server
	client, err := SSHConnect(ip, port, user, password, keyPath)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer client.Close()

	// Create SCP session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Get file stats
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %v", err)
	}
	fileSize := fileInfo.Size()

	// Create stdin pipe BEFORE starting the session
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	defer stdin.Close()

	// Validate remote path before constructing shell command
	remotePathDir := filepath.Dir(remotePath)
	if err := validateRemotePath(remotePathDir); err != nil {
		return fmt.Errorf("unsafe remote path for SCP: %v", err)
	}

	// Start SCP command on the remote server (quoted to prevent injection)
	scpCmd := fmt.Sprintf("scp -t %s", shellQuote(remotePathDir))
	if err := session.Start(scpCmd); err != nil {
		return fmt.Errorf("failed to start SCP command: %v", err)
	}

	// Validate and sanitize the filename
	fileName := filepath.Base(remotePath)
	if err := validateSCPFileName(fileName); err != nil {
		return fmt.Errorf("unsafe filename for SCP: %v", err)
	}

	// Send SCP file metadata with properly escaped filename
	_, err = fmt.Fprintf(stdin, "C0644 %d %s\n", fileSize, fileName)
	if err != nil {
		return fmt.Errorf("failed to send file metadata: %v", err)
	}

	// Initialize progress bar
	bar := progressbar.DefaultBytes(fileSize, "    Uploading...")

	// Create a proxy writer to track progress
	progressWriter := io.MultiWriter(stdin, bar)

	// Copy file data with progress tracking
	if _, err := io.Copy(progressWriter, localFile); err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	// Signal transfer completion
	if _, err := stdin.Write([]byte("\x00")); err != nil {
		return fmt.Errorf("failed to signal transfer completion: %v", err)
	}

	// Close stdin so SCP knows we're done
	stdin.Close()

	// Wait for SCP command to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("SCP session error: %v", err)
	}

	fmt.Println("    File transfer completed successfully!")
	return nil
}

// SSHMkdirIfNotExist checks if a directory exists on a remote machine via SSH and creates it if it doesn't exist.
func SSHMkdirIfNotExist(ip string, port int, user, password, keyPath, remoteDir string) error {
	// Connect to the remote server
	client, err := SSHConnect(ip, port, user, password, keyPath)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer client.Close()

	// Create SSH session for checking directory existence
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Validate remote directory path before constructing shell commands
	if err := validateRemotePath(remoteDir); err != nil {
		return fmt.Errorf("unsafe remote directory path: %v", err)
	}

	// Check if the directory exists (quoted to prevent injection)
	checkCmd := fmt.Sprintf("[ -d %s ] && echo \"exists\" || echo \"not exists\"", shellQuote(remoteDir))
	output, err := session.CombinedOutput(checkCmd)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %v", err)
	}

	// If directory exists, return early
	if string(output) == "exists\n" {
		fmt.Printf("    Remote directory %s already exists, skipping creation.\n", remoteDir)
		return nil
	}

	// Create a new session for mkdir
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session for mkdir: %v", err)
	}
	defer session.Close()

	// Execute mkdir command remotely (path already validated above)
	cmd := fmt.Sprintf("mkdir -p %s", shellQuote(remoteDir))
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	fmt.Printf("    Remote directory %s created successfully!\n", remoteDir)
	return nil
}

// RunRemoteCompose runs `docker-compose up -d` on a remote machine via SSH
func RunRemoteCompose(ip string, port int, user, password, keyPath, composePath string, down bool) error {
	// Connect to the remote server
	client, err := SSHConnect(ip, port, user, password, keyPath)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Validate compose file path before constructing shell command
	if err := validateRemotePath(composePath); err != nil {
		return fmt.Errorf("unsafe docker-compose path: %v", err)
	}

	// Construct the docker-compose command (quoted to prevent injection)
	cmd := fmt.Sprintf("docker-compose -f %s up -d", shellQuote(composePath))
	if down {
		cmd = fmt.Sprintf("docker-compose -f %s down", shellQuote(composePath))
	}

	// Run the command remotely
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to start docker-compose: %v\nOutput: %s", err, output)
	}

	fmt.Println("Docker Compose executed successfully on remote machine!")
	fmt.Println("Output:", string(output))
	return nil
}
