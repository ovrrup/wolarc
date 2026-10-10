package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	Name    string
	Running bool
	PID     int
}

type RAMInfo struct {
	TotalMB     int
	UsedMB      int
	AvailableMB int
	UsedPercent int
}

func GetRAMInfo() RAMInfo {
	var info RAMInfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return info
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.Atoi(fields[1])

		if key == "MemTotal" {
			info.TotalMB = val / 1024
		} else if key == "MemAvailable" {
			info.AvailableMB = val / 1024
		}
	}

	if info.TotalMB > 0 {
		info.UsedMB = info.TotalMB - info.AvailableMB
		if info.UsedMB < 0 {
			info.UsedMB = 0
		}
		info.UsedPercent = int(float64(info.UsedMB) / float64(info.TotalMB) * 100.0)
	}

	return info
}

func CheckProcess(pattern string) (bool, int) {
	out, err := exec.Command("pgrep", "-f", pattern).Output()
	if err != nil {
		return false, 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		if pid, err := strconv.Atoi(lines[0]); err == nil {
			return true, pid
		}
	}
	return false, 0
}

type ServiceStatus struct {
	ChatServer   ProcessInfo
	EmbedServer  ProcessInfo
	AgentGateway ProcessInfo
	Tunnel       ProcessInfo
	Indexer      ProcessInfo
}

func CheckAllServices(chatPort, embedPort string) ServiceStatus {
	var s ServiceStatus

	chatOk, chatPid := CheckProcess(fmt.Sprintf("llama-server.*%s", chatPort))
	s.ChatServer = ProcessInfo{Name: "Chat LLM", Running: chatOk, PID: chatPid}

	embedOk, embedPid := CheckProcess(fmt.Sprintf("llama-server.*%s", embedPort))
	s.EmbedServer = ProcessInfo{Name: "Embedding Engine", Running: embedOk, PID: embedPid}

	agentOk, agentPid := CheckProcess("agent.py")
	s.AgentGateway = ProcessInfo{Name: "Agent Gateway", Running: agentOk, PID: agentPid}

	tunnelOk, tunnelPid := CheckProcess("ssh.*(localhost.run|serveo.net|pinggy.io)")
	s.Tunnel = ProcessInfo{Name: "Public Tunnel", Running: tunnelOk, PID: tunnelPid}

	idxOk, idxPid := CheckProcess("indexer.py")
	s.Indexer = ProcessInfo{Name: "Vector Indexer", Running: idxOk, PID: idxPid}

	return s
}

func StartServices() error {
	cmd := exec.Command("wola.r")
	setProcessGroup(cmd)
	return cmd.Start()
}

func StopServices() error {
	cmd := exec.Command("wola.k")
	return cmd.Run()
}

func RestartServices() error {
	cmd := exec.Command("wola.rs")
	setProcessGroup(cmd)
	return cmd.Start()
}
