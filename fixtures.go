package blog

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

func DuplicateUser(r Repository, u User, username string) *User {
	u1 := u
	u1.ID = nextID()
	u1.Username = username

	_ = r.Store(&u1)

	return &u1
}

func GetContainerIP(containerID string) (string, error) {
	iOut, err := exec.Command("docker", "inspect", containerID).Output()
	if err != nil {
		return "", err
	}

	var con []struct {
		NetworkSettings struct {
			IPAddress string
		}
	}
	_ = json.NewDecoder(bytes.NewReader(iOut)).Decode(&con)
	ip := con[0].NetworkSettings.IPAddress
	return ip, nil
}

func RunDockerContainer(containerName string) (string, error) {
	out, err := exec.Command("docker", "container", "run", "--detach", "--rm", containerName).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
