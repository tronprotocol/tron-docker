package utils

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

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

// CheckSSH tests the SSH connection using the provided authentication method
func CheckSSH(ip string, port int, user, password, keyPath string) error {
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Replace with known hosts verification in production
		Timeout:         5 * time.Second,
	}

	// Determine authentication method
	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		// Key-based authentication
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
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
		return fmt.Errorf("no valid authentication method available")
	}

	sshConfig.Auth = authMethods

	// Connect to the remote server
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), sshConfig)
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
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Replace in production
		Timeout:         10 * time.Second,
	}

	var authMethods []ssh.AuthMethod

	// Key-based authentication
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Password authentication
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// Attempt SSH agent authentication
	if len(authMethods) == 0 {
		conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			defer conn.Close()
			agentClient := agent.NewClient(conn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no valid authentication method available")
	}

	sshConfig.Auth = authMethods

	// Connect to the remote server
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), sshConfig)
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

	// Start SCP command on the remote server
	scpCmd := fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))
	if err := session.Start(scpCmd); err != nil {
		return fmt.Errorf("failed to start SCP command: %v", err)
	}

	// Send SCP file metadata
	_, err = fmt.Fprintf(stdin, "C0644 %d %s\n", fileSize, filepath.Base(remotePath))
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
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Replace in production
		Timeout:         10 * time.Second,
	}

	var authMethods []ssh.AuthMethod

	// Key-based authentication
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Password authentication
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// Attempt SSH agent authentication
	if len(authMethods) == 0 {
		conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			defer conn.Close()
			agentClient := agent.NewClient(conn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no valid authentication method available")
	}

	sshConfig.Auth = authMethods

	// Connect to the remote server
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), sshConfig)
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

	// Check if the directory exists
	checkCmd := fmt.Sprintf("[ -d \"%s\" ] && echo \"exists\" || echo \"not exists\"", remoteDir)
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

	// Execute mkdir command remotely
	cmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	fmt.Printf("    Remote directory %s created successfully!\n", remoteDir)
	return nil
}

// RunRemoteCompose runs `docker-compose up -d` on a remote machine via SSH
func RunRemoteCompose(ip string, port int, user, password, keyPath, composePath string, down bool) error {
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Replace with known hosts in production
		Timeout:         15 * time.Second,
	}

	var authMethods []ssh.AuthMethod

	// Key-based authentication
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Password authentication
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// Attempt SSH agent authentication
	if len(authMethods) == 0 {
		conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err == nil {
			defer conn.Close()
			agentClient := agent.NewClient(conn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no valid authentication method available")
	}

	sshConfig.Auth = authMethods

	// Connect to the remote server
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), sshConfig)
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

	// Construct the docker-compose command
	cmd := fmt.Sprintf("docker-compose -f %s up -d", composePath)
	if down {
		cmd = fmt.Sprintf("docker-compose -f %s down", composePath)
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
